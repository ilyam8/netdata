#!/usr/bin/env python3
"""Find a completed Go PR validation for the exact tree that landed."""

import json
import os
import re
import subprocess
import sys
import tempfile
from pathlib import Path
from urllib.parse import urlencode

SCRIPTS = Path(__file__).resolve().parent
WORKFLOW = ".github/workflows/go-tests.yml"
SHA = re.compile(r"[a-f0-9]{40}")


def command(*args, **kwargs):
    return subprocess.check_output(args, text=True, stderr=subprocess.PIPE, timeout=12, **kwargs).strip()


def git(*args):
    return command("git", *args)


def expected_steps():
    # Use the same matrix generator as the workflow, including new modules and platforms.
    with tempfile.TemporaryDirectory() as tmp:
        output = Path(tmp) / "matrix"
        command(sys.executable, str(SCRIPTS / "get-go-version.py"), env=dict(os.environ, GITHUB_OUTPUT=str(output)))
        matrices = dict(line.split("=", 1) for line in output.read_text().splitlines())
    steps = {"Go ebpf libbpf-tagged tests": {"Go test (libbpf tag)"}}
    for entry in json.loads(matrices["matrix"]):
        steps[f"Go toolchain tests ({entry['module']})"] = {"Compile", "Go fmt", "Go vet", "Go test"}
    for entry in json.loads(matrices["build_matrix"]):
        steps[f"Go build tests ({entry['module']}, {entry['platform']})"] = {"Compile"}
    return steps


def checked_merge(jobs, number, expected):
    if any(job["conclusion"] != "success" for job in jobs):
        return None
    steps = {job["name"]: {s["name"] for s in job["steps"] if s["conclusion"] == "success"} for job in jobs}
    if len(steps) != len(jobs) or set(steps) != set(expected) | {"Check Modified Files", "Generate Build Matrix"}:
        return None
    if not all(required <= steps[name] for name, required in expected.items()):
        return None
    marker = re.compile(rf"Checkout PR #{number} merge ([a-f0-9]{{40}})")
    matches = [m[1] for name in steps["Generate Build Matrix"] if (m := marker.fullmatch(name))]
    return matches[0] if len(matches) == 1 else None


def github_get(repo, path):
    return json.loads(command("gh", "api", f"repos/{repo}/{path}"))


def reusable_run(repo, branch, before, landed):
    if not all(SHA.fullmatch(ref) for ref in (before, landed)):
        return None
    parents = git("show", "-s", "--format=%P", landed).split()
    if not parents or parents[0] != before or len(parents) > 2:
        return None
    # CI policy changes validate on master. Never consume PR artifacts or execute PR-provided proof.
    if git("diff", "--name-only", before, landed, "--", ".github"):
        return None
    associated = github_get(repo, f"commits/{landed}/pulls?per_page=100")
    prs = [pr for pr in associated if pr["state"] == "closed" and pr["merged_at"]
           and pr["merge_commit_sha"] == landed and pr["base"]["ref"] == branch
           and pr["base"]["repo"]["full_name"] == repo]
    if len(prs) != 1 or len(associated) >= 100:
        return None
    pr = prs[0]
    head = pr["head"]["sha"]
    if not SHA.fullmatch(head) or (len(parents) == 2 and parents[1] != head):
        return None
    query = urlencode({"event": "pull_request", "head_sha": head, "per_page": 10})
    result = github_get(repo, f"actions/workflows/go-tests.yml/runs?{query}")
    if result["total_count"] > len(result["workflow_runs"]):
        return None
    runs = [r for r in result["workflow_runs"] if r["event"] == "pull_request" and r["path"].partition("@")[0] == WORKFLOW
            and r["head_sha"] == head and r["head_branch"] == pr["head"]["ref"]
            and r["head_repository"]["id"] == pr["head"]["repo"]["id"]]
    if len(runs) != 1:
        return None
    run = runs[0]
    if (run["status"] != "completed" or run["conclusion"] != "success" or run["run_attempt"] != 1
            or run["updated_at"] > pr["merged_at"]):
        return None
    result = github_get(repo, f"actions/runs/{run['id']}/attempts/1/jobs?per_page=100")
    if len(result["jobs"]) != result["total_count"]:
        return None
    tested = checked_merge(result["jobs"], pr["number"], expected_steps())
    if not tested:
        return None
    merge = github_get(repo, f"git/commits/{tested}")
    if ([p["sha"] for p in merge["parents"]] != [before, head]
            or merge["tree"]["sha"] != git("rev-parse", f"{landed}^{{tree}}")):
        return None
    # A PR must not be able to supply its own evidence-producing workflow.
    if subprocess.run(["git", "cat-file", "-e", f"{head}^{{commit}}"], stderr=subprocess.DEVNULL).returncode:
        git("fetch", "--no-tags", "--no-recurse-submodules", "--depth=1", "origin", head)
    if git("diff", "--name-only", before, head, "--", ".github"):
        return None
    return int(run["id"])


def main():
    try:
        run = reusable_run(os.environ["GITHUB_REPOSITORY"], os.environ["GITHUB_REF_NAME"],
                           os.environ["PUSH_BEFORE"], os.environ["GITHUB_SHA"])
    except (OSError, subprocess.SubprocessError, ValueError, KeyError, TypeError) as error:
        # Missing Git history, API permissions, rate limits and incomplete records all run the matrix.
        print(f"Go PR validation could not be verified ({type(error).__name__}).")
        return
    if not run:
        print("No equivalent completed Go PR validation was verified. Running the matrix.")
        return
    url = f"{os.environ['GITHUB_SERVER_URL']}/{os.environ['GITHUB_REPOSITORY']}/actions/runs/{run}"
    message = f"Reused the complete Go matrix on this exact Git tree from [PR run {run}]({url})."
    print(message)
    with open(os.environ["GITHUB_STEP_SUMMARY"], "a") as summary:
        print(message, file=summary)
    with open(os.environ["GITHUB_OUTPUT"], "a") as output:
        print(f"run_id={run}", file=output)


if __name__ == "__main__":
    main()
