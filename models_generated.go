package envai

import "encoding/json"

type NativeDeliveryRequest struct {
	WorkspaceID  string         `json:"workspace_id"`
	EventID      string         `json:"event_id"`
	OriginTurnID string         `json:"origin_turn_id"`
	TaskID       string         `json:"task_id"`
	Output       string         `json:"output"`
	Status       DeliveryStatus `json:"status"`
}

type NativeDeliveryReceipt struct {
	EventID       string                `json:"event_id"`
	Status        DeliveryReceiptStatus `json:"status"`
	TurnID        Optional[string]      `json:"turn_id,omitzero"`
	ErrorEnvelope Optional[string]      `json:"error,omitzero"`
	Duplicate     bool                  `json:"duplicate"`
}

type TeamDelegationRequest struct {
	WorkspaceID  string         `json:"workspace_id"`
	EventID      string         `json:"event_id"`
	OriginTurnID string         `json:"origin_turn_id"`
	TargetChatID string         `json:"target_chat_id"`
	Prompt       string         `json:"prompt"`
	NotifyOrigin Optional[bool] `json:"notify_origin,omitzero"`
	StrictModel  Optional[bool] `json:"strict_model,omitzero"`
}

type TeamDelegationReceipt struct {
	TaskID       string `json:"task_id"`
	TargetChatID string `json:"target_chat_id"`
	TurnID       string `json:"turn_id"`
}

type DetachedTaskSnapshot struct {
	TaskID        string                     `json:"task_id"`
	OriginTurnID  string                     `json:"origin_turn_id"`
	ParentTaskID  Optional[string]           `json:"parent_task_id,omitzero"`
	TargetChatID  Optional[string]           `json:"target_chat_id,omitzero"`
	ResultTurnID  Optional[string]           `json:"result_turn_id,omitzero"`
	Output        Optional[string]           `json:"output,omitzero"`
	ErrorEnvelope Optional[string]           `json:"error,omitzero"`
	DeliveryError Optional[string]           `json:"delivery_error,omitzero"`
	Status        TaskStatus                 `json:"status"`
	Extra         map[string]json.RawMessage `json:"-"`
}

func (v *DetachedTaskSnapshot) UnmarshalJSON(b []byte) error {
	type plain DetachedTaskSnapshot
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"task_id", "origin_turn_id", "parent_task_id", "target_chat_id", "result_turn_id", "output", "error", "delivery_error", "status"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = DetachedTaskSnapshot(out)
	return nil
}
func (v DetachedTaskSnapshot) MarshalJSON() ([]byte, error) {
	type plain DetachedTaskSnapshot
	return marshalUnknown(plain(v), v.Extra)
}

type PhaseClassificationRequest struct {
	RequestID   string          `json:"request_id"`
	PhaseID     string          `json:"phase_id"`
	FinalOutput string          `json:"final_output"`
	TaskContext json.RawMessage `json:"task_context"`
	Attempt     int64           `json:"attempt"`
}

type PhaseClassificationResult struct {
	Model        string        `json:"model"`
	RequestID    string        `json:"request_id"`
	InvocationID string        `json:"invocation_id"`
	Decision     PhaseDecision `json:"decision"`
}

type CompletionTerminalSnapshotUsage struct {
	InputTokens   Optional[int64]        `json:"input_tokens,omitzero"`
	OutputTokens  Optional[int64]        `json:"output_tokens,omitzero"`
	DurationMs    Optional[int64]        `json:"duration_ms,omitzero"`
	SpriteMs      Optional[int64]        `json:"sprite_ms,omitzero"`
	SpriteRamGbMs Optional[int64]        `json:"sprite_ram_gb_ms,omitzero"`
	ToolCalls     Optional[int64]        `json:"tool_calls,omitzero"`
	CostUSD       Optional[float64]      `json:"cost_usd,omitzero"`
	Billing       Optional[BillingState] `json:"billing,omitzero"`
	UsageCost     Optional[UsageCost]    `json:"usage_cost,omitzero"`
}

type CompletionTerminalSnapshot struct {
	EventID                 string                         `json:"event_id"`
	OccurredAt              string                         `json:"occurred_at"`
	EventType               string                         `json:"event_type"`
	Data                    CompletionTerminalSnapshotData `json:"data"`
	V                       Optional[int64]                `json:"v,omitzero"`
	AcceptedProtocolVersion Optional[int64]                `json:"accepted_protocol_version,omitzero"`
}

type CompletionTerminalSnapshotChatUsage struct {
	InputTokens   int64                  `json:"input_tokens"`
	OutputTokens  int64                  `json:"output_tokens"`
	CostUSD       *float64               `json:"cost_usd"`
	DurationMs    int64                  `json:"duration_ms"`
	SpriteMs      Optional[int64]        `json:"sprite_ms,omitzero"`
	SpriteRamGbMs Optional[int64]        `json:"sprite_ram_gb_ms,omitzero"`
	ToolCalls     Optional[int64]        `json:"tool_calls,omitzero"`
	Billing       Optional[BillingState] `json:"billing,omitzero"`
	UsageCost     Optional[UsageCost]    `json:"usage_cost,omitzero"`
}

type BillingState struct {
	Status string                     `json:"status"`
	Extra  map[string]json.RawMessage `json:"-"`
}

func (v *BillingState) UnmarshalJSON(b []byte) error {
	if err := validateDefinition("wiki.json", "BillingState", b); err != nil {
		return err
	}
	type plain BillingState
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"status"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = BillingState(out)
	return nil
}
func (v BillingState) MarshalJSON() ([]byte, error) {
	type plain BillingState
	return marshalUnknown(plain(v), v.Extra)
}

type CurrencyAmount struct {
	Currency string                     `json:"currency"`
	Amount   string                     `json:"amount"`
	Extra    map[string]json.RawMessage `json:"-"`
}

func (v *CurrencyAmount) UnmarshalJSON(b []byte) error {
	if err := validateDefinition("wiki.json", "CurrencyAmount", b); err != nil {
		return err
	}
	type plain CurrencyAmount
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"currency", "amount"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = CurrencyAmount(out)
	return nil
}
func (v CurrencyAmount) MarshalJSON() ([]byte, error) {
	type plain CurrencyAmount
	return marshalUnknown(plain(v), v.Extra)
}

type UsageCost struct {
	Status                 string                     `json:"status"`
	Basis                  string                     `json:"basis"`
	ScopeStatus            string                     `json:"scope_status"`
	Revision               int64                      `json:"revision"`
	AsOf                   string                     `json:"as_of"`
	InvocationCount        int64                      `json:"invocation_count"`
	PendingInvocationCount int64                      `json:"pending_invocation_count"`
	Totals                 []CurrencyAmount           `json:"totals"`
	PricedSubtotals        []CurrencyAmount           `json:"priced_subtotals"`
	Extra                  map[string]json.RawMessage `json:"-"`
}

func (v *UsageCost) UnmarshalJSON(b []byte) error {
	if err := validateDefinition("wiki.json", "UsageCost", b); err != nil {
		return err
	}
	type plain UsageCost
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"status", "basis", "scope_status", "revision", "as_of", "invocation_count", "pending_invocation_count", "totals", "priced_subtotals"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = UsageCost(out)
	return nil
}
func (v UsageCost) MarshalJSON() ([]byte, error) {
	type plain UsageCost
	return marshalUnknown(plain(v), v.Extra)
}

type ConversionState struct {
	Status      string                     `json:"status"`
	ErrorCode   Optional[string]           `json:"error_code,omitzero"`
	Recoverable Optional[bool]             `json:"recoverable,omitzero"`
	NextRetryAt Optional[string]           `json:"next_retry_at,omitzero"`
	Extra       map[string]json.RawMessage `json:"-"`
}

func (v *ConversionState) UnmarshalJSON(b []byte) error {
	if err := validateDefinition("wiki.json", "ConversionState", b); err != nil {
		return err
	}
	type plain ConversionState
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"status", "error_code", "recoverable", "next_retry_at"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ConversionState(out)
	return nil
}
func (v ConversionState) MarshalJSON() ([]byte, error) {
	type plain ConversionState
	return marshalUnknown(plain(v), v.Extra)
}

type IndexingState struct {
	Status string                     `json:"status"`
	Extra  map[string]json.RawMessage `json:"-"`
}

func (v *IndexingState) UnmarshalJSON(b []byte) error {
	if err := validateDefinition("wiki.json", "IndexingState", b); err != nil {
		return err
	}
	type plain IndexingState
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"status"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = IndexingState(out)
	return nil
}
func (v IndexingState) MarshalJSON() ([]byte, error) {
	type plain IndexingState
	return marshalUnknown(plain(v), v.Extra)
}

type WikiSource struct {
	FileID           string                     `json:"file_id"`
	WorkspaceID      string                     `json:"workspace_id"`
	OriginalName     string                     `json:"original_name"`
	MimeType         string                     `json:"mime_type"`
	ByteSize         int64                      `json:"byte_size"`
	Lifecycle        string                     `json:"lifecycle"`
	SourceRevision   int64                      `json:"source_revision"`
	Conversion       ConversionState            `json:"conversion"`
	Indexing         IndexingState              `json:"indexing"`
	CanonicalRawPath *string                    `json:"canonical_raw_path"`
	Billing          BillingState               `json:"billing"`
	UsageCost        UsageCost                  `json:"usage_cost"`
	StatusURL        string                     `json:"status_url"`
	Extra            map[string]json.RawMessage `json:"-"`
}

func (v *WikiSource) UnmarshalJSON(b []byte) error {
	if err := validateDefinition("wiki.json", "WikiSource", b); err != nil {
		return err
	}
	type plain WikiSource
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"file_id", "workspace_id", "original_name", "mime_type", "byte_size", "lifecycle", "source_revision", "conversion", "indexing", "canonical_raw_path", "billing", "usage_cost", "status_url"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = WikiSource(out)
	return nil
}
func (v WikiSource) MarshalJSON() ([]byte, error) {
	type plain WikiSource
	return marshalUnknown(plain(v), v.Extra)
}

type BatchSizes struct {
	Inbox  int64                      `json:"inbox"`
	Outbox int64                      `json:"outbox"`
	Extra  map[string]json.RawMessage `json:"-"`
}

func (v *BatchSizes) UnmarshalJSON(b []byte) error {
	if err := validateDefinition("wiki.json", "BatchSizes", b); err != nil {
		return err
	}
	type plain BatchSizes
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"inbox", "outbox"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = BatchSizes(out)
	return nil
}
func (v BatchSizes) MarshalJSON() ([]byte, error) {
	type plain BatchSizes
	return marshalUnknown(plain(v), v.Extra)
}

type FileOutcome struct {
	FileID    string                     `json:"file_id"`
	Status    string                     `json:"status"`
	ErrorCode *string                    `json:"error_code"`
	Extra     map[string]json.RawMessage `json:"-"`
}

func (v *FileOutcome) UnmarshalJSON(b []byte) error {
	if err := validateDefinition("wiki.json", "FileOutcome", b); err != nil {
		return err
	}
	type plain FileOutcome
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"file_id", "status", "error_code"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = FileOutcome(out)
	return nil
}
func (v FileOutcome) MarshalJSON() ([]byte, error) {
	type plain FileOutcome
	return marshalUnknown(plain(v), v.Extra)
}

type PhaseResult struct {
	Status    string                     `json:"status"`
	Selected  int64                      `json:"selected"`
	Succeeded int64                      `json:"succeeded"`
	Withdrawn int64                      `json:"withdrawn"`
	Failed    int64                      `json:"failed"`
	Files     []FileOutcome              `json:"files"`
	Extra     map[string]json.RawMessage `json:"-"`
}

func (v *PhaseResult) UnmarshalJSON(b []byte) error {
	if err := validateDefinition("wiki.json", "PhaseResult", b); err != nil {
		return err
	}
	type plain PhaseResult
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"status", "selected", "succeeded", "withdrawn", "failed", "files"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = PhaseResult(out)
	return nil
}
func (v PhaseResult) MarshalJSON() ([]byte, error) {
	type plain PhaseResult
	return marshalUnknown(plain(v), v.Extra)
}

