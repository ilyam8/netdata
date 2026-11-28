# Rust Crates Area Instructions

The repo-root `AGENTS.md` applies in full; where the two conflict, this more specific file wins.

## Writing User-Facing Files

- When you write or edit user-facing files, write about 80% of the way to ASD-STE100. User-facing files are:
  - under this directory: `README.md`, `metadata.yaml`, comments in shipped config files;
  - the related documentation pages under `docs/`, such as `docs/opentelemetry/` and `docs/npm/network-flows/`.
- This does not apply to code, code comments, developer specs (for example `FORMAT.md`), or your chat replies.
- Do not edit generated `integrations/*.md` files; edit the crate's `metadata.yaml` instead.

## CI maintenance

When adding a crate or changing which binaries depend on it, check the Cargo dependency graph and the shipped targets
in `packaging/cmake/Modules/NetdataRust.cmake`.

- **OTel packages.** New crates get the ordinary PR package matrix. Add a crate to `OTEL_CRATES` in
  `.github/scripts/pr-package-scope.py` only if it is specific to the shipped OTel plugin and should use the
  representative package targets. Remove it from the set if it becomes shared with another shipped component.
- **Developer tools.** If `otel-streams` becomes part of a shipped target, remove its package skip from the selector.
- **NetFlow tests.** The general workspace test command excludes `netflow-plugin`. If NetFlow gains a direct or
  transitive local dependency, check its path against the NetFlow filter in `.github/workflows/rust-tests.yml`.
  Add any uncovered directory so an edit confined to that dependency runs the `Test NetFlow` step.

For package selector changes, cover a source-only edit to the affected crate in the selector tests. Run
`python3 .github/scripts/test-ci-workflow-file-selection.py` from the repository root.
