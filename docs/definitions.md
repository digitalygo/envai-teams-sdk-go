# Deploying definitions

Definitions tooling compiles local desired state, reviews changes against the
last applied state and current remote resources, and explicitly applies them.
It never creates server resources or database tables at import or client startup.
Handlers stay in the normal Go `Registry`; manifests contain only remote tool
schemas, agent configuration and team membership.

```go
manifest, err := envai.LoadDefinitions("definitions")
if err != nil { return err }
service := envai.Definitions{
    Remote: envai.ClientDefinitionRemote{Client: client},
    Store: &envai.FileDefinitionStore{Directory: ".envai-state"},
    Scope: "production/account-123", // public stable identity, not an API key
}
plan, err := service.Plan(ctx, manifest, envai.DefinitionOptions{})
if err != nil { return err }
// Present plan.Operations for review before explicitly applying.
_ = plan
report, err := service.Sync(ctx, manifest, envai.DefinitionOptions{})
if err != nil { return err }
if !report.ManifestReady { return fmt.Errorf("definitions: %s", report.Status) }
team, err := service.Lookup(ctx, manifest, "team", "support")
if err != nil { return err }
_ = team.RemoteID
```

Use a JSON file with `format_version: "1"` and object maps `tools`, `agents`,
`teams`, keyed by stable local identities. A directory can contain `manifest.json`
and `agents/*.md` with YAML frontmatter and the system prompt as the Markdown body.
Tool name must equal its key; agent and team names can change while keeping the
same stored remote ID. Team `orchestrator_id` and member `agent_definition_id`
refer to local agent keys; synchronization substitutes actual returned IDs.

```json
{
  "format_version": "1",
  "tools": {},
  "agents": {
    "coordinator": {
      "name": "Support coordinator", "mode": "primary",
      "system_prompt": "Coordinate support.",
      "can_write_knowledge": false, "can_write_skills": false
    }
  },
  "teams": {
    "support": {"name": "Support", "orchestrator_id": "coordinator", "members": []}
  }
}
```

Client tools require `kind: "client"`, `description`, an object-root JSON Schema
`parameters` and integer `timeout_seconds` between 1 and 600. Agent optional
`description`, `model`, `tools_enabled`, `external_tools_enabled`, `permissions`
and `config` retain explicit presence. `max_tokens` inside config is rejected.
Local validation needs no API credentials; planning and synchronization fetch
the model catalog and fail closed when it cannot be obtained.

## Commands

Install `go install github.com/digitalygo/envai-teams-sdk-go/cmd/envai@<version>`.
For local development use `go run ./cmd/envai`:

```sh
envai definitions:validate --path definitions --json
envai definitions:plan --connection production/account-123 --path definitions --json
envai definitions:sync --connection production/account-123 --path definitions --dry-run
envai definitions:status --connection production/account-123 --path definitions
envai definitions:state --connection production/account-123
```

Configuration: `ENVAI_BASE_URL`, `ENVAI_API_KEY`, `ENVAI_CONNECTION`,
`ENVAI_DEFINITIONS_PATH`, `ENVAI_STATE_DRIVER`, `ENVAI_STATE_PATH`,
`ENVAI_STATE_DSN`. HTTP is disabled unless `ENVAI_ALLOW_HTTP=true` (local tests).
Flags `--only tools|agents|teams|all`, `--adopt`, `--overwrite-drift`,
`--recreate-missing`, `--prune`, `--fail-on-drift` are explicit. Exit 2 means a
blocked plan/drift or an unready status, 1 means a runtime/configuration failure.
`--json` emits machine-readable data. No flags are inferred from noninteractive use.

The standard CLI links PostgreSQL and file stores. The optional
`integration/tooling/cmd/envai` binary links MySQL and pure-Go SQLite too; build
or run it from that module. Applications can call `RunDefinitionsCLI` from their
own binary with chosen database/sql drivers. This avoids forcing SQLite, MySQL
or CGO dependencies into every SDK consumer.

## State and failure semantics

`FileDefinitionStore` uses a separate process lock and private atomic snapshots.
Use a trusted private directory on a local filesystem; Unix file locks are
supported. Other platforms can use the database stores. `MemoryDefinitionStore`
is explicitly ephemeral. `PostgresDefinitionStore` uses session advisory locks;
`SQLDefinitionStore` supports SQLite/MySQL through caller-provided `*sql.DB`.
Apply `PostgresDefinitionStoreSchema` or the statements from
`DefinitionSQLSchema("sqlite"|"mysql")` through your migration system first.
No constructor or synchronization call performs DDL.

SQLite/MySQL lock rows have owner tokens and no automatic expiry. A crashed
holder leaves the lock fail-closed. After investigating remote effects, remove
only the abandoned scope/owner row through your administrative migration tooling.
Never steal a lock from an active deployment. SQLite is a local database, not a
multi-host coordination service. All processes sharing a scope must share its
store. SQL checkpoints commit independently: later operation failures preserve
earlier successful resource mappings; there is no fictitious server rollback.

Default policy blocks remote drift, conflicts, unmanaged same-name collisions and
missing managed resources. Conflict remains blocked with all override flags.
Adoption without overwrite records actual remote state but is not ready when it
differs from desired; a later plan shows the update. Prune only deletes resources
managed in this scope, in team/agent/tool order after successful upserts.

Mutation intent is saved before the HTTP request. Lost responses trigger one
readback, never an automatic mutation retry. Matching authoritative readback
resolves the write; otherwise the report is ambiguous and stops. Later applies
can resolve matching pending payloads or confirmed deletions but cannot blindly
recreate an ambiguous resource. Dry runs write nothing; partial scopes never
mark the full manifest ready. A blocked or failed real apply clears stale readiness.

`RequireSynchronized` and `Lookup` compare local manifest identity with the stored
ready marker and never make network calls. They do not detect subsequent external
remote edits until another plan/apply observes them. `Ensure` has explicit `lazy`
mode with hash/TTL checks; `manual`, `deploy` and `off` make no implicit calls.
The next lazy apply always replans under the exclusive store lock.

The portable conformance fixtures are in `testdata/parity/tooling-golden.json`.
They preserve historical PHP behavior; `safety-overrides.json` specifies the
coordinated correction preventing false readiness after differing adoption.

MySQL scope columns must use `VARBINARY(191)` in both state and lock tables.
The adapter checks existing table metadata before acquiring a lock and rejects
older case-insensitive `VARCHAR` schemas. Apply an explicit application-owned
migration after resolving any pre-existing scope collisions; the SDK never
alters an existing schema automatically.

Set `--strict-validation` or `ENVAI_STRICT_VALIDATION=true` to validate canonical HTTP request and response schemas. The default is false for compatibility; malformed boolean settings fail before any network request. Python validates canonical HTTP payloads by default.