type WikiOperation struct {
	OperationID      string                      `json:"operation_id"`
	WorkspaceID      string                      `json:"workspace_id"`
	Kind             string                      `json:"kind"`
	Status           string                      `json:"status"`
	Phase            *string                     `json:"phase"`
	Phases           WikiOperationPhases         `json:"phases"`
	StartedAt        *string                     `json:"started_at"`
	FinishedAt       *string                     `json:"finished_at"`
	BatchSizes       BatchSizes                  `json:"batch_sizes"`
	Billing          BillingState                `json:"billing"`
	UsageCost        UsageCost                   `json:"usage_cost"`
	ReportedProblems Optional[[]ReportedProblem] `json:"reported_problems,omitzero"`
	Extra            map[string]json.RawMessage  `json:"-"`
}

func (v *WikiOperation) UnmarshalJSON(b []byte) error {
	if err := validateDefinition("wiki.json", "WikiOperation", b); err != nil {
		return err
	}
	type plain WikiOperation
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"operation_id", "workspace_id", "kind", "status", "phase", "phases", "started_at", "finished_at", "batch_sizes", "billing", "usage_cost", "reported_problems"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = WikiOperation(out)
	return nil
}
func (v WikiOperation) MarshalJSON() ([]byte, error) {
	type plain WikiOperation
	return marshalUnknown(plain(v), v.Extra)
}

type ReportedProblem struct {
	FileID  string `json:"file_id"`
	Message string `json:"message"`
}

type Category struct {
	CategoryID  string                     `json:"category_id"`
	Name        string                     `json:"name"`
	Description string                     `json:"description"`
	Extra       map[string]json.RawMessage `json:"-"`
}

func (v *Category) UnmarshalJSON(b []byte) error {
	if err := validateDefinition("wiki.json", "Category", b); err != nil {
		return err
	}
	type plain Category
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"category_id", "name", "description"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = Category(out)
	return nil
}
func (v Category) MarshalJSON() ([]byte, error) {
	type plain Category
	return marshalUnknown(plain(v), v.Extra)
}

type Categories struct {
	Categories      []Category                 `json:"categories"`
	CatalogRevision int64                      `json:"catalog_revision"`
	Extra           map[string]json.RawMessage `json:"-"`
}

func (v *Categories) UnmarshalJSON(b []byte) error {
	if err := validateDefinition("wiki.json", "Categories", b); err != nil {
		return err
	}
	type plain Categories
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"categories", "catalog_revision"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = Categories(out)
	return nil
}
func (v Categories) MarshalJSON() ([]byte, error) {
	type plain Categories
	return marshalUnknown(plain(v), v.Extra)
}

type CatalogEntry struct {
	EntryID       string                     `json:"entry_id"`
	Type          string                     `json:"type"`
	TargetID      string                     `json:"target_id"`
	CanonicalPath string                     `json:"canonical_path"`
	Description   string                     `json:"description"`
	InsertedAt    string                     `json:"inserted_at"`
	Extra         map[string]json.RawMessage `json:"-"`
}

func (v *CatalogEntry) UnmarshalJSON(b []byte) error {
	if err := validateDefinition("wiki.json", "CatalogEntry", b); err != nil {
		return err
	}
	type plain CatalogEntry
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"entry_id", "type", "target_id", "canonical_path", "description", "inserted_at"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = CatalogEntry(out)
	return nil
}
func (v CatalogEntry) MarshalJSON() ([]byte, error) {
	type plain CatalogEntry
	return marshalUnknown(plain(v), v.Extra)
}

type IndexRange struct {
	From  int64                      `json:"from"`
	To    int64                      `json:"to"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *IndexRange) UnmarshalJSON(b []byte) error {
	if err := validateDefinition("wiki.json", "IndexRange", b); err != nil {
		return err
	}
	type plain IndexRange
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"from", "to"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = IndexRange(out)
	return nil
}
func (v IndexRange) MarshalJSON() ([]byte, error) {
	type plain IndexRange
	return marshalUnknown(plain(v), v.Extra)
}

type CategoryIndex struct {
	Entries         []CatalogEntry             `json:"entries"`
	TotalCount      int64                      `json:"total_count"`
	ReturnedCount   int64                      `json:"returned_count"`
	RequestedRange  *IndexRange                `json:"requested_range"`
	ActualRange     *IndexRange                `json:"actual_range"`
	CatalogRevision int64                      `json:"catalog_revision"`
	IsComplete      bool                       `json:"is_complete"`
	Extra           map[string]json.RawMessage `json:"-"`
}

func (v *CategoryIndex) UnmarshalJSON(b []byte) error {
	if err := validateDefinition("wiki.json", "CategoryIndex", b); err != nil {
		return err
	}
	type plain CategoryIndex
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"entries", "total_count", "returned_count", "requested_range", "actual_range", "catalog_revision", "is_complete"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = CategoryIndex(out)
	return nil
}
func (v CategoryIndex) MarshalJSON() ([]byte, error) {
	type plain CategoryIndex
	return marshalUnknown(plain(v), v.Extra)
}

type WikiFileProcessedEvent struct {
	SchemaVersion    string                     `json:"schema_version"`
	EventID          string                     `json:"event_id"`
	EventType        string                     `json:"event_type"`
	OperationID      string                     `json:"operation_id"`
	WorkspaceID      string                     `json:"workspace_id"`
	OccurredAt       string                     `json:"occurred_at"`
	FileID           string                     `json:"file_id"`
	SourceRevision   int64                      `json:"source_revision"`
	CanonicalRawPath string                     `json:"canonical_raw_path"`
	Availability     string                     `json:"availability"`
	Extra            map[string]json.RawMessage `json:"-"`
}

func (v *WikiFileProcessedEvent) UnmarshalJSON(b []byte) error {
	if err := validateDefinition("wiki.json", "WikiFileProcessedEvent", b); err != nil {
		return err
	}
	type plain WikiFileProcessedEvent
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"schema_version", "event_id", "event_type", "operation_id", "workspace_id", "occurred_at", "file_id", "source_revision", "canonical_raw_path", "availability"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = WikiFileProcessedEvent(out)
	return nil
}
func (v WikiFileProcessedEvent) MarshalJSON() ([]byte, error) {
	type plain WikiFileProcessedEvent
	return marshalUnknown(plain(v), v.Extra)
}

type WikiFileDeletedEvent struct {
	SchemaVersion            string                     `json:"schema_version"`
	EventID                  string                     `json:"event_id"`
	EventType                string                     `json:"event_type"`
	OperationID              string                     `json:"operation_id"`
	WorkspaceID              string                     `json:"workspace_id"`
	OccurredAt               string                     `json:"occurred_at"`
	FileID                   string                     `json:"file_id"`
	SourceRevision           int64                      `json:"source_revision"`
	PreviousCanonicalRawPath *string                    `json:"previous_canonical_raw_path"`
	Lifecycle                string                     `json:"lifecycle"`
	Extra                    map[string]json.RawMessage `json:"-"`
}

func (v *WikiFileDeletedEvent) UnmarshalJSON(b []byte) error {
	if err := validateDefinition("wiki.json", "WikiFileDeletedEvent", b); err != nil {
		return err
	}
	type plain WikiFileDeletedEvent
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"schema_version", "event_id", "event_type", "operation_id", "workspace_id", "occurred_at", "file_id", "source_revision", "previous_canonical_raw_path", "lifecycle"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = WikiFileDeletedEvent(out)
	return nil
}
func (v WikiFileDeletedEvent) MarshalJSON() ([]byte, error) {
	type plain WikiFileDeletedEvent
	return marshalUnknown(plain(v), v.Extra)
}

type WikiOperationEvent struct {
	SchemaVersion string                     `json:"schema_version"`
	EventID       string                     `json:"event_id"`
	EventType     string                     `json:"event_type"`
	OperationID   string                     `json:"operation_id"`
	WorkspaceID   string                     `json:"workspace_id"`
	OccurredAt    string                     `json:"occurred_at"`
	Status        string                     `json:"status"`
	Phases        WikiOperationEventPhases   `json:"phases"`
	FinishedAt    string                     `json:"finished_at"`
	Billing       BillingState               `json:"billing"`
	UsageCost     UsageCost                  `json:"usage_cost"`
	Extra         map[string]json.RawMessage `json:"-"`
}

func (v *WikiOperationEvent) UnmarshalJSON(b []byte) error {
	if err := validateDefinition("wiki.json", "WikiOperationEvent", b); err != nil {
		return err
	}
	type plain WikiOperationEvent
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"schema_version", "event_id", "event_type", "operation_id", "workspace_id", "occurred_at", "status", "phases", "finished_at", "billing", "usage_cost"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = WikiOperationEvent(out)
	return nil
}
func (v WikiOperationEvent) MarshalJSON() ([]byte, error) {
	type plain WikiOperationEvent
	return marshalUnknown(plain(v), v.Extra)
}

type WikiWebhookEvent json.RawMessage

type ErrorEnvelope struct {
	ErrorEnvelope ErrorEnvelopeErrorEnvelope `json:"error"`
	Extra         map[string]json.RawMessage `json:"-"`
}

func (v *ErrorEnvelope) UnmarshalJSON(b []byte) error {
	type plain ErrorEnvelope
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"error"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ErrorEnvelope(out)
	return nil
}
func (v ErrorEnvelope) MarshalJSON() ([]byte, error) {
	type plain ErrorEnvelope
	return marshalUnknown(plain(v), v.Extra)
}

type AttachmentManifest struct {
	AttachmentID string                     `json:"attachment_id"`
	OriginalName string                     `json:"original_name"`
	MimeType     Optional[string]           `json:"mime_type,omitzero"`
	Mode         Optional[string]           `json:"mode,omitzero"`
	Revision     int64                      `json:"revision"`
	Sha256       string                     `json:"sha256"`
	Bytes        int64                      `json:"bytes"`
	Conversion   map[string]json.RawMessage `json:"conversion"`
	Extra        map[string]json.RawMessage `json:"-"`
}

func (v *AttachmentManifest) UnmarshalJSON(b []byte) error {
	if err := validateDefinition("openapi.json", "AttachmentManifest", b); err != nil {
		return err
	}
	type plain AttachmentManifest
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"attachment_id", "original_name", "mime_type", "mode", "revision", "sha256", "bytes", "conversion"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = AttachmentManifest(out)
	return nil
}
func (v AttachmentManifest) MarshalJSON() ([]byte, error) {
	type plain AttachmentManifest
	return marshalUnknown(plain(v), v.Extra)
}

type AttachmentConversion struct {
	Status      string                     `json:"status"`
	ErrorCode   *string                    `json:"error_code"`
	Recoverable bool                       `json:"recoverable"`
	NextRetryAt *string                    `json:"next_retry_at"`
	Retryable   bool                       `json:"retryable"`
	Extra       map[string]json.RawMessage `json:"-"`
}

func (v *AttachmentConversion) UnmarshalJSON(b []byte) error {
	if err := validateDefinition("openapi.json", "AttachmentConversion", b); err != nil {
		return err
	}
	type plain AttachmentConversion
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"status", "error_code", "recoverable", "next_retry_at", "retryable"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = AttachmentConversion(out)
	return nil
}
func (v AttachmentConversion) MarshalJSON() ([]byte, error) {
	type plain AttachmentConversion
	return marshalUnknown(plain(v), v.Extra)
}

type ChatAttachment struct {
	AttachmentID string                     `json:"attachment_id"`
	WorkspaceID  string                     `json:"workspace_id"`
	ChatID       string                     `json:"chat_id"`
	OriginalName string                     `json:"original_name"`
	MimeType     string                     `json:"mime_type"`
	ByteSize     int64                      `json:"byte_size"`
	Lifecycle    string                     `json:"lifecycle"`
	Revision     int64                      `json:"revision"`
	Conversion   AttachmentConversion       `json:"conversion"`
	Derived      map[string]json.RawMessage `json:"derived"`
	ExpiresAt    string                     `json:"expires_at"`
	AttachedAt   *string                    `json:"attached_at"`
	UsageCost    UsageCost                  `json:"usage_cost"`
	Extra        map[string]json.RawMessage `json:"-"`
}

func (v *ChatAttachment) UnmarshalJSON(b []byte) error {
	if err := validateDefinition("openapi.json", "ChatAttachment", b); err != nil {
		return err
	}
	type plain ChatAttachment
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"attachment_id", "workspace_id", "chat_id", "original_name", "mime_type", "byte_size", "lifecycle", "revision", "conversion", "derived", "expires_at", "attached_at", "usage_cost"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ChatAttachment(out)
	return nil
}
func (v ChatAttachment) MarshalJSON() ([]byte, error) {
	type plain ChatAttachment
	return marshalUnknown(plain(v), v.Extra)
}

type ChatAttachmentResource struct {
	Data  ChatAttachment             `json:"data"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *ChatAttachmentResource) UnmarshalJSON(b []byte) error {
	type plain ChatAttachmentResource
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ChatAttachmentResource(out)
	return nil
}
func (v ChatAttachmentResource) MarshalJSON() ([]byte, error) {
	type plain ChatAttachmentResource
	return marshalUnknown(plain(v), v.Extra)
}

