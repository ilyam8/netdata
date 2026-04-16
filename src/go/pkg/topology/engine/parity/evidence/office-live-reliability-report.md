# Sanitized Live Reliability Report (Topology SNMP)

## Scope
- Gate: `T4.5` + `T4.6`
- Seed CIDR: `<redacted-rfc1918-cidr>`
- Community: `<redacted>`
- Function under validation: `topology:snmp` (`topology_view:l2|l3|merged`)

## Capture Method
- Rebuilt and installed the latest code in a local validation environment.
- Captured `l2`, `l3`, and `merged` topology views through the Netdata API v3 function endpoint.
- Authentication material, token storage paths, real seed CIDRs, hostnames, and timestamped artifact paths are intentionally omitted.

## Before/After (Identity-Split)
- Baseline snapshot: `/tmp/topology-live-sample/baseline-l2.json`
- Post-fix snapshot: `/tmp/topology-live-sample/post-fix-l2.json`

| Metric | Before | After |
|---|---:|---:|
| Actors | 89 | 75 |
| Links | 11 | 11 |
| Duplicate actors by normalized MAC | 7 | 0 |
| Duplicate actors by normalized IP | 7 | 0 |
| Unidirectional links | 11 | 11 |
| Bidirectional links | 0 | 0 |

## Stability (12 refreshes)
- Artifact directory: `/tmp/topology-live-stability`
- Hash file: `/tmp/topology-live-stability/hash-counts.txt`
- Identity summary: `/tmp/topology-live-stability/identity-summary.txt`

Results:
- Normalized structure hash: `12/12` identical.
- Per-sample counts stable: `actors=76`, `links=11`, `mac_dups=0`, `ip_dups=0`.

## Core Device Presence (single occurrence in post-fix capture)
- `router-a`: 1
- `switch-a`: 1
- `switch-b`: 1
- `switch-c`: 1
- `host-a`: 1
- `host-b`: 1
- `host-c`: 1

## View Status
- `l2`: `status=200`, topology data present.
- `l3`: `status=200`, `actors=0`, `links=0` (no OSPF/ISIS activity observed in the validation snapshot).
- `merged`: `status=200`, mirrors available L2 + L3 data.

## Conclusion
- Identity-split pathology is removed in live captures (duplicate MAC/IP actors reduced to zero).
- Topology output is deterministic across repeated refreshes.
- Reliability criteria for current protocol visibility are satisfied.

## Post-T5.3 Validation (2026-02-21)
- Artifact directory: `/tmp/topology-live-t53`
- Stability directory: `/tmp/topology-live-stability-t53`

View summary:
- `l2`: `status=200`, `actors=101`, `links=190`, `links_lldp=10`, `links_fdb=180`, `bidirectional=181`, `unidirectional=9`
- `l3`: `status=200`, `actors=0`, `links=0`
- `merged`: `status=200`, `actors=101`, `links=190`

LLDP summary:
- Bidirectional core edge present: `switch-a:8 <-> router-a:ether3`.
- Remaining `9` LLDP edges are unidirectional due missing reciprocal remote rows on peers.

Stability sampling (`12` refreshes):
- Structural hashes: `3` unique (`1 + 1 + 10` distribution).
- Samples `3..12` converged to one stable hash with fixed counts:
  - `actors=101`, `links=188`, `links_lldp=10`, `links_fdb=178`, `bidirectional=179`, `unidirectional=9`.
- Interpretation:
  - first two refreshes after restart reflected expected transient cache convergence;
  - steady-state snapshots are stable after convergence.
