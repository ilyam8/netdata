# scripts.d.plugin (preview)

`scripts.d.plugin` runs Nagios-style check scripts inside Netdata without changing
plugin output format. The active collector is `nagios` (single collector surface),
executed through one shared internal async execution service.

> **Status:** preview. Core execution, perfdata routing, and internal execution telemetry are
> implemented; config/docs may still evolve.

## Configuration

- Plugin-level toggles: `/etc/netdata/scripts.d.conf`
- Collector jobs: `/etc/netdata/scripts.d/nagios.conf`

Each job is a Nagios check definition.

Example:

```yaml
jobs:
  - name: ping_localhost
    plugin: "/usr/lib/nagios/plugins/check_ping"
    args: ["-H", "127.0.0.1", "-w", "100.0,20%", "-c", "200.0,40%"]
    timeout: 60s
    check_interval: 1m
    retry_interval: 30s
    max_check_attempts: 3
```

### Time Periods

`check_period` is supported. Custom periods are defined with `time_periods` in the
same collector config file.

```yaml
time_periods:
  - name: 24x7
    alias: Always on
    rules:
      - type: weekly
        days: [sunday, monday, tuesday, wednesday, thursday, friday, saturday]
        ranges: ["00:00-24:00"]

jobs:
  - name: local_plugins
    plugin: "/usr/lib/nagios/plugins/check_dummy"
    args: ["0", "ok"]
    check_period: 24x7
```

## Execution Model

- Jobs share one internal execution service.
- Job execution cadence is independent (`check_interval` / `retry_interval`).
- Worker/queue sizing is internal in the current preview design; there is no public
  scheduler tuning surface.
- Public collector metrics are emitted every `update_every` (default: `10s`).
- Internal execution runtime telemetry is producer-driven on the runtime service cadence.

Defaults:

- `check_interval`: `5m`
- `retry_interval`: `1m`
- `timeout`: `60s`
- `max_check_attempts`: `3`

## Metrics and Charts

Static template charts:

- `nagios.job_state`
- `nagios.job_attempts`

Perfdata is routed plugin-side and materialized via autogen (bounded lifecycle):

- Unit classes: `time`, `bytes`, `bits`, `percent`, `counter`, `generic`
- Metric identity: sanitized perfdata key (from Nagios perfdata label)
- Unit drift policy: drop sample when a key changes unit class
- Collision policy: deterministic keep-first, drop conflicting label

## Runtime Telemetry

Execution runtime telemetry is emitted through the internal `runtimecomp` seam. It is
not exposed as public Nagios scheduler charts.

## Logging

Execution events are emitted through the runtime log emitter (plugin logger path).
Collector-level OTLP logging configuration is not exposed in `nagios` collector
config yet.

## Tests

```bash
cd src/go
go test ./plugin/scripts.d/collector/nagios -count=1
go test ./plugin/scripts.d/collector/nagios/internal/execution -count=1
go test ./plugin/scripts.d/collector/nagios/internal/runtime -count=1
```

## Build

```bash
cmake -DENABLE_PLUGIN_SCRIPTS=On ..
cmake --build . --target scripts-plugin
```

Binary path:

- `usr/libexec/netdata/plugins.d/scripts.d.plugin`

Stock config path:

- `usr/lib/netdata/conf.d/scripts.d/`