type ChatAttachmentCollection struct {
	Data  ChatAttachmentCollectionData `json:"data"`
	Extra map[string]json.RawMessage   `json:"-"`
}

func (v *ChatAttachmentCollection) UnmarshalJSON(b []byte) error {
	type plain ChatAttachmentCollection
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ChatAttachmentCollection(out)
	return nil
}
func (v ChatAttachmentCollection) MarshalJSON() ([]byte, error) {
	type plain ChatAttachmentCollection
	return marshalUnknown(plain(v), v.Extra)
}

type ChatAttachmentUploadRequest struct {
	File string `json:"file"`
}

type ChatMessageRequest struct {
	Message       string                     `json:"message"`
	AttachmentIDs Optional[[]string]         `json:"attachment_ids,omitzero"`
	StrictModel   Optional[bool]             `json:"strict_model,omitzero"`
	Extra         map[string]json.RawMessage `json:"-"`
}

func (v *ChatMessageRequest) UnmarshalJSON(b []byte) error {
	type plain ChatMessageRequest
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"message", "attachment_ids", "strict_model"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ChatMessageRequest(out)
	return nil
}
func (v ChatMessageRequest) MarshalJSON() ([]byte, error) {
	type plain ChatMessageRequest
	return marshalUnknown(plain(v), v.Extra)
}

type ChatMessageAccepted struct {
	Status        string                     `json:"status"`
	TurnID        string                     `json:"turn_id"`
	AttachmentIDs Optional[[]string]         `json:"attachment_ids,omitzero"`
	Extra         map[string]json.RawMessage `json:"-"`
}

func (v *ChatMessageAccepted) UnmarshalJSON(b []byte) error {
	type plain ChatMessageAccepted
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"status", "turn_id", "attachment_ids"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ChatMessageAccepted(out)
	return nil
}
func (v ChatMessageAccepted) MarshalJSON() ([]byte, error) {
	type plain ChatMessageAccepted
	return marshalUnknown(plain(v), v.Extra)
}

type TurnAccepted struct {
	TurnID string                     `json:"turn_id"`
	Extra  map[string]json.RawMessage `json:"-"`
}

func (v *TurnAccepted) UnmarshalJSON(b []byte) error {
	type plain TurnAccepted
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"turn_id"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = TurnAccepted(out)
	return nil
}
func (v TurnAccepted) MarshalJSON() ([]byte, error) {
	type plain TurnAccepted
	return marshalUnknown(plain(v), v.Extra)
}

type ToolOutputRequest struct {
	CallID  string                     `json:"call_id"`
	Output  string                     `json:"output"`
	IsError Optional[bool]             `json:"is_error,omitzero"`
	Extra   map[string]json.RawMessage `json:"-"`
}

func (v *ToolOutputRequest) UnmarshalJSON(b []byte) error {
	type plain ToolOutputRequest
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"call_id", "output", "is_error"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ToolOutputRequest(out)
	return nil
}
func (v ToolOutputRequest) MarshalJSON() ([]byte, error) {
	type plain ToolOutputRequest
	return marshalUnknown(plain(v), v.Extra)
}

type ApprovalRequest struct {
	ApprovalID string                     `json:"approval_id"`
	Approve    bool                       `json:"approve"`
	Reason     Optional[string]           `json:"reason,omitzero"`
	Extra      map[string]json.RawMessage `json:"-"`
}

func (v *ApprovalRequest) UnmarshalJSON(b []byte) error {
	type plain ApprovalRequest
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"approval_id", "approve", "reason"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ApprovalRequest(out)
	return nil
}
func (v ApprovalRequest) MarshalJSON() ([]byte, error) {
	type plain ApprovalRequest
	return marshalUnknown(plain(v), v.Extra)
}

type SteerRequest struct {
	Text          string                     `json:"text"`
	TargetAgentID Optional[string]           `json:"target_agent_id,omitzero"`
	Extra         map[string]json.RawMessage `json:"-"`
}

func (v *SteerRequest) UnmarshalJSON(b []byte) error {
	type plain SteerRequest
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"text", "target_agent_id"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = SteerRequest(out)
	return nil
}
func (v SteerRequest) MarshalJSON() ([]byte, error) {
	type plain SteerRequest
	return marshalUnknown(plain(v), v.Extra)
}

type ForkChatRequest struct {
	DestinationChatID string                     `json:"destination_chat_id"`
	Extra             map[string]json.RawMessage `json:"-"`
}

func (v *ForkChatRequest) UnmarshalJSON(b []byte) error {
	type plain ForkChatRequest
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"destination_chat_id"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ForkChatRequest(out)
	return nil
}
func (v ForkChatRequest) MarshalJSON() ([]byte, error) {
	type plain ForkChatRequest
	return marshalUnknown(plain(v), v.Extra)
}

type ChatForked struct {
	SourceChatID      string                     `json:"source_chat_id"`
	DestinationChatID string                     `json:"destination_chat_id"`
	Extra             map[string]json.RawMessage `json:"-"`
}

func (v *ChatForked) UnmarshalJSON(b []byte) error {
	type plain ChatForked
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"source_chat_id", "destination_chat_id"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ChatForked(out)
	return nil
}
func (v ChatForked) MarshalJSON() ([]byte, error) {
	type plain ChatForked
	return marshalUnknown(plain(v), v.Extra)
}

type OrchestratorContextItem struct {
	Kind    string                     `json:"kind"`
	TurnID  Optional[string]           `json:"turn_id,omitzero"`
	Content string                     `json:"content"`
	Extra   map[string]json.RawMessage `json:"-"`
}

func (v *OrchestratorContextItem) UnmarshalJSON(b []byte) error {
	type plain OrchestratorContextItem
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"kind", "turn_id", "content"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = OrchestratorContextItem(out)
	return nil
}
func (v OrchestratorContextItem) MarshalJSON() ([]byte, error) {
	type plain OrchestratorContextItem
	return marshalUnknown(plain(v), v.Extra)
}

type RestoreChatRequest struct {
	OrchestratorContext []OrchestratorContextItem  `json:"orchestrator_context"`
	Extra               map[string]json.RawMessage `json:"-"`
}

func (v *RestoreChatRequest) UnmarshalJSON(b []byte) error {
	type plain RestoreChatRequest
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"orchestrator_context"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = RestoreChatRequest(out)
	return nil
}
func (v RestoreChatRequest) MarshalJSON() ([]byte, error) {
	type plain RestoreChatRequest
	return marshalUnknown(plain(v), v.Extra)
}

type RestoreChatAccepted struct {
	Status string                     `json:"status"`
	TurnID *string                    `json:"turn_id"`
	Extra  map[string]json.RawMessage `json:"-"`
}

func (v *RestoreChatAccepted) UnmarshalJSON(b []byte) error {
	type plain RestoreChatAccepted
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"status", "turn_id"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = RestoreChatAccepted(out)
	return nil
}
func (v RestoreChatAccepted) MarshalJSON() ([]byte, error) {
	type plain RestoreChatAccepted
	return marshalUnknown(plain(v), v.Extra)
}

type Capabilities struct {
	ProtocolVersion           int64                      `json:"protocol_version"`
	Families                  []string                   `json:"families"`
	EventTypes                []string                   `json:"event_types"`
	ToolKinds                 []string                   `json:"tool_kinds"`
	Features                  []string                   `json:"features"`
	ContractVersion           int64                      `json:"contract_version"`
	SupportedContractVersions []int64                    `json:"supported_contract_versions"`
	LegacyV2Mode              string                     `json:"legacy_v2_mode"`
	Extra                     map[string]json.RawMessage `json:"-"`
}

