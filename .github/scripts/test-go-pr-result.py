#!/usr/bin/env python3

import copy
import importlib.util
import os
from pathlib import Path
import subprocess
import tempfile
import unittest
from unittest.mock import patch

SPEC = importlib.util.spec_from_file_location("go_pr_result", Path(__file__).with_name("go-pr-result.py"))
reuse = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(reuse)


class GoResultTest(unittest.TestCase):
    def test_landed_tree_and_complete_pr_evidence(self):
        with tempfile.TemporaryDirectory() as temp:
            previous = Path.cwd()
            self.addCleanup(os.chdir, previous)
            os.chdir(temp)
            env = {k: v for k, v in os.environ.items() if not k.startswith("GIT_")}
            env.update(GIT_CONFIG_GLOBAL=os.devnull, GIT_CONFIG_NOSYSTEM="1")
            with patch.dict(os.environ, env, clear=True):
                self.exercise_repository()

    def exercise_repository(self):
        def git(*args):
            return subprocess.check_output(["git", "-c", "user.name=CI", "-c", "user.email=ci@example.invalid",
                                            *args], text=True).strip()

        git("init", "-q", "-b", "master")
        git("commit", "-qm", "base", "--allow-empty")
        base = git("rev-parse", "HEAD")
        git("switch", "-qc", "feature")
        Path("main.go").write_text("package main\n")
        git("add", "--", "main.go")
        git("commit", "-qm", "Go input")
        head = git("rev-parse", "HEAD")
        git("switch", "-q", "master")
        git("merge", "-q", "--no-ff", "-m", "PR test merge", "feature")
        tested = git("rev-parse", "HEAD")
        tree = git("rev-parse", "HEAD^{tree}")
        # Make a distinct squash commit with the same files through Git's commit-tree.
        landed = git("commit-tree", tree, "-p", base, "-m", "landed")
        finished, merged = "2026-01-01T00:04:00Z", "2026-01-01T00:05:00Z"
        expected = reuse.expected_steps()
        names = {**expected, "Generate Build Matrix": {f"Checkout PR #8 merge {tested}"}, "Check Modified Files": set()}
        jobs = [{"name": n, "conclusion": "success", "steps": [
            {"name": step, "conclusion": "success"} for step in steps]} for n, steps in names.items()]
        pr = {"number": 8, "state": "closed", "merged_at": merged, "merge_commit_sha": landed,
              "base": {"ref": "master", "repo": {"full_name": "example/agent"}},
              "head": {"sha": head, "ref": "feature", "repo": {"id": 123}}}
        run = {"id": 77, "path": reuse.WORKFLOW, "event": "pull_request", "head_sha": head, "head_branch": "feature",
               "head_repository": {"id": 123}, "status": "completed", "conclusion": "success",
               "run_attempt": 1, "updated_at": finished}
        merge = {"parents": [{"sha": base}, {"sha": head}], "tree": {"sha": tree}}

        def find(prs=None, run_changes=None, job_rows=None, tested_commit=None):
            rows = jobs if job_rows is None else job_rows
            def api(repo, path):
                self.assertEqual(repo, "example/agent")
                if path.startswith("commits/"):
                    return [pr] if prs is None else prs
                if path.startswith("actions/workflows/go-tests.yml/runs?"):
                    return {"total_count": 1, "workflow_runs": [run | (run_changes or {})]}
                if path == "actions/runs/77/attempts/1/jobs?per_page=100":
                    return {"total_count": len(rows), "jobs": rows}
                self.assertEqual(path, f"git/commits/{tested}")
                return merge if tested_commit is None else tested_commit
            with patch.object(reuse, "github_get", side_effect=api):
                return reuse.reusable_run("example/agent", "master", base, landed)

        self.assertEqual(find(), 77)
        self.assertIsNone(find(prs=[]))
        self.assertIsNone(find(prs=[pr, pr]))
        for changes in ({"status": "in_progress"}, {"conclusion": "failure"}, {"run_attempt": 2},
                        {"updated_at": "2026-01-01T00:06:00Z"}, {"head_repository": {"id": 456}}):
            with self.subTest(changes=changes):
                self.assertIsNone(find(run_changes=changes))
        self.assertIsNone(find(job_rows=jobs[1:]))
        skipped = copy.deepcopy(jobs)
        skipped[0]["steps"][0]["conclusion"] = "skipped"
        self.assertIsNone(find(job_rows=skipped))
        self.assertIsNone(find(tested_commit=merge | {"tree": {"sha": "0" * 40}}))
        self.assertIsNone(find(tested_commit=merge | {"parents": [{"sha": "1" * 40}, {"sha": head}]}))

        git("switch", "-q", "feature")
        policy = Path(".github/workflows/go-tests.yml")
        policy.parent.mkdir(parents=True)
        policy.write_text("changed CI policy\n")
        git("add", "--", str(policy))
        git("commit", "-qm", "CI policy")
        changed = git("rev-parse", "HEAD")
        with patch.object(reuse, "github_get", side_effect=AssertionError("CI changes must validate locally")):
            self.assertIsNone(reuse.reusable_run("example/agent", "master", head, changed))


if __name__ == "__main__":
    unittest.main()
