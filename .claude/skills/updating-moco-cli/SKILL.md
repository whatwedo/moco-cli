---
name: updating-moco-cli
description: Use when the MOCO CLI must pick up a newer MOCO OpenAPI spec, when MOCO adds or removes API endpoints, or when regenerating the commands under internal/commands
---

# Updating moco-cli from the MOCO OpenAPI spec

## Overview

Every command in `internal/commands` is generated from MOCO's live OpenAPI spec
(`https://docs.mocoapp.com/api/docs/v1.yaml`, never vendored). Updating the CLI
means regenerating, then closing the translation gap the new spec opens — a
generated command whose German description is missing falls back to the English
summary and ships that way.

## Steps

1. **Regenerate.** Call the generator directly, not through `go generate`:

   ```bash
   go run ./tools/gen -out internal/commands
   ```

   `go generate ./...` exits 1 with `open .../<group>_gen.go: no such file or
   directory` whenever the new spec drops a group — the generator deletes the
   stale file that `go generate` still wants to scan. The generation itself
   succeeded; only the exit code is wrong.

   The `gen: skipping … (reserved or duplicate)` lines on stderr are normal —
   flags colliding with `--endpoint`, `--token`, `--output`, `--data`.

2. **Close the translation gap.** Count the untranslated commands:

   ```bash
   go run ./tools/gen -dump 2>&1 >/dev/null | tail -1   # "N commands, 0 English fallbacks"
   go run ./tools/gen -dump 2>/dev/null | grep FALLBACK  # which ones
   ```

   Fix them in `tools/gen/translations_data.go`: CRUD commands need a `nouns`
   entry per OpenAPI tag, everything else an `actions` entry keyed
   `"<group>/<name>"`. Take the German terms from the `<sub>Synonyme: …</sub>`
   hints in the spec's tag descriptions — that is MOCO's own terminology:

   ```bash
   curl -sL https://docs.mocoapp.com/api/docs/v1.yaml -o /tmp/v1.yaml
   grep -n "name: <NewTag>" -A 10 /tmp/v1.yaml
   ```

   Also delete `nouns`/`actions` entries for tags the new spec no longer has.
   **Ship only at 0 fallbacks**, then regenerate once more so the new texts land
   in the generated files.

3. **Verify.** The `internal/cli` tests read the real environment, so a set
   `MOCO_TOKEN`/`MOCO_ENDPOINT` makes `TestExecuteMissingConfig` fail with a
   live 404. Always unset:

   ```bash
   go build ./... && go vet ./... && \
     env -u MOCO_TOKEN -u MOCO_ENDPOINT -u MOCO_API_KEY -u MOCO_SUBDOMAIN go test ./...
   ```

   Smoke-test each new group: `go run ./cmd/moco <group> --help`.

4. **Commit.** Dependencies (`go get -u ./... && go mod tidy`) go in their own
   `chore(deps)` commit. The regenerated commands plus their translations are
   one `feat(commands)` commit — a removed group makes it breaking, because a
   CLI command disappears:

   ```
   feat(commands)!: sync commands with the current MOCO API

   BREAKING CHANGE: remove `moco <group>` calls from scripts — MOCO dropped the endpoint.
   ```

   Follow `/entwicklung:git-commit-messages` for the wording.

5. **Release.** Add the section to `CHANGELOG.md` (`Added` for new groups,
   `Removed` for dropped ones) plus its compare link, commit as
   `chore(release): vX.Y.Z`, then tag and push:

   ```bash
   git tag -s vX.Y.Z -m "moco-cli vX.Y.Z"
   git push origin main && git push origin vX.Y.Z
   ```

   A dropped command group means a minor bump at minimum. The tag triggers
   GoReleaser and the Homebrew cask.

## Common mistakes

| Mistake | Consequence |
| --- | --- |
| Trusting `go generate ./...`'s exit code | Looks like a failure although the code was written |
| Committing with English fallbacks left | Untranslated help texts ship to users |
| Leaving `nouns` entries for removed tags | Dead translations nobody notices going stale |
| Editing `internal/commands/*_gen.go` | Overwritten on the next run — fix the generator or the translations |
| `feat` instead of `feat!` when a group vanishes | Breaking CLI change missing from the release notes |
