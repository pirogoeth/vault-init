# `vault-init`

## Rationale

Previously, I could use Nomad's templating to insert data from Vault into an application.
Since we're now using Docker's Swarm Mode instead, I can't use any fancy templating.
But I also do not want to go back to placing secrets inside of workload definitions.

## Design Decisions

- Will run as an init system inside a container
  - Luckily this will not require a ton of functionality
  - We need to be able to:
    - Spawn processes
    - Reap dead children
    - Perform signal forwarding to children
    - Forward all environment variables to children
      - **EXCLUDING** Vinit configuration
- Get Vault connect token from environment var or from file
  - VAULT_TOKEN_FILE, which would load in to VAULT_TOKEN
  - (this supports `docker secrets` well)
- We can piggyback on Vault's preset client configuration environment variables
  - https://github.com/hashicorp/vault/blob/master/api/client.go#L28
- Connect to Vault using `VAULT_TOKEN`/`VAULT_TOKEN_FILE`
  - Generate a token with policies given by `VINIT_ACCESS_POLICIES`
  - Token should have `VAULT_TOKEN` as parent unless `VINIT_ORPHAN_TOKEN` is `true`
    - Token roles?
  - Token should be renewable unless `VINIT_DISABLE_RENEW` is `true`
  - Token should be provided to child as `VAULT_TOKEN` unless `VINIT_NO_INHERIT_TOKEN` is `true`
- Use Go's `text/template` library to do templating into environment variables and files in the container
  - Template context loaded in based on comma-separated `VINIT_PATHS`
    - Example:
      - `export VINIT_PATHS="/secret/services/concourse"`
      - `export VINIT_PATHS="/secret/services/sourcegraph,/secret/services/oauth2-proxy/sourcegraph"`
  - When multiple paths are provided, try to contextually diff the URLs to create nested structure
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