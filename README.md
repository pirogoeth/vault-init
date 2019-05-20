# `vault-init`

## Rationale

Previously, I could use Nomad's templating to insert data from Vault into an application.
Since we're now using Docker's Swarm Mode instead, I can't use any fancy templating.
But I also do not want to go back to placing secrets inside of workload definitions.

## Design Decisions / Roadmap

- [X] Will run as an init system inside a container
  - Luckily this will not require a ton of functionality
  - We need to be able to:
    - [X] Spawn processes
    - [X] Reap dead children
    - Perform signal forwarding to children
    - [X] Forward all environment variables to children
      - **EXCLUDING** Vault-init configuration (`INIT_*`, optionally `VAULT_*` when `--no-inherit-token` is unset)
- [X] Get Vault connect token from environment var or from file
  - [X] VAULT_TOKEN_FILE, which would load in to VAULT_TOKEN
  - (this supports `docker secrets` well)
- [X] We can piggyback on Vault's preset client configuration environment variables
  - https://github.com/hashicorp/vault/blob/master/api/client.go#L28
- [X] Connect to Vault using `VAULT_TOKEN`/`VAULT_TOKEN_FILE`
  - [X] Generate a token with policies given by `INIT_ACCESS_POLICIES`
    - [X] Token should have `VAULT_TOKEN` as parent unless `INIT_ORPHAN_TOKEN` is `true`
      - [ ] Token roles?
    - [X] Token should be renewable unless `INIT_DISABLE_RENEW` is `true`
    - [X] Token should be provided to child as `VAULT_TOKEN` unless `INIT_NO_INHERIT_TOKEN` is `true`
    - [X] Token should be revoked on `vault-init` exit
- [X] Use Go's `text/template` library to do templating into environment variables and files in the container
  - [X] Template context loaded in based on comma-separated `INIT_PATHS`
    - Example:
      - `export INIT_PATHS="/secret/services/concourse"`
      - `export INIT_PATHS="/secret/services/sourcegraph,/secret/services/oauth2-proxy/sourcegraph"`
  - [ ] When multiple paths are provided, try to contextually diff the URLs to create nested structure
    - If only one path is provided, it would become the top-level data
    - If more than one path is provided, and the paths share ancestry:
      - Example:
        - `path:"/secret/data/services/concourse"`
        - `path:"/secret/data/services/sourcegraph"`
        - `      ^^^^^^^^^^^^^^^^^^^^^^ shared ancestry`
        - `.Data.concourse.some_value`
        - `.Data.sourcegraph.some_value`
    - If more than one path is provided and the paths do not share ancestry:
      - Example:
        - `path:"/secret/data/services/concourse"`
        - `path:"/kv1/services/haproxy"`
        - `.Data.secret.data.services.concourse.some_value`
        - `.Data.kv1.services.haproxy`
  - Helpers for certain actions(?)
    - Undetermined
- [~] Correctly handle renewable secrets
  - [~] Leased secrets
    - [X] Should be renewed
    - [ ] Should be revoked when `vault-init` exits
  - [~] Auth secrets
    - [X] Should be renewed
    - [ ] Should be revoked when `vault-init` exits