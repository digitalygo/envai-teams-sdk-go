# Migration to contract 3 (breaking)

New clients automatically send `X-Envai-Contract-Version: 3` for ATP, completion reads, and authenticated Wiki/source operations, and require a matching successful-response acknowledgment. ATP paths are under `/api/v1/atp/v3`; keep the same configured base URL ending in `/api/v1`. Unrelated CRUD and public JWKS remain unnegotiated.

Before mutations on those surfaces, the SDK verifies versioned capabilities: protocol/contract version 3, supported version 3, and the declared historical-only v2 mode. An old server's missing route, ignored header or unsupported capability response fails closed before a mutation wherever preflight can detect it. The versioned mutation route also protects an old-server switch after preflight. `CompatibilityError` carries a cause where available. No fallback or automatically repeated mutation is performed, even if a success lacks its acknowledgment and the mutation may already have happened. Reconcile an ambiguous result with the backend before deciding whether to replay the same idempotency key and original payload.

For historical reads, construct `Config{LegacyReadOnly: true}`. It rejects every mutation, including general CRUD; it does not let old servers start paid tracked-cost work. Explicit `GetCapabilitiesLegacy` and `LegacyEvents`/`StreamChatEventsLegacy` are also available. Legacy events are finite snapshots ending in EOF, without live reconnects. The entire response is buffered and validated before any event is returned, with a 64 MiB default bound configurable through `StreamOptions.MaxLegacySnapshotBytes`. Oversize, truncated, mixed-version, malformed-cost, wrong-chat, backwards/duplicate-sequence or gapped snapshots fail without yielding a prefix; context cancellation stops the read. A legacy replay containing v3 or null-bearing v2 is rejected.

`DecodeEvent` validates version before hydration. V2 is checked against its separately frozen numeric schema. V3 accepts nullable usage/terminal costs. V3 streams accept unchanged valid numeric v2 historical events and current v3 events. Raw event bytes, IDs, sequence numbers and versions are preserved; reconnects retain selected contract 3. Unknown versions are errors.

Update scalar assumptions:

- `Completion.CostUSD`, `ATPUsage.CostUSD` and terminal cost/chat-cost fields change from `float64` to `*float64`. Check nil before dereferencing; do not default nil to zero. Numeric costs must be finite and nonnegative before hydration, including nested usage and terminal metadata.
- Wiki/accounting DTOs expose typed fields and preserve allowed unknown JSON in `Extra`. Currency amounts remain decimal strings.
- Optional terminal snapshots, billing and usage keep missing/null/value semantics. Snapshot data is metadata, not response text.
- New operations require explicit idempotency keys where the shared API requires them. Contract selection is automatic and does not add a per-call version argument.

Server-first deployment is required for normal new-client operation. Legacy default server requests can only read truthful original numeric representations; HTTP 426 requires upgrade and cannot be treated as a zero-cost response. A whole completion page may fail if any row cannot be represented truthfully. Existing numeric histories and closed empty known-zero scopes remain valid.

Full Wiki lifecycle, exact reads, webhook verification and atomic PostgreSQL intake are documented in [wiki.md](wiki.md). Provider callbacks and client-tool signatures retain separate signing protocols. No deployment, migration execution, paid provider call or credential provisioning is implied by installing the SDK.

## Alpha release prediction

A token-free local dry-run of the existing locked Semantic Release tooling, using the verified existing alpha/main topology, published tag and channel note, predicts `v1.0.0-alpha.2` for this breaking change. Its release type is `major`, but Semantic Release continues the unreleased 1.0.0 alpha line. The existing module path and major-version guard remain valid and unchanged. This prediction is a past measurement for the earlier breaking change, not a publication claim; refresh tags before release. The project operations listed below are unreleased, and this guide asserts no tag, publication or version for them.

## Project operations (unreleased, additive)

On `feature/user-company-projects` the reviewed subset grew from 78 to 92 operations (75 to 89 callable, 3 unchanged API-key stubs) against the pinned canonical `digitalygo/envai-teams` commit `7f4085dbfdd3f208b82716119cff1860fbd502bb` (146 server operations). The published `v1.0.0-alpha.1` does not contain the 14 additions: `DeleteWorkspace`, `UpdateChat`, `DeleteChat`, `ListChatCrons`, `GetChatCron`, `UpdateChatCron`, `DeleteChatCron`, `ListProjectSkills`, `CreateProjectSkill`, `GetProjectSkill`, `UpdateProjectSkill`, `DeleteProjectSkill`, `GetProjectUsage` and `GetProjectLogs`. They add no required migration for existing callers:

- No ATP event types changed. The 39 typed events stay as they are (`chat.cron.wakeup`, `chat.title.updated`, `billing.limit_reached` and `billing.credit_exhausted` were already typed); billing events may have a null `turn_id` before admission or a real `turn_id` midturn and are not turn terminals, a null terminal cost is still unknown, v2 remains frozen, and native tasks stay opt-in.
- Deletion returns a `CleanupResult` with pending cleanup, not a physical purge. 409 conflicts (such as `turn_in_progress` or active cron protection) and 402 or 503 failures return `*APIError` with the server `Status`, `Type` and `Code`; mutations are never auto-cancelled or retried. Active cron protection can persist even after turn cancellation without guaranteeing removal of the deletion guard; inspect the error code and existing cron, occurrence, and turn state to take explicit operator action rather than using destructive turn cancellation as a generic recipe. The new 503 `provider_configuration_error` applies to image generation.
- `ProjectSkillCreateRequest` and `ProjectSkillUpdateRequest` remain maps of `json.RawMessage`; they are not structs.
- Human administration, financial and API-key management operations remain unsupported.

Details and samples are in the [API reference](api-reference.md#project-resources-deletion-and-observability).

## Project tenancy and authentication

Workspaces, chats, agents and teams replace UserID ownership with ProjectID and nullable CreatedByUserID audit provenance. Completion adds ProjectID and APIKeyID, and UserID is now a nullable *string. Use project ownership checks, never the key creator. Workspace metadata is a json.RawMessage alias supporting objects and lists, including the server default [].

Tool registration is project-wide, not user-wide. Use ProjectTool methods and models. Deprecated UserTool aliases forward to project-scoped operations; they do not retain user-wide visibility. A project key selects its project, and any supplied project header must match it. Definition normalization ignores server-owned identity and audit metadata.

The byte-exact server snapshot contains 146 operations. The explicit scripts/sdk-operations.json allowlist retains a reviewed 92-operation subset, not entire-server parity: 89 callable operations plus 3 discoverable API-key administration stubs. API-key list/create/delete fail locally before HTTP or capability discovery with ErrAPIKeyManagementUnsupported (errors.Is supported). The server requires a human session and CSRF for /api/v1/api-keys; use its authenticated application flow. This client never bypasses that boundary. Human financial and administration operations are not generated.

Explicit historical legacy completion reads retain their frozen numeric/owner contract; current strict response validation still requires project ownership fields.
