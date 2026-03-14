# Secrets Management

## Overview

Netdata collectors can resolve secrets at job startup instead of storing plain-text credentials in collector configs.

Supported direct secret references:

- `${env:VAR_NAME}`
- `${file:/absolute/path/to/secret}`
- `${cmd:/absolute/path/to/command args}`

Remote secret backends are configured as **secretstores** and referenced from collector configs with:

```text
${store:<kind>:<name>:<operand>}
```

Supported secretstore kinds:

- `vault`
- `aws-sm`
- `azure-kv`
- `gcp-sm`

:::tip

**TL;DR** — Use `${env:...}`, `${file:...}`, or `${cmd:...}` for local/direct secrets. Use `${store:<kind>:<name>:<operand>}` for remote backends configured as secretstores.

:::

## Direct Secret References

### Environment Variables

Use `${env:VARIABLE_NAME}` to resolve an environment variable:

```yaml
jobs:
  - name: local
    dsn: "${env:MYSQL_USER}:${env:MYSQL_PASSWORD}@tcp(127.0.0.1:3306)/"
```

### Files

Use `${file:/absolute/path}` to read a secret from a file. Leading and trailing whitespace is trimmed.

```yaml
jobs:
  - name: myapp
    password: "${file:/run/secrets/myapp_password}"
```

### Commands

Use `${cmd:/absolute/path/to/command args}` to execute a command and use its stdout as the secret value.

```yaml
jobs:
  - name: prod
    password: "${cmd:/usr/bin/op read op://vault/netdata/mysql/password}"
```

:::important

Command paths must be absolute. Commands have a 10-second timeout.

:::

## Secretstores

Remote secret backends are not referenced directly. They are configured as secretstore objects and then used from collector configs through `${store:...}` references.

Reference format:

```text
${store:<kind>:<name>:<operand>}
```

Examples:

```yaml
jobs:
  - name: mysql_prod
    password: "${store:vault:vault_prod:secret/data/netdata/mysql#password}"

  - name: redis_prod
    password: "${store:aws-sm:aws_prod:netdata/redis#password}"

  - name: api_prod
    token: "${store:azure-kv:azure_prod:my-vault/api-token}"

  - name: app_prod
    password: "${store:gcp-sm:gcp_prod:my-project/mysql-password}"
```

Meaning of each part:

- `kind`: secretstore provider kind (`vault`, `aws-sm`, `azure-kv`, `gcp-sm`)
- `name`: configured secretstore object name
- `operand`: provider-specific secret selector

### Provider Operands

- `vault`: `path#key`
  - Example: `secret/data/netdata/mysql#password`
- `aws-sm`: `secret-name` or `secret-name#key`
  - Example: `netdata/mysql#password`
- `azure-kv`: `vault-name/secret-name`
  - Example: `my-keyvault/mysql-password`
- `gcp-sm`: `project/secret` or `project/secret/version`
  - Example: `my-project/mysql-password`

## How Secretstores Are Used

1. Configure a secretstore object for the provider/backend you want to use.
2. Reference that secretstore from collector configs with `${store:<kind>:<name>:<operand>}`.
3. When the collector job starts or restarts, Netdata resolves the secret through the active secretstore runtime.

If resolution fails, the collector job fails to start or restart with an error.

## Current Provisioning Flow

Today, secretstores are created and managed through the Dynamic Configuration Manager control plane.

- Create a secretstore object of the required kind (`vault`, `aws-sm`, `azure-kv`, `gcp-sm`).
- Give it a `name` that collectors will reference.
- Provide the provider-specific payload defined by that kind's schema.
- Then use `${store:<kind>:<name>:<operand>}` from collector configs.

The `kind` and `name` are external secretstore object metadata. They are not part of the provider payload itself.

File-based secretstore config ingest is not documented here because that flow is not the current shipped provisioning path.

## Dynamic Behavior

- Secret resolution happens when a collector job starts or restarts.
- Secretstores are part of the control plane.
- Updating an active secretstore restarts dependent jobs so they pick up the new secretstore runtime.
- `test` validates submitted or stored secretstore config without mutating runtime state.

## What Is Not Supported

These direct remote-provider syntaxes are **not** supported:

- `${vault:...}`
- `${aws-sm:...}`
- `${azure-kv:...}`
- `${gcp-sm:...}`

Use secretstores instead:

```text
${store:<kind>:<name>:<operand>}
```

## Security Notes

- Only string config values are scanned for secret references.
- Resolution is single-pass. A resolved value is not scanned again.
- Internal metadata keys (`__...__`) are not resolved as secrets.
- Avoid plain-text credentials in collector configs when an environment variable, file, command, or secretstore can be used instead.
