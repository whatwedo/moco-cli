# moco-cli

A command-line client for the [MOCO API v1](https://docs.mocoapp.com/api/docs/v1.yaml).

`moco-cli` covers **all** MOCO API endpoints. The commands are generated directly
from the official OpenAPI specification, so the CLI stays in sync with the API.
Command help texts are in German; the codebase itself is English.

## Installation

```sh
go install github.com/whatwedo/moco-cli/cmd/moco@latest
```

Pre-built binaries for Linux, macOS and Windows are attached to each
[GitHub release](https://github.com/whatwedo/moco-cli/releases).

## Authentication

`moco` needs your MOCO host and an API token. Prefer the environment variables —
a token passed as `--token` is visible to other local users via the process list
(`ps`) and tends to end up in your shell history:

```sh
export MOCO_ENDPOINT=whatwedo.mocoapp.com
export MOCO_TOKEN=your-api-token

moco projects list
```

The `--endpoint` and `--token` flags exist for one-off use and take precedence
over the environment variables. You can find your personal API token in MOCO
under *Profile → Integrations*.

## Usage

Commands are grouped by resource (the OpenAPI tag), with a verb per operation:

```sh
moco projects list                 # GET  /projects
moco projects get 123              # GET  /projects/123
moco projects create --name "New"  # POST /projects
moco projects update 123 --name "Renamed"
moco projects delete 123
moco invoices send-email 42        # custom action: POST /invoices/42/send_email
```

Explore the available commands and flags with `--help` at any level:

```sh
moco --help
moco projects --help
moco projects list --help
```

### Path parameters, query parameters and request bodies

- **Path parameters** are positional arguments: `moco projects get <id>`.
- **Query parameters** are flags: `moco activities list --from 2026-01-01 --to 2026-01-31`.
- **Request bodies**: scalar fields are exposed as flags; for full or nested
  payloads use `--data` with a raw JSON object. Flags override fields from `--data`.

```sh
# Scalar flags
moco companies create --name "ACME" --type customer

# Raw JSON body (e.g. for nested fields)
moco projects create --data '{"name":"New","customer_id":1,"finish_date":"2026-12-31"}'

# Combined: --data as base, flags override individual fields
moco projects create --data '{"customer_id":1}' --name "New"
```

## Output

Responses are printed as pretty JSON on stdout, which makes `moco` easy to
combine with [`jq`](https://jqlang.github.io/jq/):

```sh
moco projects list | jq '.[].name'
```

Use `--output raw` to pass the response body through unchanged. Errors and
diagnostics go to stderr. Exit codes: `0` success, `1` API error, `2` usage or
configuration error.

## Development

```sh
go test ./...        # run the tests
go generate ./...    # regenerate the commands from MOCO's OpenAPI spec
go build ./cmd/moco  # build the binary
```

The generator lives in [`tools/gen`](tools/gen). It fetches MOCO's OpenAPI spec
from `https://docs.mocoapp.com/api/docs/v1.yaml` at generation time — MOCO's spec
is MOCO's own work and is deliberately **not** vendored into this repository.
The generated command files (`internal/commands/*_gen.go`) are committed, so
building and `go install` need no network access; only regeneration does. Run
`go generate ./...` after MOCO updates the API or after editing the translation
tables. Pass `-spec <path>` to the generator to use a local spec copy instead.

## License

Licensed under the [GNU Affero General Public License v3.0](LICENSE) or later.
