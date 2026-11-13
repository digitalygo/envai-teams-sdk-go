# Per-turn model selection

[README](../README.md) · [Getting started](getting-started.md) · [API reference](api-reference.md) · [Streaming](streaming.md) · [Chat attachments](chat-attachments.md)

This is unreleased source, verified with offline tests and fake providers only. The bundled contract is fixed at `digitalygo/envai-teams` commit `f970714c0a358ff94b14f15eabde197c55861e74`. That is a source pin, not a claim about any deployed server.

## Usage

Set `Model` to pin the root orchestrator model for one turn. `ChatMessageRequest.Model` is an `Optional[string]` used by `SendChatMessage`. `TurnOptions.Model` is the same field for `RunTurn`.

```go
list, err := client.ListModels(ctx, envai.ListModelsParams{})
if err != nil {
    return err
}
var model string
for _, raw := range list.Data {
    encoded, err := json.Marshal(raw)
    if err != nil {
        return err
    }
    var entry envai.LLMModelEntry
    if json.Unmarshal(encoded, &entry) != nil || entry.Type != "llm" || !entry.Selectable {
        continue
    }
    model = entry.ID
    break
}
if model == "" {
    return errors.New("no model reported as selectable")
}

result, err := client.RunTurn(ctx, chat.ID, "Summarize the project.", envai.TurnOptions{
    Model:          envai.Present(model),
    IdempotencyKey: "summary-1",
})
if err != nil {
    return err
}

next, err := client.RunTurn(ctx, chat.ID, "And the risks?", envai.TurnOptions{})
if err != nil {
    return err
}

accepted, err := client.SendChatMessage(ctx, envai.SendChatMessageParams{
    ChatID: chat.ID,
    Body:   envai.ChatMessageRequest{Message: "Draft a reply.", Model: envai.Present(model)},
})
if err != nil {
    return err
}
```

The fragment assumes existing imports (`context`, `encoding/json`, `errors`, `github.com/digitalygo/envai-teams-sdk-go`) and existing `ctx`, `client`, and `chat` variables. It returns errors between calls, and subsequent calls omitting `Model` use the normal configured cascade.

## Rules

- Use an exact, case-sensitive `id` copied from an `llm` entry of `GET /models`. The SDK does not trim, fold case, translate aliases or prefetch the catalog. It keeps no client-side model list and applies no allowlist.
- `selectable` is guidance only. `Selectable == false` is retained as received and the SDK does not block the request; the server decides.
- The model applies to the root of that one turn. It never changes team, agent, chat or default configuration and is not carried to the next call. Leave `Model` as the zero `Optional[string]` for the normal configured cascade.
- `envai.Null[string]()`, an empty string, a padded string (including Unicode spaces) and a raw `Extra["model"]` that is `null` or not a string are rejected before message submission, so no message POST is sent. Observation and read requests such as opening the live event stream can precede validation. None is treated as omission or a fallback. This holds with and without `StrictValidation`.
- An explicit `Model` needs no native tasks or `StrictModel`. If you also set `StrictModel`, its existing gates (`NativeTasks` and the advertised `strict_model` feature) still apply. With `StrictModel` unset, false or true, an explicit model is pinned exactly with no fallback.
- `AttachmentIDs` require `IdempotencyKey`. On a deliberate retry send the same `Model`, `StrictModel` and native policy with the same key. A changed or omitted `Model` answers `409 idempotency_conflict`. The SDK never resends or falls back to another model on its own.
- Failures are `*envai.APIError` with `Status` and `Code`: `422` `invalid_model`, `unknown_model`, `model_unavailable`, `model_incompatible`, and `503` `model_catalog_unavailable`. There is no `model_not_selectable` error.
- `agent.started` `model` and the agent envelope `model` report the selected logical id. On `agent.done`, `RequestedModel` (the logical request) and `Model` (the provider-reported model) are separate optional fields and can differ. Their existing native-task gating is unchanged.
- Server restart or eviction loses the resident buffer with no paid rerun; completion terminal snapshot metadata is available only when recorded. Reconcile with the original key and policy.
- Only chat messages select a model. Jobs, delegation, deliveries and wiki endpoints have no model selection.