func (v *Capabilities) UnmarshalJSON(b []byte) error {
	type plain Capabilities
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(b, &fields); err != nil {
		return err
	}
	features, hasFeatures := fields["features"]
	delete(fields, "features")
	body, err := json.Marshal(fields)
	if err != nil {
		return err
	}
	var out plain
	if err = json.Unmarshal(body, &out); err != nil {
		return err
	}
	if hasFeatures {
		out.Features, err = normalizeFeatureNames(features)
		if err != nil {
			return err
		}
	}
	extra, err := unknownFields(body, []string{"protocol_version", "families", "event_types", "tool_kinds", "features", "contract_version", "supported_contract_versions", "legacy_v2_mode"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = Capabilities(out)
	return nil
}
func (v Capabilities) MarshalJSON() ([]byte, error) {
	type plain Capabilities
	return marshalUnknown(plain(v), v.Extra)
}

type Workspace struct {
	ID        string                     `json:"id"`
	UserID    string                     `json:"user_id"`
	Name      string                     `json:"name"`
	Metadata  WorkspaceMetadata          `json:"metadata"`
	CreatedAt string                     `json:"created_at"`
	UpdatedAt string                     `json:"updated_at"`
	Extra     map[string]json.RawMessage `json:"-"`
}

func (v *Workspace) UnmarshalJSON(b []byte) error {
	type plain Workspace
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"id", "user_id", "name", "metadata", "created_at", "updated_at"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = Workspace(out)
	return nil
}
func (v Workspace) MarshalJSON() ([]byte, error) {
	type plain Workspace
	return marshalUnknown(plain(v), v.Extra)
}

type WorkspaceResource struct {
	Data  Workspace                  `json:"data"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *WorkspaceResource) UnmarshalJSON(b []byte) error {
	type plain WorkspaceResource
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = WorkspaceResource(out)
	return nil
}
func (v WorkspaceResource) MarshalJSON() ([]byte, error) {
	type plain WorkspaceResource
	return marshalUnknown(plain(v), v.Extra)
}

type WorkspaceCollection struct {
	Data  []Workspace                `json:"data"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *WorkspaceCollection) UnmarshalJSON(b []byte) error {
	type plain WorkspaceCollection
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = WorkspaceCollection(out)
	return nil
}
func (v WorkspaceCollection) MarshalJSON() ([]byte, error) {
	type plain WorkspaceCollection
	return marshalUnknown(plain(v), v.Extra)
}

type WorkspaceMetadata map[string]json.RawMessage

type WorkspaceCreateRequest struct {
	Name     string                      `json:"name"`
	Metadata Optional[WorkspaceMetadata] `json:"metadata,omitzero"`
	Extra    map[string]json.RawMessage  `json:"-"`
}

func (v *WorkspaceCreateRequest) UnmarshalJSON(b []byte) error {
	type plain WorkspaceCreateRequest
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"name", "metadata"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = WorkspaceCreateRequest(out)
	return nil
}
func (v WorkspaceCreateRequest) MarshalJSON() ([]byte, error) {
	type plain WorkspaceCreateRequest
	return marshalUnknown(plain(v), v.Extra)
}

type WorkspaceUpdateRequest struct {
	Name     Optional[string]            `json:"name,omitzero"`
	Metadata Optional[WorkspaceMetadata] `json:"metadata,omitzero"`
	Extra    map[string]json.RawMessage  `json:"-"`
}

func (v *WorkspaceUpdateRequest) UnmarshalJSON(b []byte) error {
	type plain WorkspaceUpdateRequest
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"name", "metadata"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = WorkspaceUpdateRequest(out)
	return nil
}
func (v WorkspaceUpdateRequest) MarshalJSON() ([]byte, error) {
	type plain WorkspaceUpdateRequest
	return marshalUnknown(plain(v), v.Extra)
}

type UserTool struct {
	ID             string                     `json:"id"`
	Name           string                     `json:"name"`
	Kind           string                     `json:"kind"`
	Description    *string                    `json:"description"`
	Parameters     map[string]json.RawMessage `json:"parameters"`
	TimeoutSeconds int64                      `json:"timeout_seconds"`
	CreatedAt      string                     `json:"created_at"`
	UpdatedAt      string                     `json:"updated_at"`
}

type UserToolResource struct {
	Data  UserTool                   `json:"data"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *UserToolResource) UnmarshalJSON(b []byte) error {
	type plain UserToolResource
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = UserToolResource(out)
	return nil
}
func (v UserToolResource) MarshalJSON() ([]byte, error) {
	type plain UserToolResource
	return marshalUnknown(plain(v), v.Extra)
}

type UserToolCollection struct {
	Data  []UserTool                 `json:"data"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *UserToolCollection) UnmarshalJSON(b []byte) error {
	type plain UserToolCollection
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = UserToolCollection(out)
	return nil
}
func (v UserToolCollection) MarshalJSON() ([]byte, error) {
	type plain UserToolCollection
	return marshalUnknown(plain(v), v.Extra)
}

type UserToolCreateRequest struct {
	Name           string                     `json:"name"`
	Kind           Optional[string]           `json:"kind,omitzero"`
	Description    Optional[string]           `json:"description,omitzero"`
	Parameters     map[string]json.RawMessage `json:"parameters"`
	TimeoutSeconds Optional[int64]            `json:"timeout_seconds,omitzero"`
}

type UserToolUpdateRequest struct {
	Description    Optional[string]                     `json:"description,omitzero"`
	Parameters     Optional[map[string]json.RawMessage] `json:"parameters,omitzero"`
	TimeoutSeconds Optional[int64]                      `json:"timeout_seconds,omitzero"`
}

type Chat struct {
	ID          string                     `json:"id"`
	WorkspaceID string                     `json:"workspace_id"`
	UserID      string                     `json:"user_id"`
	TeamID      string                     `json:"team_id"`
	Title       *string                    `json:"title"`
	CreatedAt   string                     `json:"created_at"`
	UpdatedAt   string                     `json:"updated_at"`
	Extra       map[string]json.RawMessage `json:"-"`
}

func (v *Chat) UnmarshalJSON(b []byte) error {
	type plain Chat
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"id", "workspace_id", "user_id", "team_id", "title", "created_at", "updated_at"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = Chat(out)
	return nil
}
func (v Chat) MarshalJSON() ([]byte, error) {
	type plain Chat
	return marshalUnknown(plain(v), v.Extra)
}

type ChatResource struct {
	Data  Chat                       `json:"data"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *ChatResource) UnmarshalJSON(b []byte) error {
	type plain ChatResource
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ChatResource(out)
	return nil
}
func (v ChatResource) MarshalJSON() ([]byte, error) {
	type plain ChatResource
	return marshalUnknown(plain(v), v.Extra)
}

type ChatCollection struct {
	Data  []Chat                     `json:"data"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *ChatCollection) UnmarshalJSON(b []byte) error {
	type plain ChatCollection
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ChatCollection(out)
	return nil
}
func (v ChatCollection) MarshalJSON() ([]byte, error) {
	type plain ChatCollection
	return marshalUnknown(plain(v), v.Extra)
}

type ChatCreateRequest struct {
	TeamID string                     `json:"team_id"`
	Extra  map[string]json.RawMessage `json:"-"`
}

func (v *ChatCreateRequest) UnmarshalJSON(b []byte) error {
	type plain ChatCreateRequest
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"team_id"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ChatCreateRequest(out)
	return nil
}
func (v ChatCreateRequest) MarshalJSON() ([]byte, error) {
	type plain ChatCreateRequest
	return marshalUnknown(plain(v), v.Extra)
}

type Completion struct {
	ID                string                               `json:"id"`
	WorkspaceID       string                               `json:"workspace_id"`
	ChatID            string                               `json:"chat_id"`
	TeamID            string                               `json:"team_id"`
	UserID            string                               `json:"user_id"`
	ModelOrchestrator Optional[string]                     `json:"model_orchestrator,omitzero"`
	ModelsSubagents   []string                             `json:"models_subagents"`
	StartedAt         string                               `json:"started_at"`
	FinishedAt        Optional[string]                     `json:"finished_at,omitzero"`
	DurationMs        Optional[int64]                      `json:"duration_ms,omitzero"`
	Usage             CompletionUsage                      `json:"usage"`
	Billing           Optional[BillingState]               `json:"billing,omitzero"`
	UsageCost         Optional[UsageCost]                  `json:"usage_cost,omitzero"`
	TerminalSnapshot  Optional[CompletionTerminalSnapshot] `json:"terminal_snapshot,omitzero"`
	CostUSD           *float64                             `json:"cost_usd"`
	Status            string                               `json:"status"`
	FinishReason      Optional[string]                     `json:"finish_reason,omitzero"`
	ContentHash       Optional[string]                     `json:"content_hash,omitzero"`
	CreatedAt         string                               `json:"created_at"`
	UpdatedAt         string                               `json:"updated_at"`
	Extra             map[string]json.RawMessage           `json:"-"`
}

func (v *Completion) UnmarshalJSON(b []byte) error {
	type plain Completion
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"id", "workspace_id", "chat_id", "team_id", "user_id", "model_orchestrator", "models_subagents", "started_at", "finished_at", "duration_ms", "usage", "billing", "usage_cost", "terminal_snapshot", "cost_usd", "status", "finish_reason", "content_hash", "created_at", "updated_at"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = Completion(out)
	return nil
}
func (v Completion) MarshalJSON() ([]byte, error) {
	type plain Completion
	return marshalUnknown(plain(v), v.Extra)
}

type CompletionResource struct {
	Data  Completion                 `json:"data"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *CompletionResource) UnmarshalJSON(b []byte) error {
	type plain CompletionResource
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = CompletionResource(out)
	return nil
}
func (v CompletionResource) MarshalJSON() ([]byte, error) {
	type plain CompletionResource
	return marshalUnknown(plain(v), v.Extra)
}

type CompletionCollection struct {
	Data  []Completion               `json:"data"`
	Links CompletionCollectionLinks  `json:"links"`
	Meta  CompletionCollectionMeta   `json:"meta"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *CompletionCollection) UnmarshalJSON(b []byte) error {
	type plain CompletionCollection
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data", "links", "meta"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = CompletionCollection(out)
	return nil
}
func (v CompletionCollection) MarshalJSON() ([]byte, error) {
	type plain CompletionCollection
	return marshalUnknown(plain(v), v.Extra)
}

type Agent struct {
	ID                   string                     `json:"id"`
	UserID               string                     `json:"user_id"`
	Name                 string                     `json:"name"`
	Description          *string                    `json:"description"`
	SystemPrompt         string                     `json:"system_prompt"`
	Model                *string                    `json:"model"`
	Mode                 string                     `json:"mode"`
	ToolsEnabled         []AgentTool                `json:"tools_enabled"`
	ExternalToolsEnabled []AgentTool                `json:"external_tools_enabled"`
	Permissions          map[string]json.RawMessage `json:"permissions"`
	CanWriteKnowledge    bool                       `json:"can_write_knowledge"`
	CanWriteSkills       bool                       `json:"can_write_skills"`
	Config               map[string]json.RawMessage `json:"config"`
	CreatedAt            string                     `json:"created_at"`
	UpdatedAt            string                     `json:"updated_at"`
	Extra                map[string]json.RawMessage `json:"-"`
}

func (v *Agent) UnmarshalJSON(b []byte) error {
	type plain Agent
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"id", "user_id", "name", "description", "system_prompt", "model", "mode", "tools_enabled", "external_tools_enabled", "permissions", "can_write_knowledge", "can_write_skills", "config", "created_at", "updated_at"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = Agent(out)
	return nil
}
func (v Agent) MarshalJSON() ([]byte, error) {
	type plain Agent
	return marshalUnknown(plain(v), v.Extra)
}

type AgentResource struct {
	Data  Agent                      `json:"data"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *AgentResource) UnmarshalJSON(b []byte) error {
	type plain AgentResource
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = AgentResource(out)
	return nil
}
func (v AgentResource) MarshalJSON() ([]byte, error) {
	type plain AgentResource
	return marshalUnknown(plain(v), v.Extra)
}

type AgentCollection struct {
	Data  []Agent                    `json:"data"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *AgentCollection) UnmarshalJSON(b []byte) error {
	type plain AgentCollection
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = AgentCollection(out)
	return nil
}
func (v AgentCollection) MarshalJSON() ([]byte, error) {
	type plain AgentCollection
	return marshalUnknown(plain(v), v.Extra)
}

type AgentCreateRequest struct {
	Name                 string                               `json:"name"`
	Description          Optional[string]                     `json:"description,omitzero"`
	SystemPrompt         string                               `json:"system_prompt"`
	Model                Optional[string]                     `json:"model,omitzero"`
	Mode                 string                               `json:"mode"`
	ToolsEnabled         Optional[[]AgentTool]                `json:"tools_enabled,omitzero"`
	ExternalToolsEnabled Optional[[]AgentTool]                `json:"external_tools_enabled,omitzero"`
	Permissions          Optional[map[string]json.RawMessage] `json:"permissions,omitzero"`
	CanWriteKnowledge    Optional[bool]                       `json:"can_write_knowledge,omitzero"`
	CanWriteSkills       Optional[bool]                       `json:"can_write_skills,omitzero"`
	Config               Optional[map[string]json.RawMessage] `json:"config,omitzero"`
	Extra                map[string]json.RawMessage           `json:"-"`
}

func (v *AgentCreateRequest) UnmarshalJSON(b []byte) error {
	type plain AgentCreateRequest
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"name", "description", "system_prompt", "model", "mode", "tools_enabled", "external_tools_enabled", "permissions", "can_write_knowledge", "can_write_skills", "config"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = AgentCreateRequest(out)
	return nil
}
func (v AgentCreateRequest) MarshalJSON() ([]byte, error) {
	type plain AgentCreateRequest
	return marshalUnknown(plain(v), v.Extra)
}

type AgentUpdateRequest struct {
	Name                 Optional[string]                     `json:"name,omitzero"`
	Description          Optional[string]                     `json:"description,omitzero"`
	SystemPrompt         Optional[string]                     `json:"system_prompt,omitzero"`
	Model                Optional[string]                     `json:"model,omitzero"`
	Mode                 Optional[string]                     `json:"mode,omitzero"`
	ToolsEnabled         Optional[[]AgentTool]                `json:"tools_enabled,omitzero"`
	ExternalToolsEnabled Optional[[]AgentTool]                `json:"external_tools_enabled,omitzero"`
	Permissions          Optional[map[string]json.RawMessage] `json:"permissions,omitzero"`
	CanWriteKnowledge    Optional[bool]                       `json:"can_write_knowledge,omitzero"`
	CanWriteSkills       Optional[bool]                       `json:"can_write_skills,omitzero"`
	Config               Optional[map[string]json.RawMessage] `json:"config,omitzero"`
	Extra                map[string]json.RawMessage           `json:"-"`
}

func (v *AgentUpdateRequest) UnmarshalJSON(b []byte) error {
	type plain AgentUpdateRequest
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"name", "description", "system_prompt", "model", "mode", "tools_enabled", "external_tools_enabled", "permissions", "can_write_knowledge", "can_write_skills", "config"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = AgentUpdateRequest(out)
	return nil
}
func (v AgentUpdateRequest) MarshalJSON() ([]byte, error) {
	type plain AgentUpdateRequest
	return marshalUnknown(plain(v), v.Extra)
}

type AgentTool map[string]json.RawMessage

type Team struct {
	ID             string                     `json:"id"`
	UserID         string                     `json:"user_id"`
	Name           string                     `json:"name"`
	Description    *string                    `json:"description"`
	OrchestratorID string                     `json:"orchestrator_id"`
	Members        []TeamMember               `json:"members"`
	CreatedAt      string                     `json:"created_at"`
	UpdatedAt      string                     `json:"updated_at"`
	Extra          map[string]json.RawMessage `json:"-"`
}

func (v *Team) UnmarshalJSON(b []byte) error {
	type plain Team
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"id", "user_id", "name", "description", "orchestrator_id", "members", "created_at", "updated_at"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = Team(out)
	return nil
}
func (v Team) MarshalJSON() ([]byte, error) {
	type plain Team
	return marshalUnknown(plain(v), v.Extra)
}

type TeamResource struct {
	Data  Team                       `json:"data"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *TeamResource) UnmarshalJSON(b []byte) error {
	type plain TeamResource
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = TeamResource(out)
	return nil
}
func (v TeamResource) MarshalJSON() ([]byte, error) {
	type plain TeamResource
	return marshalUnknown(plain(v), v.Extra)
}

type TeamCollection struct {
	Data  []Team                     `json:"data"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *TeamCollection) UnmarshalJSON(b []byte) error {
	type plain TeamCollection
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = TeamCollection(out)
	return nil
}
func (v TeamCollection) MarshalJSON() ([]byte, error) {
	type plain TeamCollection
	return marshalUnknown(plain(v), v.Extra)
}

type TeamCreateRequest struct {
	Name           string                              `json:"name"`
	Description    Optional[string]                    `json:"description,omitzero"`
	OrchestratorID string                              `json:"orchestrator_id"`
	Members        Optional[[]TeamMemberCreateRequest] `json:"members,omitzero"`
	Extra          map[string]json.RawMessage          `json:"-"`
}

func (v *TeamCreateRequest) UnmarshalJSON(b []byte) error {
	type plain TeamCreateRequest
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"name", "description", "orchestrator_id", "members"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = TeamCreateRequest(out)
	return nil
}
func (v TeamCreateRequest) MarshalJSON() ([]byte, error) {
	type plain TeamCreateRequest
	return marshalUnknown(plain(v), v.Extra)
}

type TeamUpdateRequest struct {
	Name           Optional[string]                    `json:"name,omitzero"`
	Description    Optional[string]                    `json:"description,omitzero"`
	OrchestratorID Optional[string]                    `json:"orchestrator_id,omitzero"`
	Members        Optional[[]TeamMemberCreateRequest] `json:"members,omitzero"`
	Extra          map[string]json.RawMessage          `json:"-"`
}

func (v *TeamUpdateRequest) UnmarshalJSON(b []byte) error {
	type plain TeamUpdateRequest
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"name", "description", "orchestrator_id", "members"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = TeamUpdateRequest(out)
	return nil
}
func (v TeamUpdateRequest) MarshalJSON() ([]byte, error) {
	type plain TeamUpdateRequest
	return marshalUnknown(plain(v), v.Extra)
}

type TeamMember struct {
	AgentDefinitionID string                     `json:"agent_definition_id"`
	Name              string                     `json:"name"`
	Position          int64                      `json:"position"`
	Role              *string                    `json:"role"`
	Extra             map[string]json.RawMessage `json:"-"`
}

func (v *TeamMember) UnmarshalJSON(b []byte) error {
	type plain TeamMember
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"agent_definition_id", "name", "position", "role"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = TeamMember(out)
	return nil
}
func (v TeamMember) MarshalJSON() ([]byte, error) {
	type plain TeamMember
	return marshalUnknown(plain(v), v.Extra)
}

type TeamMemberCreateRequest struct {
	AgentDefinitionID string                     `json:"agent_definition_id"`
	Role              Optional[string]           `json:"role,omitzero"`
	Extra             map[string]json.RawMessage `json:"-"`
}

func (v *TeamMemberCreateRequest) UnmarshalJSON(b []byte) error {
	type plain TeamMemberCreateRequest
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"agent_definition_id", "role"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = TeamMemberCreateRequest(out)
	return nil
}
func (v TeamMemberCreateRequest) MarshalJSON() ([]byte, error) {
	type plain TeamMemberCreateRequest
	return marshalUnknown(plain(v), v.Extra)
}

type APIKey struct {
	ID        string                     `json:"id"`
	Name      string                     `json:"name"`
	Scopes    []string                   `json:"scopes"`
	ExpiresAt *string                    `json:"expires_at"`
	RevokedAt *string                    `json:"revoked_at"`
	CreatedAt string                     `json:"created_at"`
	Extra     map[string]json.RawMessage `json:"-"`
}

func (v *APIKey) UnmarshalJSON(b []byte) error {
	type plain APIKey
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"id", "name", "scopes", "expires_at", "revoked_at", "created_at"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = APIKey(out)
	return nil
}
func (v APIKey) MarshalJSON() ([]byte, error) {
	type plain APIKey
	return marshalUnknown(plain(v), v.Extra)
}

type APIKeyCreated struct {
	Data  json.RawMessage            `json:"data"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *APIKeyCreated) UnmarshalJSON(b []byte) error {
	type plain APIKeyCreated
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = APIKeyCreated(out)
	return nil
}
func (v APIKeyCreated) MarshalJSON() ([]byte, error) {
	type plain APIKeyCreated
	return marshalUnknown(plain(v), v.Extra)
}

type APIKeyCollection struct {
	Data  []APIKey                   `json:"data"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *APIKeyCollection) UnmarshalJSON(b []byte) error {
	type plain APIKeyCollection
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = APIKeyCollection(out)
	return nil
}
func (v APIKeyCollection) MarshalJSON() ([]byte, error) {
	type plain APIKeyCollection
	return marshalUnknown(plain(v), v.Extra)
}

type APIKeyCreate struct {
	Name      string                     `json:"name"`
	Scopes    Optional[[]string]         `json:"scopes,omitzero"`
	ExpiresAt Optional[string]           `json:"expires_at,omitzero"`
	Extra     map[string]json.RawMessage `json:"-"`
}

func (v *APIKeyCreate) UnmarshalJSON(b []byte) error {
	type plain APIKeyCreate
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"name", "scopes", "expires_at"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = APIKeyCreate(out)
	return nil
}
func (v APIKeyCreate) MarshalJSON() ([]byte, error) {
	type plain APIKeyCreate
	return marshalUnknown(plain(v), v.Extra)
}

type JWKSDocument struct {
	Keys  []JWK                      `json:"keys"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *JWKSDocument) UnmarshalJSON(b []byte) error {
	type plain JWKSDocument
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"keys"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = JWKSDocument(out)
	return nil
}
func (v JWKSDocument) MarshalJSON() ([]byte, error) {
	type plain JWKSDocument
	return marshalUnknown(plain(v), v.Extra)
}

type JWK struct {
	Kty   string                     `json:"kty"`
	Crv   string                     `json:"crv"`
	X     string                     `json:"x"`
	Kid   string                     `json:"kid"`
	Use   string                     `json:"use"`
	Alg   string                     `json:"alg"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *JWK) UnmarshalJSON(b []byte) error {
	type plain JWK
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"kty", "crv", "x", "kid", "use", "alg"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = JWK(out)
	return nil
}
func (v JWK) MarshalJSON() ([]byte, error) { type plain JWK; return marshalUnknown(plain(v), v.Extra) }

type ModelList struct {
	Object   string                     `json:"object"`
	Data     []ModelListEntry           `json:"data"`
	Defaults ModelDefaults              `json:"defaults"`
	Extra    map[string]json.RawMessage `json:"-"`
}

func (v *ModelList) UnmarshalJSON(b []byte) error {
	type plain ModelList
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"object", "data", "defaults"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ModelList(out)
	return nil
}
func (v ModelList) MarshalJSON() ([]byte, error) {
	type plain ModelList
	return marshalUnknown(plain(v), v.Extra)
}

type ModelListEntry map[string]json.RawMessage

type LLMModelEntry struct {
	ID                     string                               `json:"id"`
	Object                 string                               `json:"object"`
	Created                int64                                `json:"created"`
	OwnedBy                string                               `json:"owned_by"`
	Type                   string                               `json:"type"`
	Deprecated             bool                                 `json:"deprecated"`
	SunsetAt               *string                              `json:"sunset_at"`
	DisplayName            Optional[string]                     `json:"display_name,omitzero"`
	Description            Optional[string]                     `json:"description,omitzero"`
	CanonicalSlug          Optional[string]                     `json:"canonical_slug,omitzero"`
	HuggingFaceID          Optional[string]                     `json:"hugging_face_id,omitzero"`
	ContextLength          Optional[int64]                      `json:"context_length,omitzero"`
	MaxCompletionTokens    Optional[int64]                      `json:"max_completion_tokens,omitzero"`
	Pricing                Optional[OpenRouterPricing]          `json:"pricing,omitzero"`
	Architecture           Optional[OpenRouterArchitecture]     `json:"architecture,omitzero"`
	SupportedParameters    Optional[[]string]                   `json:"supported_parameters,omitzero"`
	DefaultParameters      Optional[map[string]json.RawMessage] `json:"default_parameters,omitzero"`
	Reasoning              Optional[map[string]json.RawMessage] `json:"reasoning,omitzero"`
	KnowledgeCutoff        Optional[string]                     `json:"knowledge_cutoff,omitzero"`
	UpstreamExpirationDate Optional[string]                     `json:"upstream_expiration_date,omitzero"`
	AliasTarget            Optional[OpenRouterAliasTarget]      `json:"alias_target,omitzero"`
	IsModerated            Optional[bool]                       `json:"is_moderated,omitzero"`
	SyncedAt               Optional[string]                     `json:"synced_at,omitzero"`
	Extra                  map[string]json.RawMessage           `json:"-"`
}

func (v *LLMModelEntry) UnmarshalJSON(b []byte) error {
	type plain LLMModelEntry
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"id", "object", "created", "owned_by", "type", "deprecated", "sunset_at", "display_name", "description", "canonical_slug", "hugging_face_id", "context_length", "max_completion_tokens", "pricing", "architecture", "supported_parameters", "default_parameters", "reasoning", "knowledge_cutoff", "upstream_expiration_date", "alias_target", "is_moderated", "synced_at"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = LLMModelEntry(out)
	return nil
}
func (v LLMModelEntry) MarshalJSON() ([]byte, error) {
	type plain LLMModelEntry
	return marshalUnknown(plain(v), v.Extra)
}

type OpenRouterPricing map[string]json.RawMessage

type OpenRouterArchitecture struct {
	Modality         Optional[string]           `json:"modality,omitzero"`
	InputModalities  Optional[[]string]         `json:"input_modalities,omitzero"`
	OutputModalities Optional[[]string]         `json:"output_modalities,omitzero"`
	Tokenizer        Optional[string]           `json:"tokenizer,omitzero"`
	InstructType     Optional[string]           `json:"instruct_type,omitzero"`
	Extra            map[string]json.RawMessage `json:"-"`
}

func (v *OpenRouterArchitecture) UnmarshalJSON(b []byte) error {
	type plain OpenRouterArchitecture
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"modality", "input_modalities", "output_modalities", "tokenizer", "instruct_type"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = OpenRouterArchitecture(out)
	return nil
}
func (v OpenRouterArchitecture) MarshalJSON() ([]byte, error) {
	type plain OpenRouterArchitecture
	return marshalUnknown(plain(v), v.Extra)
}

type OpenRouterAliasTarget struct {
	Name  Optional[string]           `json:"name,omitzero"`
	Slug  Optional[string]           `json:"slug,omitzero"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *OpenRouterAliasTarget) UnmarshalJSON(b []byte) error {
	type plain OpenRouterAliasTarget
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"name", "slug"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = OpenRouterAliasTarget(out)
	return nil
}
func (v OpenRouterAliasTarget) MarshalJSON() ([]byte, error) {
	type plain OpenRouterAliasTarget
	return marshalUnknown(plain(v), v.Extra)
}

type TeamModelEntry struct {
	ID         string                     `json:"id"`
	Object     string                     `json:"object"`
	Type       string                     `json:"type"`
	OwnedBy    string                     `json:"owned_by"`
	TeamID     string                     `json:"team_id"`
	Created    Optional[int64]            `json:"created,omitzero"`
	Deprecated bool                       `json:"deprecated"`
	SunsetAt   *string                    `json:"sunset_at"`
	Extra      map[string]json.RawMessage `json:"-"`
}

func (v *TeamModelEntry) UnmarshalJSON(b []byte) error {
	type plain TeamModelEntry
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"id", "object", "type", "owned_by", "team_id", "created", "deprecated", "sunset_at"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = TeamModelEntry(out)
	return nil
}
func (v TeamModelEntry) MarshalJSON() ([]byte, error) {
	type plain TeamModelEntry
	return marshalUnknown(plain(v), v.Extra)
}

type ModelDefaults struct {
	Orchestrator string                     `json:"orchestrator"`
	Subagent     string                     `json:"subagent"`
	Extra        map[string]json.RawMessage `json:"-"`
}

func (v *ModelDefaults) UnmarshalJSON(b []byte) error {
	type plain ModelDefaults
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"orchestrator", "subagent"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ModelDefaults(out)
	return nil
}
func (v ModelDefaults) MarshalJSON() ([]byte, error) {
	type plain ModelDefaults
	return marshalUnknown(plain(v), v.Extra)
}

type TestImageGenerateRequest struct {
	Files     []string                   `json:"files"`
	Prompt    string                     `json:"prompt"`
	ModelCode string                     `json:"model_code"`
	Extra     map[string]json.RawMessage `json:"-"`
}

func (v *TestImageGenerateRequest) UnmarshalJSON(b []byte) error {
	type plain TestImageGenerateRequest
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"files", "prompt", "model_code"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = TestImageGenerateRequest(out)
	return nil
}
func (v TestImageGenerateRequest) MarshalJSON() ([]byte, error) {
	type plain TestImageGenerateRequest
	return marshalUnknown(plain(v), v.Extra)
}

type TestImageFileMetadata struct {
	OriginalName string                     `json:"original_name"`
	MimeType     string                     `json:"mime_type"`
	Size         int64                      `json:"size"`
	Extra        map[string]json.RawMessage `json:"-"`
}

func (v *TestImageFileMetadata) UnmarshalJSON(b []byte) error {
	type plain TestImageFileMetadata
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"original_name", "mime_type", "size"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = TestImageFileMetadata(out)
	return nil
}
func (v TestImageFileMetadata) MarshalJSON() ([]byte, error) {
	type plain TestImageFileMetadata
	return marshalUnknown(plain(v), v.Extra)
}

type TestImageGenerateResource struct {
	Data  TestImageGenerateResourceData `json:"data"`
	Extra map[string]json.RawMessage    `json:"-"`
}

func (v *TestImageGenerateResource) UnmarshalJSON(b []byte) error {
	type plain TestImageGenerateResource
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = TestImageGenerateResource(out)
	return nil
}
func (v TestImageGenerateResource) MarshalJSON() ([]byte, error) {
	type plain TestImageGenerateResource
	return marshalUnknown(plain(v), v.Extra)
}

type InputImage struct {
	Type      string           `json:"type"`
	Payload   string           `json:"payload"`
	Reference Optional[string] `json:"reference,omitzero"`
	MimeType  Optional[string] `json:"mime_type,omitzero"`
}

type ImageDimensions struct {
	Width  int64 `json:"width"`
	Height int64 `json:"height"`
}

type ImageGenerationRequest struct {
	Model         string                               `json:"model"`
	Prompt        string                               `json:"prompt"`
	WebhookURL    string                               `json:"webhook_url"`
	InputImages   Optional[[]InputImage]               `json:"input_images,omitzero"`
	InputImage    Optional[InputImage]                 `json:"input_image,omitzero"`
	MaskImage     Optional[InputImage]                 `json:"mask_image,omitzero"`
	ObjectsImage  Optional[InputImage]                 `json:"objects_image,omitzero"`
	Resolution    Optional[string]                     `json:"resolution,omitzero"`
	Dimensions    Optional[ImageDimensions]            `json:"dimensions,omitzero"`
	AspectRatio   Optional[string]                     `json:"aspect_ratio,omitzero"`
	Format        Optional[string]                     `json:"format,omitzero"`
	Quality       Optional[int64]                      `json:"quality,omitzero"`
	ModelSettings Optional[map[string]json.RawMessage] `json:"model_settings,omitzero"`
	Providers     Optional[[]string]                   `json:"providers,omitzero"`
}

type ImageGenerationAcceptedRecord struct {
	ID              string                         `json:"id"`
	Status          ImageGenerationLifecycleStatus `json:"status"`
	Provider        string                         `json:"provider"`
	InitialProvider string                         `json:"initial_provider"`
	Routed          bool                           `json:"routed"`
	Model           string                         `json:"model"`
	WebhookURL      string                         `json:"webhook_url"`
	StatusURL       string                         `json:"status_url"`
	CreatedAt       string                         `json:"created_at"`
	ExpiresAt       string                         `json:"expires_at"`
	Extra           map[string]json.RawMessage     `json:"-"`
}

func (v *ImageGenerationAcceptedRecord) UnmarshalJSON(b []byte) error {
	type plain ImageGenerationAcceptedRecord
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"id", "status", "provider", "initial_provider", "routed", "model", "webhook_url", "status_url", "created_at", "expires_at"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ImageGenerationAcceptedRecord(out)
	return nil
}
func (v ImageGenerationAcceptedRecord) MarshalJSON() ([]byte, error) {
	type plain ImageGenerationAcceptedRecord
	return marshalUnknown(plain(v), v.Extra)
}

type ImageGenerationLifecycleStatus string

type ImageGenerationAcceptedResource struct {
	Data  ImageGenerationAcceptedRecord `json:"data"`
	Extra map[string]json.RawMessage    `json:"-"`
}

func (v *ImageGenerationAcceptedResource) UnmarshalJSON(b []byte) error {
	type plain ImageGenerationAcceptedResource
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ImageGenerationAcceptedResource(out)
	return nil
}
func (v ImageGenerationAcceptedResource) MarshalJSON() ([]byte, error) {
	type plain ImageGenerationAcceptedResource
	return marshalUnknown(plain(v), v.Extra)
}

type ImageGenerationStatusRecord struct {
	ID              string                                             `json:"id"`
	Status          ImageGenerationLifecycleStatus                     `json:"status"`
	Provider        string                                             `json:"provider"`
	InitialProvider string                                             `json:"initial_provider"`
	Routed          bool                                               `json:"routed"`
	Model           string                                             `json:"model"`
	CreatedAt       string                                             `json:"created_at"`
	SubmittedAt     *string                                            `json:"submitted_at"`
	CompletedAt     *string                                            `json:"completed_at"`
	ExpiresAt       string                                             `json:"expires_at"`
	Result          Optional[ImageGenerationStatusRecordResult]        `json:"result,omitzero"`
	ErrorEnvelope   Optional[ImageGenerationStatusRecordErrorEnvelope] `json:"error,omitzero"`
	Files           Optional[[]ImageGenerationStatusRecordFilesItem]   `json:"files,omitzero"`
	Price           Optional[ImagePrice]                               `json:"price,omitzero"`
	Usage           Optional[ImageUsage]                               `json:"usage,omitzero"`
	CostUSD         Optional[float64]                                  `json:"cost_usd,omitzero"`
	DurationMs      Optional[int64]                                    `json:"duration_ms,omitzero"`
	Extra           map[string]json.RawMessage                         `json:"-"`
}

func (v *ImageGenerationStatusRecord) UnmarshalJSON(b []byte) error {
	type plain ImageGenerationStatusRecord
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"id", "status", "provider", "initial_provider", "routed", "model", "created_at", "submitted_at", "completed_at", "expires_at", "result", "error", "files", "price", "usage", "cost_usd", "duration_ms"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ImageGenerationStatusRecord(out)
	return nil
}
func (v ImageGenerationStatusRecord) MarshalJSON() ([]byte, error) {
	type plain ImageGenerationStatusRecord
	return marshalUnknown(plain(v), v.Extra)
}

type ImageGenerationStatusResource struct {
	Data  ImageGenerationStatusRecord `json:"data"`
	Extra map[string]json.RawMessage  `json:"-"`
}

func (v *ImageGenerationStatusResource) UnmarshalJSON(b []byte) error {
	type plain ImageGenerationStatusResource
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ImageGenerationStatusResource(out)
	return nil
}
func (v ImageGenerationStatusResource) MarshalJSON() ([]byte, error) {
	type plain ImageGenerationStatusResource
	return marshalUnknown(plain(v), v.Extra)
}

type ImagePrice struct {
	Amount   string                     `json:"amount"`
	Currency string                     `json:"currency"`
	Source   string                     `json:"source"`
	Extra    map[string]json.RawMessage `json:"-"`
}

func (v *ImagePrice) UnmarshalJSON(b []byte) error {
	type plain ImagePrice
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"amount", "currency", "source"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ImagePrice(out)
	return nil
}
func (v ImagePrice) MarshalJSON() ([]byte, error) {
	type plain ImagePrice
	return marshalUnknown(plain(v), v.Extra)
}

type ImageUsage struct {
	SchemaVersion  int64                      `json:"schema_version"`
	AttemptsTotal  int64                      `json:"attempts_total"`
	AttemptsPriced int64                      `json:"attempts_priced"`
	Extra          map[string]json.RawMessage `json:"-"`
}

func (v *ImageUsage) UnmarshalJSON(b []byte) error {
	type plain ImageUsage
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"schema_version", "attempts_total", "attempts_priced"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ImageUsage(out)
	return nil
}
func (v ImageUsage) MarshalJSON() ([]byte, error) {
	type plain ImageUsage
	return marshalUnknown(plain(v), v.Extra)
}

type ImageUploadRecord struct {
	ID               string                     `json:"id"`
	MimeType         string                     `json:"mime_type"`
	Size             int64                      `json:"size"`
	OriginalFilename string                     `json:"original_filename"`
	ExpiresAt        string                     `json:"expires_at"`
	CreatedAt        string                     `json:"created_at"`
	Extra            map[string]json.RawMessage `json:"-"`
}

func (v *ImageUploadRecord) UnmarshalJSON(b []byte) error {
	type plain ImageUploadRecord
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"id", "mime_type", "size", "original_filename", "expires_at", "created_at"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ImageUploadRecord(out)
	return nil
}
func (v ImageUploadRecord) MarshalJSON() ([]byte, error) {
	type plain ImageUploadRecord
	return marshalUnknown(plain(v), v.Extra)
}

type ImageUploadResource struct {
	Data  ImageUploadRecord          `json:"data"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *ImageUploadResource) UnmarshalJSON(b []byte) error {
	type plain ImageUploadResource
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ImageUploadResource(out)
	return nil
}
func (v ImageUploadResource) MarshalJSON() ([]byte, error) {
	type plain ImageUploadResource
	return marshalUnknown(plain(v), v.Extra)
}

type ImageUploadRequest struct {
	File  string                     `json:"file"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *ImageUploadRequest) UnmarshalJSON(b []byte) error {
	type plain ImageUploadRequest
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"file"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ImageUploadRequest(out)
	return nil
}
func (v ImageUploadRequest) MarshalJSON() ([]byte, error) {
	type plain ImageUploadRequest
	return marshalUnknown(plain(v), v.Extra)
}

type ModelCatalogEntry struct {
	Model         string                     `json:"model"`
	Providers     []string                   `json:"providers"`
	PriceClasses  []ModelPriceClass          `json:"price_classes"`
	ModelSettings []ModelSetting             `json:"model_settings"`
	Extra         map[string]json.RawMessage `json:"-"`
}

func (v *ModelCatalogEntry) UnmarshalJSON(b []byte) error {
	type plain ModelCatalogEntry
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"model", "providers", "price_classes", "model_settings"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ModelCatalogEntry(out)
	return nil
}
func (v ModelCatalogEntry) MarshalJSON() ([]byte, error) {
	type plain ModelCatalogEntry
	return marshalUnknown(plain(v), v.Extra)
}

type ModelPriceClass struct {
	Key            string                      `json:"key"`
	Label          string                      `json:"label"`
	ProvidersCount int64                       `json:"providers_count"`
	AveragePrice   ModelPriceClassAveragePrice `json:"average_price"`
	Extra          map[string]json.RawMessage  `json:"-"`
}

func (v *ModelPriceClass) UnmarshalJSON(b []byte) error {
	type plain ModelPriceClass
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"key", "label", "providers_count", "average_price"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ModelPriceClass(out)
	return nil
}
func (v ModelPriceClass) MarshalJSON() ([]byte, error) {
	type plain ModelPriceClass
	return marshalUnknown(plain(v), v.Extra)
}

type ModelSetting struct {
	Key                string                               `json:"key"`
	Type               string                               `json:"type"`
	Providers          []string                             `json:"providers"`
	Default            Optional[json.RawMessage]            `json:"default,omitzero"`
	Enum               Optional[[]string]                   `json:"enum,omitzero"`
	Range              Optional[[]int64]                    `json:"range,omitzero"`
	DefaultsByProvider Optional[map[string]json.RawMessage] `json:"defaults_by_provider,omitzero"`
	Extra              map[string]json.RawMessage           `json:"-"`
}

func (v *ModelSetting) UnmarshalJSON(b []byte) error {
	type plain ModelSetting
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"key", "type", "providers", "default", "enum", "range", "defaults_by_provider"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ModelSetting(out)
	return nil
}
func (v ModelSetting) MarshalJSON() ([]byte, error) {
	type plain ModelSetting
	return marshalUnknown(plain(v), v.Extra)
}

type ModelCatalogResource struct {
	Data  []ModelCatalogEntry        `json:"data"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *ModelCatalogResource) UnmarshalJSON(b []byte) error {
	type plain ModelCatalogResource
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ModelCatalogResource(out)
	return nil
}
func (v ModelCatalogResource) MarshalJSON() ([]byte, error) {
	type plain ModelCatalogResource
	return marshalUnknown(plain(v), v.Extra)
}

type ProviderWebhookPayload map[string]json.RawMessage

type WebhookOk struct {
	Status string                     `json:"status"`
	Extra  map[string]json.RawMessage `json:"-"`
}

func (v *WebhookOk) UnmarshalJSON(b []byte) error {
	type plain WebhookOk
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"status"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = WebhookOk(out)
	return nil
}
func (v WebhookOk) MarshalJSON() ([]byte, error) {
	type plain WebhookOk
	return marshalUnknown(plain(v), v.Extra)
}

type WorkspaceFile struct {
	Path       string                         `json:"path"`
	Name       string                         `json:"name"`
	Type       string                         `json:"type"`
	Size       int64                          `json:"size"`
	ModifiedAt *string                        `json:"modified_at"`
	MimeType   *string                        `json:"mime_type"`
	Content    Optional[string]               `json:"content,omitzero"`
	Entries    Optional[[]WorkspaceFileEntry] `json:"entries,omitzero"`
	Extra      map[string]json.RawMessage     `json:"-"`
}

func (v *WorkspaceFile) UnmarshalJSON(b []byte) error {
	type plain WorkspaceFile
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"path", "name", "type", "size", "modified_at", "mime_type", "content", "entries"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = WorkspaceFile(out)
	return nil
}
func (v WorkspaceFile) MarshalJSON() ([]byte, error) {
	type plain WorkspaceFile
	return marshalUnknown(plain(v), v.Extra)
}

type WorkspaceFileEntry struct {
	Path       string                     `json:"path"`
	Name       string                     `json:"name"`
	Type       string                     `json:"type"`
	Size       int64                      `json:"size"`
	ModifiedAt *string                    `json:"modified_at"`
	MimeType   *string                    `json:"mime_type"`
	Extra      map[string]json.RawMessage `json:"-"`
}

func (v *WorkspaceFileEntry) UnmarshalJSON(b []byte) error {
	type plain WorkspaceFileEntry
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"path", "name", "type", "size", "modified_at", "mime_type"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = WorkspaceFileEntry(out)
	return nil
}
func (v WorkspaceFileEntry) MarshalJSON() ([]byte, error) {
	type plain WorkspaceFileEntry
	return marshalUnknown(plain(v), v.Extra)
}

type WorkspaceFileResource struct {
	Data  WorkspaceFile              `json:"data"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *WorkspaceFileResource) UnmarshalJSON(b []byte) error {
	type plain WorkspaceFileResource
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = WorkspaceFileResource(out)
	return nil
}
func (v WorkspaceFileResource) MarshalJSON() ([]byte, error) {
	type plain WorkspaceFileResource
	return marshalUnknown(plain(v), v.Extra)
}

type WorkspaceFileRequest struct {
	Path    string                     `json:"path"`
	Content string                     `json:"content"`
	Extra   map[string]json.RawMessage `json:"-"`
}

func (v *WorkspaceFileRequest) UnmarshalJSON(b []byte) error {
	type plain WorkspaceFileRequest
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"path", "content"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = WorkspaceFileRequest(out)
	return nil
}
func (v WorkspaceFileRequest) MarshalJSON() ([]byte, error) {
	type plain WorkspaceFileRequest
	return marshalUnknown(plain(v), v.Extra)
}

type ConvertToPDFRequest struct {
	Path  string                     `json:"path"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *ConvertToPDFRequest) UnmarshalJSON(b []byte) error {
	type plain ConvertToPDFRequest
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"path"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ConvertToPDFRequest(out)
	return nil
}
func (v ConvertToPDFRequest) MarshalJSON() ([]byte, error) {
	type plain ConvertToPDFRequest
	return marshalUnknown(plain(v), v.Extra)
}

type ThrottleError struct {
	ErrorEnvelope ThrottleErrorErrorEnvelope `json:"error"`
	Extra         map[string]json.RawMessage `json:"-"`
}

func (v *ThrottleError) UnmarshalJSON(b []byte) error {
	type plain ThrottleError
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"error"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ThrottleError(out)
	return nil
}
func (v ThrottleError) MarshalJSON() ([]byte, error) {
	type plain ThrottleError
	return marshalUnknown(plain(v), v.Extra)
}

type CapabilitiesLegacy struct {
	ProtocolVersion int64                      `json:"protocol_version"`
	Families        []string                   `json:"families"`
	EventTypes      []string                   `json:"event_types"`
	ToolKinds       []string                   `json:"tool_kinds"`
	Features        []string                   `json:"features"`
	Extra           map[string]json.RawMessage `json:"-"`
}

func (v *CapabilitiesLegacy) UnmarshalJSON(b []byte) error {
	type plain CapabilitiesLegacy
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(b, &fields); err != nil {
		return err
	}
	features, hasFeatures := fields["features"]
	delete(fields, "features")
	body, err := json.Marshal(fields)
	if err != nil {
		return err
	}
	var out plain
	if err = json.Unmarshal(body, &out); err != nil {
		return err
	}
	if hasFeatures {
		out.Features, err = normalizeFeatureNames(features)
		if err != nil {
			return err
		}
	}
	extra, err := unknownFields(body, []string{"protocol_version", "families", "event_types", "tool_kinds", "features"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = CapabilitiesLegacy(out)
	return nil
}
func (v CapabilitiesLegacy) MarshalJSON() ([]byte, error) {
	type plain CapabilitiesLegacy
	return marshalUnknown(plain(v), v.Extra)
}

type CreateWikiSyncRunBody struct {
	WebhookURL string                     `json:"webhook_url"`
	Extra      map[string]json.RawMessage `json:"-"`
}

func (v *CreateWikiSyncRunBody) UnmarshalJSON(b []byte) error {
	type plain CreateWikiSyncRunBody
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"webhook_url"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = CreateWikiSyncRunBody(out)
	return nil
}
func (v CreateWikiSyncRunBody) MarshalJSON() ([]byte, error) {
	type plain CreateWikiSyncRunBody
	return marshalUnknown(plain(v), v.Extra)
}

type CreateWikiOutboxRunBody struct {
	WebhookURL string                     `json:"webhook_url"`
	Extra      map[string]json.RawMessage `json:"-"`
}

func (v *CreateWikiOutboxRunBody) UnmarshalJSON(b []byte) error {
	type plain CreateWikiOutboxRunBody
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"webhook_url"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = CreateWikiOutboxRunBody(out)
	return nil
}
func (v CreateWikiOutboxRunBody) MarshalJSON() ([]byte, error) {
	type plain CreateWikiOutboxRunBody
	return marshalUnknown(plain(v), v.Extra)
}

type CreateWikiInboxRunBody struct {
	WebhookURL string                     `json:"webhook_url"`
	Extra      map[string]json.RawMessage `json:"-"`
}

func (v *CreateWikiInboxRunBody) UnmarshalJSON(b []byte) error {
	type plain CreateWikiInboxRunBody
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"webhook_url"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = CreateWikiInboxRunBody(out)
	return nil
}
func (v CreateWikiInboxRunBody) MarshalJSON() ([]byte, error) {
	type plain CreateWikiInboxRunBody
	return marshalUnknown(plain(v), v.Extra)
}

type WithdrawWikiSourceBody struct {
	FileID string                     `json:"file_id"`
	Extra  map[string]json.RawMessage `json:"-"`
}

func (v *WithdrawWikiSourceBody) UnmarshalJSON(b []byte) error {
	type plain WithdrawWikiSourceBody
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"file_id"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = WithdrawWikiSourceBody(out)
	return nil
}
func (v WithdrawWikiSourceBody) MarshalJSON() ([]byte, error) {
	type plain WithdrawWikiSourceBody
	return marshalUnknown(plain(v), v.Extra)
}

type ListChatAttachmentsResult struct {
	Attachments []ChatAttachment           `json:"attachments"`
	Extra       map[string]json.RawMessage `json:"-"`
}

func (v *ListChatAttachmentsResult) UnmarshalJSON(b []byte) error {
	type plain ListChatAttachmentsResult
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"attachments"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ListChatAttachmentsResult(out)
	return nil
}
func (v ListChatAttachmentsResult) MarshalJSON() ([]byte, error) {
	type plain ListChatAttachmentsResult
	return marshalUnknown(plain(v), v.Extra)
}

type TestGenerateImageResult struct {
	Status     string                     `json:"status"`
	Prompt     string                     `json:"prompt"`
	ModelCode  string                     `json:"model_code"`
	FilesCount int64                      `json:"files_count"`
	Files      []TestImageFileMetadata    `json:"files"`
	ReceivedAt string                     `json:"received_at"`
	Extra      map[string]json.RawMessage `json:"-"`
}

func (v *TestGenerateImageResult) UnmarshalJSON(b []byte) error {
	type plain TestGenerateImageResult
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"status", "prompt", "model_code", "files_count", "files", "received_at"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = TestGenerateImageResult(out)
	return nil
}
func (v TestGenerateImageResult) MarshalJSON() ([]byte, error) {
	type plain TestGenerateImageResult
	return marshalUnknown(plain(v), v.Extra)
}

type HealthCheckResult struct {
	Status    Optional[string]           `json:"status,omitzero"`
	Timestamp Optional[string]           `json:"timestamp,omitzero"`
	App       Optional[string]           `json:"app,omitzero"`
	Env       Optional[string]           `json:"env,omitzero"`
	Extra     map[string]json.RawMessage `json:"-"`
}

func (v *HealthCheckResult) UnmarshalJSON(b []byte) error {
	type plain HealthCheckResult
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"status", "timestamp", "app", "env"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = HealthCheckResult(out)
	return nil
}
func (v HealthCheckResult) MarshalJSON() ([]byte, error) {
	type plain HealthCheckResult
	return marshalUnknown(plain(v), v.Extra)
}

type CancelDetachedTaskResult struct {
	TaskID  string `json:"task_id"`
	Status  string `json:"status"`
	Subtree bool   `json:"subtree"`
}

type CancelDetachedTaskBody struct {
	Subtree Optional[bool] `json:"subtree,omitzero"`
}

type ThrottleErrorErrorEnvelope struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type ModelPriceClassAveragePrice struct {
	Amount   *string                    `json:"amount"`
	Currency *string                    `json:"currency"`
	Extra    map[string]json.RawMessage `json:"-"`
}

func (v *ModelPriceClassAveragePrice) UnmarshalJSON(b []byte) error {
	type plain ModelPriceClassAveragePrice
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"amount", "currency"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ModelPriceClassAveragePrice(out)
	return nil
}
func (v ModelPriceClassAveragePrice) MarshalJSON() ([]byte, error) {
	type plain ModelPriceClassAveragePrice
	return marshalUnknown(plain(v), v.Extra)
}

type ImageGenerationStatusRecordFilesItem struct {
	ArtifactID  string                     `json:"artifact_id"`
	MimeType    string                     `json:"mime_type"`
	Size        int64                      `json:"size"`
	DownloadURL *string                    `json:"download_url"`
	Extra       map[string]json.RawMessage `json:"-"`
}

func (v *ImageGenerationStatusRecordFilesItem) UnmarshalJSON(b []byte) error {
	type plain ImageGenerationStatusRecordFilesItem
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"artifact_id", "mime_type", "size", "download_url"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ImageGenerationStatusRecordFilesItem(out)
	return nil
}
func (v ImageGenerationStatusRecordFilesItem) MarshalJSON() ([]byte, error) {
	type plain ImageGenerationStatusRecordFilesItem
	return marshalUnknown(plain(v), v.Extra)
}

type ImageGenerationStatusRecordErrorEnvelope struct {
	Code    Optional[string]                     `json:"code,omitzero"`
	Message Optional[string]                     `json:"message,omitzero"`
	Meta    Optional[map[string]json.RawMessage] `json:"meta,omitzero"`
	Extra   map[string]json.RawMessage           `json:"-"`
}

func (v *ImageGenerationStatusRecordErrorEnvelope) UnmarshalJSON(b []byte) error {
	type plain ImageGenerationStatusRecordErrorEnvelope
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"code", "message", "meta"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ImageGenerationStatusRecordErrorEnvelope(out)
	return nil
}
func (v ImageGenerationStatusRecordErrorEnvelope) MarshalJSON() ([]byte, error) {
	type plain ImageGenerationStatusRecordErrorEnvelope
	return marshalUnknown(plain(v), v.Extra)
}

type ImageGenerationStatusRecordResult struct {
	Model       Optional[string]                                        `json:"model,omitzero"`
	Provider    Optional[string]                                        `json:"provider,omitzero"`
	Prompt      Optional[string]                                        `json:"prompt,omitzero"`
	Images      Optional[[]ImageGenerationStatusRecordResultImagesItem] `json:"images,omitzero"`
	Dimensions  Optional[ImageGenerationStatusRecordResultDimensions]   `json:"dimensions,omitzero"`
	AspectRatio Optional[string]                                        `json:"aspect_ratio,omitzero"`
	Format      Optional[string]                                        `json:"format,omitzero"`
	Quality     Optional[int64]                                         `json:"quality,omitzero"`
	Warnings    Optional[json.RawMessage]                               `json:"warnings,omitzero"`
	Extra       map[string]json.RawMessage                              `json:"-"`
}

func (v *ImageGenerationStatusRecordResult) UnmarshalJSON(b []byte) error {
	type plain ImageGenerationStatusRecordResult
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"model", "provider", "prompt", "images", "dimensions", "aspect_ratio", "format", "quality", "warnings"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ImageGenerationStatusRecordResult(out)
	return nil
}
func (v ImageGenerationStatusRecordResult) MarshalJSON() ([]byte, error) {
	type plain ImageGenerationStatusRecordResult
	return marshalUnknown(plain(v), v.Extra)
}

type ImageGenerationStatusRecordResultDimensions struct {
	Width  Optional[int64]            `json:"width,omitzero"`
	Height Optional[int64]            `json:"height,omitzero"`
	Extra  map[string]json.RawMessage `json:"-"`
}

func (v *ImageGenerationStatusRecordResultDimensions) UnmarshalJSON(b []byte) error {
	type plain ImageGenerationStatusRecordResultDimensions
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"width", "height"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ImageGenerationStatusRecordResultDimensions(out)
	return nil
}
func (v ImageGenerationStatusRecordResultDimensions) MarshalJSON() ([]byte, error) {
	type plain ImageGenerationStatusRecordResultDimensions
	return marshalUnknown(plain(v), v.Extra)
}

type ImageGenerationStatusRecordResultImagesItem struct {
	ProviderAssetID Optional[string]           `json:"provider_asset_id,omitzero"`
	MimeType        Optional[string]           `json:"mime_type,omitzero"`
	SourceURL       Optional[string]           `json:"source_url,omitzero"`
	Format          Optional[string]           `json:"format,omitzero"`
	Extra           map[string]json.RawMessage `json:"-"`
}

func (v *ImageGenerationStatusRecordResultImagesItem) UnmarshalJSON(b []byte) error {
	type plain ImageGenerationStatusRecordResultImagesItem
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"provider_asset_id", "mime_type", "source_url", "format"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ImageGenerationStatusRecordResultImagesItem(out)
	return nil
}
func (v ImageGenerationStatusRecordResultImagesItem) MarshalJSON() ([]byte, error) {
	type plain ImageGenerationStatusRecordResultImagesItem
	return marshalUnknown(plain(v), v.Extra)
}

type TestImageGenerateResourceData struct {
	Status     string                     `json:"status"`
	Prompt     string                     `json:"prompt"`
	ModelCode  string                     `json:"model_code"`
	FilesCount int64                      `json:"files_count"`
	Files      []TestImageFileMetadata    `json:"files"`
	ReceivedAt string                     `json:"received_at"`
	Extra      map[string]json.RawMessage `json:"-"`
}

func (v *TestImageGenerateResourceData) UnmarshalJSON(b []byte) error {
	type plain TestImageGenerateResourceData
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"status", "prompt", "model_code", "files_count", "files", "received_at"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = TestImageGenerateResourceData(out)
	return nil
}
func (v TestImageGenerateResourceData) MarshalJSON() ([]byte, error) {
	type plain TestImageGenerateResourceData
	return marshalUnknown(plain(v), v.Extra)
}

type CompletionCollectionMeta struct {
	CurrentPage Optional[int64]            `json:"current_page,omitzero"`
	From        Optional[int64]            `json:"from,omitzero"`
	LastPage    Optional[int64]            `json:"last_page,omitzero"`
	PerPage     Optional[int64]            `json:"per_page,omitzero"`
	To          Optional[int64]            `json:"to,omitzero"`
	Total       Optional[int64]            `json:"total,omitzero"`
	Extra       map[string]json.RawMessage `json:"-"`
}

func (v *CompletionCollectionMeta) UnmarshalJSON(b []byte) error {
	type plain CompletionCollectionMeta
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"current_page", "from", "last_page", "per_page", "to", "total"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = CompletionCollectionMeta(out)
	return nil
}
func (v CompletionCollectionMeta) MarshalJSON() ([]byte, error) {
	type plain CompletionCollectionMeta
	return marshalUnknown(plain(v), v.Extra)
}

type CompletionCollectionLinks struct {
	First Optional[string]           `json:"first,omitzero"`
	Last  Optional[string]           `json:"last,omitzero"`
	Prev  Optional[string]           `json:"prev,omitzero"`
	Next  Optional[string]           `json:"next,omitzero"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *CompletionCollectionLinks) UnmarshalJSON(b []byte) error {
	type plain CompletionCollectionLinks
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"first", "last", "prev", "next"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = CompletionCollectionLinks(out)
	return nil
}
func (v CompletionCollectionLinks) MarshalJSON() ([]byte, error) {
	type plain CompletionCollectionLinks
	return marshalUnknown(plain(v), v.Extra)
}

type CompletionUsage struct {
	PromptTokens     Optional[int64]            `json:"prompt_tokens,omitzero"`
	CompletionTokens Optional[int64]            `json:"completion_tokens,omitzero"`
	SpriteCpuMs      Optional[int64]            `json:"sprite_cpu_ms,omitzero"`
	SpriteRamGbMs    Optional[int64]            `json:"sprite_ram_gb_ms,omitzero"`
	Extra            map[string]json.RawMessage `json:"-"`
}

func (v *CompletionUsage) UnmarshalJSON(b []byte) error {
	type plain CompletionUsage
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"prompt_tokens", "completion_tokens", "sprite_cpu_ms", "sprite_ram_gb_ms"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = CompletionUsage(out)
	return nil
}
func (v CompletionUsage) MarshalJSON() ([]byte, error) {
	type plain CompletionUsage
	return marshalUnknown(plain(v), v.Extra)
}

type ChatAttachmentCollectionData struct {
	Attachments []ChatAttachment           `json:"attachments"`
	Extra       map[string]json.RawMessage `json:"-"`
}

func (v *ChatAttachmentCollectionData) UnmarshalJSON(b []byte) error {
	type plain ChatAttachmentCollectionData
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"attachments"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ChatAttachmentCollectionData(out)
	return nil
}
func (v ChatAttachmentCollectionData) MarshalJSON() ([]byte, error) {
	type plain ChatAttachmentCollectionData
	return marshalUnknown(plain(v), v.Extra)
}

type ErrorEnvelopeErrorEnvelope struct {
	Type    string                               `json:"type"`
	Code    Optional[string]                     `json:"code,omitzero"`
	Message string                               `json:"message"`
	Details Optional[map[string]json.RawMessage] `json:"details,omitzero"`
	Extra   map[string]json.RawMessage           `json:"-"`
}

func (v *ErrorEnvelopeErrorEnvelope) UnmarshalJSON(b []byte) error {
	type plain ErrorEnvelopeErrorEnvelope
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"type", "code", "message", "details"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ErrorEnvelopeErrorEnvelope(out)
	return nil
}
func (v ErrorEnvelopeErrorEnvelope) MarshalJSON() ([]byte, error) {
	type plain ErrorEnvelopeErrorEnvelope
	return marshalUnknown(plain(v), v.Extra)
}

type WikiOperationEventPhases struct {
	Inbox  Optional[PhaseResult] `json:"inbox,omitzero"`
	Outbox Optional[PhaseResult] `json:"outbox,omitzero"`
}

type WikiOperationPhases struct {
	Inbox  Optional[PhaseResult] `json:"inbox,omitzero"`
	Outbox Optional[PhaseResult] `json:"outbox,omitzero"`
}

type CompletionTerminalSnapshotData struct {
	TurnID       string                                        `json:"turn_id"`
	Usage        Optional[CompletionTerminalSnapshotUsage]     `json:"usage,omitzero"`
	Cost         Optional[float64]                             `json:"cost,omitzero"`
	FinishReason Optional[string]                              `json:"finish_reason,omitzero"`
	Reason       Optional[string]                              `json:"reason,omitzero"`
	Code         Optional[string]                              `json:"code,omitzero"`
	Message      Optional[string]                              `json:"message,omitzero"`
	ChatUsage    Optional[CompletionTerminalSnapshotChatUsage] `json:"chat_usage,omitzero"`
	ChatCost     Optional[float64]                             `json:"chat_cost,omitzero"`
}
