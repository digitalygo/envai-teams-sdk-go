# Migration to contract 3 (breaking)

New clients automatically send `X-Envai-Contract-Version: 3` for ATP, completion reads, and authenticated Wiki/source operations, and require a matching successful-response acknowledgment. ATP paths are under `/api/v1/atp/v3`; keep the same configured base URL ending in `/api/v1`. Unrelated CRUD and public JWKS remain unnegotiated.

Before mutations on those surfaces, the SDK verifies versioned capabilities: protocol/contract version 3, supported version 3, and the declared historical-only v2 mode. An old server's missing route, ignored header or unsupported capability response fails closed before a mutation wherever preflight can detect it. The versioned mutation route also protects an old-server switch after preflight. `CompatibilityError` carries a cause where available. No fallback or automatically repeated mutation is performed, even if a success lacks its acknowledgment and the mutation may already have happened. Reconcile an ambiguous result with the backend before deciding whether to replay the same idempotency key and original payload.

For historical reads, construct `Config{LegacyReadOnly: true}`. It rejects every mutation, including general CRUD; it does not let old servers start paid tracked-cost work. Explicit `GetCapabilitiesLegacy` and `LegacyEvents`/`StreamChatEventsLegacy` are also available. Legacy events are finite snapshots ending in EOF, without live reconnects. The entire response is buffered and validated before any event is returned, with a 64 MiB default bound configurable through `StreamOptions.MaxLegacySnapshotBytes`. Oversize, truncated, mixed-version, malformed-cost or gapped snapshots fail without yielding a prefix; context cancellation stops the read. A legacy replay containing v3 or null-bearing v2 is rejected.

`DecodeEvent` validates version before hydration. V2 is checked against its separately frozen numeric schema. V3 accepts nullable usage/terminal costs. V3 streams accept unchanged valid numeric v2 historical events and current v3 events. Raw event bytes, IDs, sequence numbers and versions are preserved; reconnects retain selected contract 3. Unknown versions are errors.

Update scalar assumptions:

- `Completion.CostUSD`, `ATPUsage.CostUSD` and terminal cost/chat-cost fields change from `float64` to `*float64`. Check nil before dereferencing; do not default nil to zero. Numeric costs must be finite and nonnegative before hydration, including nested usage and terminal metadata.
- Wiki/accounting DTOs expose typed fields and preserve allowed unknown JSON in `Extra`. Currency amounts remain decimal strings.
- Optional terminal snapshots, billing and usage keep missing/null/value semantics. Snapshot data is metadata, not response text.
- New operations require explicit idempotency keys where the shared API requires them. Contract selection is automatic and does not add a per-call version argument.

Server-first deployment is required for normal new-client operation. Legacy default server requests can only read truthful original numeric representations; HTTP 426 requires upgrade and cannot be treated as a zero-cost response. A whole completion page may fail if any row cannot be represented truthfully. Existing numeric histories and closed empty known-zero scopes remain valid.

Full Wiki lifecycle, exact reads, webhook verification and atomic PostgreSQL intake are documented in [wiki.md](wiki.md). Provider callbacks and client-tool signatures retain separate signing protocols. No deployment, migration execution, paid provider call or credential provisioning is implied by installing the SDK.

## Alpha release prediction

A token-free local dry-run of the existing locked Semantic Release tooling, using the verified existing alpha/main topology, published tag and channel note, predicts `v1.0.0-alpha.2` for this breaking change. Its release type is `major`, but Semantic Release continues the unreleased 1.0.0 alpha line. The existing module path and major-version guard remain valid and unchanged. This prediction is not a publication claim; refresh tags before release.
