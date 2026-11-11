package envai

import "encoding/json"

type NativeDeliveryRequest struct {
	StrictModel  Optional[bool] `json:"strict_model,omitzero"`
	WorkspaceID  string         `json:"workspace_id"`
	EventID      string         `json:"event_id"`
	OriginTurnID string         `json:"origin_turn_id,omitempty"`
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

type IndependentJobRequest struct {
	WorkspaceID string         `json:"workspace_id"`
	RequestID   string         `json:"request_id"`
	Prompt      string         `json:"prompt"`
	StrictModel Optional[bool] `json:"strict_model,omitzero"`
}

type IndependentJobReceipt struct {
	TaskID string `json:"task_id"`
	TurnID string `json:"turn_id"`
}

type ChoiceClassificationRequest struct {
	RequestID string                                   `json:"request_id"`
	Context   json.RawMessage                          `json:"context"`
	Question  string                                   `json:"question"`
	Choices   []ChoiceClassificationRequestChoicesItem `json:"choices"`
}

type ChoiceClassificationResult struct {
	Choice       string `json:"choice"`
	Model        string `json:"model"`
	RequestID    string `json:"request_id"`
	InvocationID string `json:"invocation_id"`
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

type WikiWebhookEvent = json.RawMessage

type AuthUser struct {
	ID              string                     `json:"id"`
	Name            string                     `json:"name"`
	Email           string                     `json:"email"`
	EmailVerified   bool                       `json:"email_verified"`
	EmailVerifiedAt *string                    `json:"email_verified_at"`
	IsSuperAdmin    bool                       `json:"is_super_admin"`
	Extra           map[string]json.RawMessage `json:"-"`
}

func (v *AuthUser) UnmarshalJSON(b []byte) error {
	type plain AuthUser
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"id", "name", "email", "email_verified", "email_verified_at", "is_super_admin"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = AuthUser(out)
	return nil
}
func (v AuthUser) MarshalJSON() ([]byte, error) {
	type plain AuthUser
	return marshalUnknown(plain(v), v.Extra)
}

type AuthUserResource struct {
	Data  AuthUser                   `json:"data"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *AuthUserResource) UnmarshalJSON(b []byte) error {
	type plain AuthUserResource
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = AuthUserResource(out)
	return nil
}
func (v AuthUserResource) MarshalJSON() ([]byte, error) {
	type plain AuthUserResource
	return marshalUnknown(plain(v), v.Extra)
}

type EmailVerificationResource struct {
	Data EmailVerificationResourceData `json:"data"`
}

type AuthUserCollection struct {
	Data  []AuthUser                 `json:"data"`
	Links AuthUserCollectionLinks    `json:"links"`
	Meta  AuthUserCollectionMeta     `json:"meta"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *AuthUserCollection) UnmarshalJSON(b []byte) error {
	type plain AuthUserCollection
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data", "links", "meta"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = AuthUserCollection(out)
	return nil
}
func (v AuthUserCollection) MarshalJSON() ([]byte, error) {
	type plain AuthUserCollection
	return marshalUnknown(plain(v), v.Extra)
}

type CsrfResource struct {
	Data  CsrfResourceData           `json:"data"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *CsrfResource) UnmarshalJSON(b []byte) error {
	type plain CsrfResource
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = CsrfResource(out)
	return nil
}
func (v CsrfResource) MarshalJSON() ([]byte, error) {
	type plain CsrfResource
	return marshalUnknown(plain(v), v.Extra)
}

type MessageResource struct {
	Data  MessageResourceData        `json:"data"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *MessageResource) UnmarshalJSON(b []byte) error {
	type plain MessageResource
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = MessageResource(out)
	return nil
}
func (v MessageResource) MarshalJSON() ([]byte, error) {
	type plain MessageResource
	return marshalUnknown(plain(v), v.Extra)
}

type LoginRequest struct {
	Email    string                     `json:"email"`
	Password string                     `json:"password"`
	Extra    map[string]json.RawMessage `json:"-"`
}

func (v *LoginRequest) UnmarshalJSON(b []byte) error {
	type plain LoginRequest
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"email", "password"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = LoginRequest(out)
	return nil
}
func (v LoginRequest) MarshalJSON() ([]byte, error) {
	type plain LoginRequest
	return marshalUnknown(plain(v), v.Extra)
}

type RegisterRequest struct {
	Name        string                     `json:"name"`
	Email       string                     `json:"email"`
	Password    string                     `json:"password"`
	CompanyName string                     `json:"company_name"`
	Extra       map[string]json.RawMessage `json:"-"`
}

func (v *RegisterRequest) UnmarshalJSON(b []byte) error {
	type plain RegisterRequest
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"name", "email", "password", "company_name"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = RegisterRequest(out)
	return nil
}
func (v RegisterRequest) MarshalJSON() ([]byte, error) {
	type plain RegisterRequest
	return marshalUnknown(plain(v), v.Extra)
}

type RegisterFromInvitationRequest struct {
	Token    string                     `json:"token"`
	Name     string                     `json:"name"`
	Email    string                     `json:"email"`
	Password string                     `json:"password"`
	Extra    map[string]json.RawMessage `json:"-"`
}

func (v *RegisterFromInvitationRequest) UnmarshalJSON(b []byte) error {
	type plain RegisterFromInvitationRequest
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"token", "name", "email", "password"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = RegisterFromInvitationRequest(out)
	return nil
}
func (v RegisterFromInvitationRequest) MarshalJSON() ([]byte, error) {
	type plain RegisterFromInvitationRequest
	return marshalUnknown(plain(v), v.Extra)
}

type ForgotPasswordRequest struct {
	Email string                     `json:"email"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *ForgotPasswordRequest) UnmarshalJSON(b []byte) error {
	type plain ForgotPasswordRequest
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"email"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ForgotPasswordRequest(out)
	return nil
}
func (v ForgotPasswordRequest) MarshalJSON() ([]byte, error) {
	type plain ForgotPasswordRequest
	return marshalUnknown(plain(v), v.Extra)
}

type ResetPasswordRequest struct {
	Token    string                     `json:"token"`
	Email    string                     `json:"email"`
	Password string                     `json:"password"`
	Extra    map[string]json.RawMessage `json:"-"`
}

func (v *ResetPasswordRequest) UnmarshalJSON(b []byte) error {
	type plain ResetPasswordRequest
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"token", "email", "password"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ResetPasswordRequest(out)
	return nil
}
func (v ResetPasswordRequest) MarshalJSON() ([]byte, error) {
	type plain ResetPasswordRequest
	return marshalUnknown(plain(v), v.Extra)
}

type ResolveInvitationRequest struct {
	Token string `json:"token"`
}

type InvitationResolution struct {
	Email         string `json:"email"`
	AccountExists bool   `json:"account_exists"`
}

type InvitationResolutionResource struct {
	Data InvitationResolution `json:"data"`
}

type AcceptInvitationRequest struct {
	Token string                     `json:"token"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *AcceptInvitationRequest) UnmarshalJSON(b []byte) error {
	type plain AcceptInvitationRequest
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"token"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = AcceptInvitationRequest(out)
	return nil
}
func (v AcceptInvitationRequest) MarshalJSON() ([]byte, error) {
	type plain AcceptInvitationRequest
	return marshalUnknown(plain(v), v.Extra)
}

type PlatformRoleUpdateRequest struct {
	IsSuperAdmin bool                       `json:"is_super_admin"`
	Extra        map[string]json.RawMessage `json:"-"`
}

func (v *PlatformRoleUpdateRequest) UnmarshalJSON(b []byte) error {
	type plain PlatformRoleUpdateRequest
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"is_super_admin"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = PlatformRoleUpdateRequest(out)
	return nil
}
func (v PlatformRoleUpdateRequest) MarshalJSON() ([]byte, error) {
	type plain PlatformRoleUpdateRequest
	return marshalUnknown(plain(v), v.Extra)
}

type Company struct {
	ID              string                     `json:"id"`
	Name            string                     `json:"name"`
	CreatedByUserID *string                    `json:"created_by_user_id"`
	Role            *string                    `json:"role"`
	CreatedAt       string                     `json:"created_at"`
	UpdatedAt       string                     `json:"updated_at"`
	Extra           map[string]json.RawMessage `json:"-"`
}

func (v *Company) UnmarshalJSON(b []byte) error {
	type plain Company
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"id", "name", "created_by_user_id", "role", "created_at", "updated_at"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = Company(out)
	return nil
}
func (v Company) MarshalJSON() ([]byte, error) {
	type plain Company
	return marshalUnknown(plain(v), v.Extra)
}

type CompanyResource struct {
	Data  Company                    `json:"data"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *CompanyResource) UnmarshalJSON(b []byte) error {
	type plain CompanyResource
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = CompanyResource(out)
	return nil
}
func (v CompanyResource) MarshalJSON() ([]byte, error) {
	type plain CompanyResource
	return marshalUnknown(plain(v), v.Extra)
}

type CompanyCollection struct {
	Data  []Company                  `json:"data"`
	Links CompanyCollectionLinks     `json:"links"`
	Meta  CompanyCollectionMeta      `json:"meta"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *CompanyCollection) UnmarshalJSON(b []byte) error {
	type plain CompanyCollection
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data", "links", "meta"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = CompanyCollection(out)
	return nil
}
func (v CompanyCollection) MarshalJSON() ([]byte, error) {
	type plain CompanyCollection
	return marshalUnknown(plain(v), v.Extra)
}

type CompanyCreateRequest struct {
	Name  string                     `json:"name"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *CompanyCreateRequest) UnmarshalJSON(b []byte) error {
	type plain CompanyCreateRequest
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"name"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = CompanyCreateRequest(out)
	return nil
}
func (v CompanyCreateRequest) MarshalJSON() ([]byte, error) {
	type plain CompanyCreateRequest
	return marshalUnknown(plain(v), v.Extra)
}

type CompanyUpdateRequest struct {
	Name  string                     `json:"name"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *CompanyUpdateRequest) UnmarshalJSON(b []byte) error {
	type plain CompanyUpdateRequest
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"name"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = CompanyUpdateRequest(out)
	return nil
}
func (v CompanyUpdateRequest) MarshalJSON() ([]byte, error) {
	type plain CompanyUpdateRequest
	return marshalUnknown(plain(v), v.Extra)
}

type CompanyMember struct {
	ID        string                     `json:"id"`
	CompanyID string                     `json:"company_id"`
	UserID    string                     `json:"user_id"`
	Name      *string                    `json:"name"`
	Email     *string                    `json:"email"`
	Role      string                     `json:"role"`
	CreatedAt string                     `json:"created_at"`
	UpdatedAt string                     `json:"updated_at"`
	Extra     map[string]json.RawMessage `json:"-"`
}

func (v *CompanyMember) UnmarshalJSON(b []byte) error {
	type plain CompanyMember
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"id", "company_id", "user_id", "name", "email", "role", "created_at", "updated_at"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = CompanyMember(out)
	return nil
}
func (v CompanyMember) MarshalJSON() ([]byte, error) {
	type plain CompanyMember
	return marshalUnknown(plain(v), v.Extra)
}

type CompanyMemberResource struct {
	Data  CompanyMember              `json:"data"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *CompanyMemberResource) UnmarshalJSON(b []byte) error {
	type plain CompanyMemberResource
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = CompanyMemberResource(out)
	return nil
}
func (v CompanyMemberResource) MarshalJSON() ([]byte, error) {
	type plain CompanyMemberResource
	return marshalUnknown(plain(v), v.Extra)
}

type CompanyMemberCollection struct {
	Data  []CompanyMember              `json:"data"`
	Links CompanyMemberCollectionLinks `json:"links"`
	Meta  CompanyMemberCollectionMeta  `json:"meta"`
	Extra map[string]json.RawMessage   `json:"-"`
}

func (v *CompanyMemberCollection) UnmarshalJSON(b []byte) error {
	type plain CompanyMemberCollection
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data", "links", "meta"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = CompanyMemberCollection(out)
	return nil
}
func (v CompanyMemberCollection) MarshalJSON() ([]byte, error) {
	type plain CompanyMemberCollection
	return marshalUnknown(plain(v), v.Extra)
}

type CompanyMemberRoleUpdateRequest struct {
	Role  string                     `json:"role"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *CompanyMemberRoleUpdateRequest) UnmarshalJSON(b []byte) error {
	type plain CompanyMemberRoleUpdateRequest
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"role"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = CompanyMemberRoleUpdateRequest(out)
	return nil
}
func (v CompanyMemberRoleUpdateRequest) MarshalJSON() ([]byte, error) {
	type plain CompanyMemberRoleUpdateRequest
	return marshalUnknown(plain(v), v.Extra)
}

type CompanyInvitation struct {
	ID              string                     `json:"id"`
	CompanyID       string                     `json:"company_id"`
	Email           string                     `json:"email"`
	Role            string                     `json:"role"`
	Status          string                     `json:"status"`
	ExpiresAt       *string                    `json:"expires_at"`
	AcceptedUserID  *string                    `json:"accepted_user_id"`
	AcceptedAt      *string                    `json:"accepted_at"`
	InvitedByUserID *string                    `json:"invited_by_user_id"`
	CreatedAt       string                     `json:"created_at"`
	UpdatedAt       string                     `json:"updated_at"`
	Extra           map[string]json.RawMessage `json:"-"`
}

func (v *CompanyInvitation) UnmarshalJSON(b []byte) error {
	type plain CompanyInvitation
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"id", "company_id", "email", "role", "status", "expires_at", "accepted_user_id", "accepted_at", "invited_by_user_id", "created_at", "updated_at"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = CompanyInvitation(out)
	return nil
}
func (v CompanyInvitation) MarshalJSON() ([]byte, error) {
	type plain CompanyInvitation
	return marshalUnknown(plain(v), v.Extra)
}

type CompanyInvitationResource struct {
	Data  CompanyInvitation          `json:"data"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *CompanyInvitationResource) UnmarshalJSON(b []byte) error {
	type plain CompanyInvitationResource
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = CompanyInvitationResource(out)
	return nil
}
func (v CompanyInvitationResource) MarshalJSON() ([]byte, error) {
	type plain CompanyInvitationResource
	return marshalUnknown(plain(v), v.Extra)
}

type CompanyInvitationCollection struct {
	Data  []CompanyInvitation              `json:"data"`
	Links CompanyInvitationCollectionLinks `json:"links"`
	Meta  CompanyInvitationCollectionMeta  `json:"meta"`
	Extra map[string]json.RawMessage       `json:"-"`
}

func (v *CompanyInvitationCollection) UnmarshalJSON(b []byte) error {
	type plain CompanyInvitationCollection
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data", "links", "meta"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = CompanyInvitationCollection(out)
	return nil
}
func (v CompanyInvitationCollection) MarshalJSON() ([]byte, error) {
	type plain CompanyInvitationCollection
	return marshalUnknown(plain(v), v.Extra)
}

type CompanyInvitationCreateRequest struct {
	Email string                     `json:"email"`
	Role  string                     `json:"role"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *CompanyInvitationCreateRequest) UnmarshalJSON(b []byte) error {
	type plain CompanyInvitationCreateRequest
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"email", "role"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = CompanyInvitationCreateRequest(out)
	return nil
}
func (v CompanyInvitationCreateRequest) MarshalJSON() ([]byte, error) {
	type plain CompanyInvitationCreateRequest
	return marshalUnknown(plain(v), v.Extra)
}

type Project struct {
	ID          string                     `json:"id"`
	CompanyID   string                     `json:"company_id"`
	Name        string                     `json:"name"`
	Description *string                    `json:"description"`
	CreatedAt   string                     `json:"created_at"`
	UpdatedAt   string                     `json:"updated_at"`
	Extra       map[string]json.RawMessage `json:"-"`
}

func (v *Project) UnmarshalJSON(b []byte) error {
	type plain Project
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"id", "company_id", "name", "description", "created_at", "updated_at"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = Project(out)
	return nil
}
func (v Project) MarshalJSON() ([]byte, error) {
	type plain Project
	return marshalUnknown(plain(v), v.Extra)
}

type ProjectResource struct {
	Data  Project                    `json:"data"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *ProjectResource) UnmarshalJSON(b []byte) error {
	type plain ProjectResource
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ProjectResource(out)
	return nil
}
func (v ProjectResource) MarshalJSON() ([]byte, error) {
	type plain ProjectResource
	return marshalUnknown(plain(v), v.Extra)
}

type ProjectCollection struct {
	Data  []Project                  `json:"data"`
	Links ProjectCollectionLinks     `json:"links"`
	Meta  ProjectCollectionMeta      `json:"meta"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *ProjectCollection) UnmarshalJSON(b []byte) error {
	type plain ProjectCollection
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data", "links", "meta"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ProjectCollection(out)
	return nil
}
func (v ProjectCollection) MarshalJSON() ([]byte, error) {
	type plain ProjectCollection
	return marshalUnknown(plain(v), v.Extra)
}

type ProjectCreateRequest struct {
	Name        string                     `json:"name"`
	Description Optional[string]           `json:"description,omitzero"`
	Extra       map[string]json.RawMessage `json:"-"`
}

func (v *ProjectCreateRequest) UnmarshalJSON(b []byte) error {
	type plain ProjectCreateRequest
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"name", "description"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ProjectCreateRequest(out)
	return nil
}
func (v ProjectCreateRequest) MarshalJSON() ([]byte, error) {
	type plain ProjectCreateRequest
	return marshalUnknown(plain(v), v.Extra)
}

type ProjectUpdateRequest struct {
	Name        string                     `json:"name"`
	Description Optional[string]           `json:"description,omitzero"`
	Extra       map[string]json.RawMessage `json:"-"`
}

func (v *ProjectUpdateRequest) UnmarshalJSON(b []byte) error {
	type plain ProjectUpdateRequest
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"name", "description"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ProjectUpdateRequest(out)
	return nil
}
func (v ProjectUpdateRequest) MarshalJSON() ([]byte, error) {
	type plain ProjectUpdateRequest
	return marshalUnknown(plain(v), v.Extra)
}

type BillingWallet struct {
	CompanyID               string                            `json:"company_id"`
	Currency                string                            `json:"currency"`
	Balance                 *string                           `json:"balance"`
	Reserved                *string                           `json:"reserved"`
	Available               *string                           `json:"available"`
	Revision                int64                             `json:"revision"`
	Enabled                 bool                              `json:"enabled"`
	PaymentsConfigured      bool                              `json:"payments_configured"`
	PaymentMethodConfigured bool                              `json:"payment_method_configured"`
	FinancialState          Optional[string]                  `json:"financial_state,omitzero"`
	FinancialReviewReason   Optional[string]                  `json:"financial_review_reason,omitzero"`
	AutoRecharge            BillingAutoRecharge               `json:"auto_recharge"`
	Capability              Optional[BillingWalletCapability] `json:"capability,omitzero"`
}

type BillingWalletResource struct {
	Data  BillingWallet              `json:"data"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *BillingWalletResource) UnmarshalJSON(b []byte) error {
	type plain BillingWalletResource
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = BillingWalletResource(out)
	return nil
}
func (v BillingWalletResource) MarshalJSON() ([]byte, error) {
	type plain BillingWalletResource
	return marshalUnknown(plain(v), v.Extra)
}

type BillingTransaction struct {
	ID           string           `json:"id"`
	Kind         string           `json:"kind"`
	Amount       string           `json:"amount"`
	Currency     string           `json:"currency"`
	Status       string           `json:"status"`
	CreatedAt    string           `json:"created_at"`
	CreditAmount Optional[string] `json:"credit_amount,omitzero"`
	AmountPaid   Optional[string] `json:"amount_paid,omitzero"`
	TaxAmount    Optional[string] `json:"tax_amount,omitzero"`
}

type BillingTransactionCollection struct {
	Data  []BillingTransaction              `json:"data"`
	Links BillingTransactionCollectionLinks `json:"links"`
	Meta  BillingTransactionCollectionMeta  `json:"meta"`
	Extra map[string]json.RawMessage        `json:"-"`
}

func (v *BillingTransactionCollection) UnmarshalJSON(b []byte) error {
	type plain BillingTransactionCollection
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data", "links", "meta"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = BillingTransactionCollection(out)
	return nil
}
func (v BillingTransactionCollection) MarshalJSON() ([]byte, error) {
	type plain BillingTransactionCollection
	return marshalUnknown(plain(v), v.Extra)
}

type BillingTopupQuoteRequest struct {
	CreditAmount string `json:"credit_amount"`
}

type BillingTopupQuote struct {
	CreditAmount    string `json:"credit_amount"`
	AmountBeforeTax string `json:"amount_before_tax"`
	Currency        string `json:"currency"`
	QuoteVersion    string `json:"quote_version"`
	TaxStatus       string `json:"tax_status"`
}

type BillingTopupQuoteResource struct {
	Data  BillingTopupQuote          `json:"data"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *BillingTopupQuoteResource) UnmarshalJSON(b []byte) error {
	type plain BillingTopupQuoteResource
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = BillingTopupQuoteResource(out)
	return nil
}
func (v BillingTopupQuoteResource) MarshalJSON() ([]byte, error) {
	type plain BillingTopupQuoteResource
	return marshalUnknown(plain(v), v.Extra)
}

type BillingTopupRequest struct {
	CreditAmount   string           `json:"credit_amount"`
	QuoteVersion   string           `json:"quote_version"`
	ConsentVersion Optional[string] `json:"consent_version,omitzero"`
}

type BillingTopupCheckout struct {
	TopupID         string `json:"topup_id"`
	CheckoutURL     string `json:"checkout_url"`
	CreditAmount    string `json:"credit_amount"`
	AmountBeforeTax string `json:"amount_before_tax"`
	Currency        string `json:"currency"`
	QuoteVersion    string `json:"quote_version"`
	TaxStatus       string `json:"tax_status"`
}

type BillingTopupCheckoutResource struct {
	Data  BillingTopupCheckout       `json:"data"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *BillingTopupCheckoutResource) UnmarshalJSON(b []byte) error {
	type plain BillingTopupCheckoutResource
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = BillingTopupCheckoutResource(out)
	return nil
}
func (v BillingTopupCheckoutResource) MarshalJSON() ([]byte, error) {
	type plain BillingTopupCheckoutResource
	return marshalUnknown(plain(v), v.Extra)
}

type BillingPaymentMethodSetupRequest struct {
	ConsentVersion string `json:"consent_version"`
}

type BillingPaymentMethodSetup struct {
	SetupID        string `json:"setup_id"`
	CheckoutURL    string `json:"checkout_url"`
	ConsentVersion string `json:"consent_version"`
}

type BillingPaymentMethodSetupResource struct {
	Data  BillingPaymentMethodSetup  `json:"data"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *BillingPaymentMethodSetupResource) UnmarshalJSON(b []byte) error {
	type plain BillingPaymentMethodSetupResource
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = BillingPaymentMethodSetupResource(out)
	return nil
}
func (v BillingPaymentMethodSetupResource) MarshalJSON() ([]byte, error) {
	type plain BillingPaymentMethodSetupResource
	return marshalUnknown(plain(v), v.Extra)
}

type BillingAutoRechargeUpdateRequest struct {
	Enabled        bool             `json:"enabled"`
	Amount         Optional[string] `json:"amount,omitzero"`
	Threshold      Optional[string] `json:"threshold,omitzero"`
	ConsentVersion Optional[string] `json:"consent_version,omitzero"`
}

type BillingAutoRecharge struct {
	Enabled        bool    `json:"enabled"`
	Amount         string  `json:"amount"`
	Threshold      string  `json:"threshold"`
	State          string  `json:"state"`
	ConsentVersion *string `json:"consent_version"`
}

type BillingAutoRechargeResource struct {
	Data  BillingAutoRecharge        `json:"data"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *BillingAutoRechargeResource) UnmarshalJSON(b []byte) error {
	type plain BillingAutoRechargeResource
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = BillingAutoRechargeResource(out)
	return nil
}
func (v BillingAutoRechargeResource) MarshalJSON() ([]byte, error) {
	type plain BillingAutoRechargeResource
	return marshalUnknown(plain(v), v.Extra)
}

type BillingSpendLimitReset struct {
	Unit     string `json:"unit"`
	Every    int64  `json:"every"`
	AnchorAt string `json:"anchor_at"`
	Timezone string `json:"timezone"`
}

type BillingSpendPeriod struct {
	StartsAt     string  `json:"starts_at"`
	EndsAt       *string `json:"ends_at"`
	Spent        string  `json:"spent"`
	Reserved     string  `json:"reserved"`
	Remaining    *string `json:"remaining"`
	ResetVersion int64   `json:"reset_version"`
}

type BillingSpendLimit struct {
	Scope     string          `json:"scope"`
	ScopeID   string          `json:"scope_id"`
	ProjectID string          `json:"project_id"`
	APIKeyID  *string         `json:"api_key_id"`
	Amount    *string         `json:"amount"`
	Reset     json.RawMessage `json:"reset"`
	Period    json.RawMessage `json:"period"`
	CanManage bool            `json:"can_manage"`
}

type BillingSpendLimitResource struct {
	Data  BillingSpendLimit          `json:"data"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *BillingSpendLimitResource) UnmarshalJSON(b []byte) error {
	type plain BillingSpendLimitResource
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = BillingSpendLimitResource(out)
	return nil
}
func (v BillingSpendLimitResource) MarshalJSON() ([]byte, error) {
	type plain BillingSpendLimitResource
	return marshalUnknown(plain(v), v.Extra)
}

type BillingSpendLimitUpdateRequest struct {
	Amount *string         `json:"amount"`
	Reset  json.RawMessage `json:"reset"`
}

type BillingWebhookOk struct {
	Status string                     `json:"status"`
	Extra  map[string]json.RawMessage `json:"-"`
}

func (v *BillingWebhookOk) UnmarshalJSON(b []byte) error {
	type plain BillingWebhookOk
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"status"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = BillingWebhookOk(out)
	return nil
}
func (v BillingWebhookOk) MarshalJSON() ([]byte, error) {
	type plain BillingWebhookOk
	return marshalUnknown(plain(v), v.Extra)
}

type BillingError struct {
	ErrorEnvelope BillingErrorErrorEnvelope  `json:"error"`
	Extra         map[string]json.RawMessage `json:"-"`
}

func (v *BillingError) UnmarshalJSON(b []byte) error {
	type plain BillingError
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"error"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = BillingError(out)
	return nil
}
func (v BillingError) MarshalJSON() ([]byte, error) {
	type plain BillingError
	return marshalUnknown(plain(v), v.Extra)
}

type ProjectMember struct {
	ID        string                     `json:"id"`
	CompanyID string                     `json:"company_id"`
	ProjectID string                     `json:"project_id"`
	UserID    string                     `json:"user_id"`
	Name      *string                    `json:"name"`
	Email     *string                    `json:"email"`
	CreatedAt string                     `json:"created_at"`
	UpdatedAt string                     `json:"updated_at"`
	Extra     map[string]json.RawMessage `json:"-"`
}

func (v *ProjectMember) UnmarshalJSON(b []byte) error {
	type plain ProjectMember
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"id", "company_id", "project_id", "user_id", "name", "email", "created_at", "updated_at"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ProjectMember(out)
	return nil
}
func (v ProjectMember) MarshalJSON() ([]byte, error) {
	type plain ProjectMember
	return marshalUnknown(plain(v), v.Extra)
}

type ProjectMemberResource struct {
	Data  ProjectMember              `json:"data"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *ProjectMemberResource) UnmarshalJSON(b []byte) error {
	type plain ProjectMemberResource
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ProjectMemberResource(out)
	return nil
}
func (v ProjectMemberResource) MarshalJSON() ([]byte, error) {
	type plain ProjectMemberResource
	return marshalUnknown(plain(v), v.Extra)
}

type ProjectMemberCollection struct {
	Data  []ProjectMember              `json:"data"`
	Links ProjectMemberCollectionLinks `json:"links"`
	Meta  ProjectMemberCollectionMeta  `json:"meta"`
	Extra map[string]json.RawMessage   `json:"-"`
}

func (v *ProjectMemberCollection) UnmarshalJSON(b []byte) error {
	type plain ProjectMemberCollection
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data", "links", "meta"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ProjectMemberCollection(out)
	return nil
}
func (v ProjectMemberCollection) MarshalJSON() ([]byte, error) {
	type plain ProjectMemberCollection
	return marshalUnknown(plain(v), v.Extra)
}

type CleanupIntentState struct {
	ID            string                     `json:"id"`
	Kind          string                     `json:"kind"`
	Status        string                     `json:"status"`
	Attempts      int64                      `json:"attempts"`
	DeadlineAt    *string                    `json:"deadline_at"`
	StartedAt     *string                    `json:"started_at"`
	FinishedAt    *string                    `json:"finished_at"`
	LastErrorCode *string                    `json:"last_error_code"`
	Extra         map[string]json.RawMessage `json:"-"`
}

func (v *CleanupIntentState) UnmarshalJSON(b []byte) error {
	type plain CleanupIntentState
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"id", "kind", "status", "attempts", "deadline_at", "started_at", "finished_at", "last_error_code"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = CleanupIntentState(out)
	return nil
}
func (v CleanupIntentState) MarshalJSON() ([]byte, error) {
	type plain CleanupIntentState
	return marshalUnknown(plain(v), v.Extra)
}

type CleanupResult struct {
	ResourceType string                     `json:"resource_type"`
	ResourceID   string                     `json:"resource_id"`
	Cleanup      []CleanupIntentState       `json:"cleanup"`
	Extra        map[string]json.RawMessage `json:"-"`
}

func (v *CleanupResult) UnmarshalJSON(b []byte) error {
	type plain CleanupResult
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"resource_type", "resource_id", "cleanup"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = CleanupResult(out)
	return nil
}
func (v CleanupResult) MarshalJSON() ([]byte, error) {
	type plain CleanupResult
	return marshalUnknown(plain(v), v.Extra)
}

type CleanupResource struct {
	Data  CleanupResult              `json:"data"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *CleanupResource) UnmarshalJSON(b []byte) error {
	type plain CleanupResource
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = CleanupResource(out)
	return nil
}
func (v CleanupResource) MarshalJSON() ([]byte, error) {
	type plain CleanupResource
	return marshalUnknown(plain(v), v.Extra)
}

type UsageMoney struct {
	Currency string                     `json:"currency"`
	Amount   string                     `json:"amount"`
	Extra    map[string]json.RawMessage `json:"-"`
}

func (v *UsageMoney) UnmarshalJSON(b []byte) error {
	type plain UsageMoney
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"currency", "amount"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = UsageMoney(out)
	return nil
}
func (v UsageMoney) MarshalJSON() ([]byte, error) {
	type plain UsageMoney
	return marshalUnknown(plain(v), v.Extra)
}

type UsageKeyBreakdown struct {
	APIKeyID                    string                     `json:"api_key_id"`
	Totals                      []UsageMoney               `json:"totals"`
	InvocationCount             int64                      `json:"invocation_count"`
	UnknownInvocations          int64                      `json:"unknown_invocations"`
	NotSubmittedInvocations     int64                      `json:"not_submitted_invocations"`
	CompletionCount             int64                      `json:"completion_count"`
	UnknownCurrencyObservations int64                      `json:"unknown_currency_observations"`
	Extra                       map[string]json.RawMessage `json:"-"`
}

func (v *UsageKeyBreakdown) UnmarshalJSON(b []byte) error {
	type plain UsageKeyBreakdown
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"api_key_id", "totals", "invocation_count", "unknown_invocations", "not_submitted_invocations", "completion_count", "unknown_currency_observations"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = UsageKeyBreakdown(out)
	return nil
}
func (v UsageKeyBreakdown) MarshalJSON() ([]byte, error) {
	type plain UsageKeyBreakdown
	return marshalUnknown(plain(v), v.Extra)
}

type UsageProjectBreakdown struct {
	ProjectID               string                     `json:"project_id"`
	Totals                  []UsageMoney               `json:"totals"`
	InvocationCount         int64                      `json:"invocation_count"`
	UnknownInvocations      int64                      `json:"unknown_invocations"`
	NotSubmittedInvocations int64                      `json:"not_submitted_invocations"`
	CompletionCount         int64                      `json:"completion_count"`
	Media                   UsageMediaTotals           `json:"media"`
	Extra                   map[string]json.RawMessage `json:"-"`
}

func (v *UsageProjectBreakdown) UnmarshalJSON(b []byte) error {
	type plain UsageProjectBreakdown
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"project_id", "totals", "invocation_count", "unknown_invocations", "not_submitted_invocations", "completion_count", "media"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = UsageProjectBreakdown(out)
	return nil
}
func (v UsageProjectBreakdown) MarshalJSON() ([]byte, error) {
	type plain UsageProjectBreakdown
	return marshalUnknown(plain(v), v.Extra)
}

type UsageMediaTotals struct {
	Totals                      []UsageMoney               `json:"totals"`
	PricedObservations          int64                      `json:"priced_observations"`
	UnknownObservations         int64                      `json:"unknown_observations"`
	UnknownCurrencyObservations int64                      `json:"unknown_currency_observations"`
	ByAPIKey                    []UsageKeyBreakdown        `json:"by_api_key"`
	Extra                       map[string]json.RawMessage `json:"-"`
}

func (v *UsageMediaTotals) UnmarshalJSON(b []byte) error {
	type plain UsageMediaTotals
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"totals", "priced_observations", "unknown_observations", "unknown_currency_observations", "by_api_key"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = UsageMediaTotals(out)
	return nil
}
func (v UsageMediaTotals) MarshalJSON() ([]byte, error) {
	type plain UsageMediaTotals
	return marshalUnknown(plain(v), v.Extra)
}

type UsageReportDay struct {
	Date                    string                     `json:"date"`
	Totals                  []UsageMoney               `json:"totals"`
	InvocationCount         int64                      `json:"invocation_count"`
	UnknownInvocations      int64                      `json:"unknown_invocations"`
	NotSubmittedInvocations int64                      `json:"not_submitted_invocations"`
	CompletionCount         int64                      `json:"completion_count"`
	ByAPIKey                []UsageKeyBreakdown        `json:"by_api_key"`
	Media                   UsageMediaTotals           `json:"media"`
	ByProject               []UsageProjectBreakdown    `json:"by_project"`
	Extra                   map[string]json.RawMessage `json:"-"`
}

func (v *UsageReportDay) UnmarshalJSON(b []byte) error {
	type plain UsageReportDay
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"date", "totals", "invocation_count", "unknown_invocations", "not_submitted_invocations", "completion_count", "by_api_key", "media", "by_project"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = UsageReportDay(out)
	return nil
}
func (v UsageReportDay) MarshalJSON() ([]byte, error) {
	type plain UsageReportDay
	return marshalUnknown(plain(v), v.Extra)
}

type UsageReportCollection struct {
	Data  []UsageReportDay           `json:"data"`
	Links UsageReportCollectionLinks `json:"links"`
	Meta  UsageReportCollectionMeta  `json:"meta"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *UsageReportCollection) UnmarshalJSON(b []byte) error {
	type plain UsageReportCollection
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data", "links", "meta"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = UsageReportCollection(out)
	return nil
}
func (v UsageReportCollection) MarshalJSON() ([]byte, error) {
	type plain UsageReportCollection
	return marshalUnknown(plain(v), v.Extra)
}

type ProjectLogEntry struct {
	ID            string                     `json:"id"`
	ProjectID     string                     `json:"project_id"`
	APIKeyID      *string                    `json:"api_key_id"`
	ActorAPIKeyID *string                    `json:"actor_api_key_id"`
	ChatID        string                     `json:"chat_id"`
	WorkspaceID   string                     `json:"workspace_id"`
	Seq           int64                      `json:"seq"`
	TurnID        *string                    `json:"turn_id"`
	Type          string                     `json:"type"`
	Ts            string                     `json:"ts"`
	Agent         map[string]json.RawMessage `json:"agent"`
	Data          map[string]json.RawMessage `json:"data"`
	Extra         map[string]json.RawMessage `json:"-"`
}

func (v *ProjectLogEntry) UnmarshalJSON(b []byte) error {
	type plain ProjectLogEntry
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"id", "project_id", "api_key_id", "actor_api_key_id", "chat_id", "workspace_id", "seq", "turn_id", "type", "ts", "agent", "data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ProjectLogEntry(out)
	return nil
}
func (v ProjectLogEntry) MarshalJSON() ([]byte, error) {
	type plain ProjectLogEntry
	return marshalUnknown(plain(v), v.Extra)
}

type ProjectEventTypeFilters struct {
	Types []string `json:"types"`
}

type ProjectEventTypeFiltersResource struct {
	Data ProjectEventTypeFilters `json:"data"`
}

type ProjectLogCollection struct {
	Data  []ProjectLogEntry          `json:"data"`
	Links ProjectLogCollectionLinks  `json:"links"`
	Meta  ProjectLogCollectionMeta   `json:"meta"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *ProjectLogCollection) UnmarshalJSON(b []byte) error {
	type plain ProjectLogCollection
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data", "links", "meta"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ProjectLogCollection(out)
	return nil
}
func (v ProjectLogCollection) MarshalJSON() ([]byte, error) {
	type plain ProjectLogCollection
	return marshalUnknown(plain(v), v.Extra)
}

type InternalWikiDispatchRequest struct {
	ProjectID   string                           `json:"project_id"`
	APIKeyID    string                           `json:"api_key_id"`
	WorkspaceID string                           `json:"workspace_id"`
	Actor       InternalWikiDispatchRequestActor `json:"actor"`
	Action      string                           `json:"action"`
	Arguments   json.RawMessage                  `json:"arguments"`
	Extra       map[string]json.RawMessage       `json:"-"`
}

func (v *InternalWikiDispatchRequest) UnmarshalJSON(b []byte) error {
	type plain InternalWikiDispatchRequest
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"project_id", "api_key_id", "workspace_id", "actor", "action", "arguments"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = InternalWikiDispatchRequest(out)
	return nil
}
func (v InternalWikiDispatchRequest) MarshalJSON() ([]byte, error) {
	type plain InternalWikiDispatchRequest
	return marshalUnknown(plain(v), v.Extra)
}

type InternalDispatchResponse struct {
	Data  map[string]json.RawMessage `json:"data"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *InternalDispatchResponse) UnmarshalJSON(b []byte) error {
	type plain InternalDispatchResponse
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = InternalDispatchResponse(out)
	return nil
}
func (v InternalDispatchResponse) MarshalJSON() ([]byte, error) {
	type plain InternalDispatchResponse
	return marshalUnknown(plain(v), v.Extra)
}

type InternalAttachmentResolveRequest struct {
	ProjectID      string                     `json:"project_id"`
	APIKeyID       string                     `json:"api_key_id"`
	WorkspaceID    string                     `json:"workspace_id"`
	ChatID         string                     `json:"chat_id"`
	Message        string                     `json:"message"`
	IdempotencyKey string                     `json:"idempotency_key"`
	AttachmentIDs  []string                   `json:"attachment_ids"`
	Extra          map[string]json.RawMessage `json:"-"`
}

func (v *InternalAttachmentResolveRequest) UnmarshalJSON(b []byte) error {
	type plain InternalAttachmentResolveRequest
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"project_id", "api_key_id", "workspace_id", "chat_id", "message", "idempotency_key", "attachment_ids"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = InternalAttachmentResolveRequest(out)
	return nil
}
func (v InternalAttachmentResolveRequest) MarshalJSON() ([]byte, error) {
	type plain InternalAttachmentResolveRequest
	return marshalUnknown(plain(v), v.Extra)
}

type InternalAttachmentResolveResource struct {
	Data  InternalAttachmentResolveResourceData `json:"data"`
	Extra map[string]json.RawMessage            `json:"-"`
}

func (v *InternalAttachmentResolveResource) UnmarshalJSON(b []byte) error {
	type plain InternalAttachmentResolveResource
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = InternalAttachmentResolveResource(out)
	return nil
}
func (v InternalAttachmentResolveResource) MarshalJSON() ([]byte, error) {
	type plain InternalAttachmentResolveResource
	return marshalUnknown(plain(v), v.Extra)
}

type InternalAttachmentManifest struct {
	AttachmentID string                     `json:"attachment_id"`
	OriginalName string                     `json:"original_name"`
	Revision     int64                      `json:"revision"`
	Sha256       string                     `json:"sha256"`
	Bytes        int64                      `json:"bytes"`
	Conversion   AttachmentConversion       `json:"conversion"`
	Extra        map[string]json.RawMessage `json:"-"`
}

func (v *InternalAttachmentManifest) UnmarshalJSON(b []byte) error {
	type plain InternalAttachmentManifest
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"attachment_id", "original_name", "revision", "sha256", "bytes", "conversion"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = InternalAttachmentManifest(out)
	return nil
}
func (v InternalAttachmentManifest) MarshalJSON() ([]byte, error) {
	type plain InternalAttachmentManifest
	return marshalUnknown(plain(v), v.Extra)
}

type InternalBillingToolLimits struct {
	MaxPoweredMs   int64 `json:"max_powered_ms"`
	CleanupGraceMs int64 `json:"cleanup_grace_ms"`
}

type InternalBillingLLMRates struct {
	Prompt     string `json:"prompt"`
	Completion string `json:"completion"`
	Request    string `json:"request"`
}

type InternalBillingReservation struct {
	ReservationID string                              `json:"reservation_id"`
	MaxAmount     string                              `json:"max_amount"`
	Currency      string                              `json:"currency"`
	MaxTokens     int64                               `json:"max_tokens"`
	MaxPrice      json.RawMessage                     `json:"max_price"`
	ToolLimits    Optional[InternalBillingToolLimits] `json:"tool_limits,omitzero"`
	Status        string                              `json:"status"`
}

type InternalBillingReservationResource struct {
	Data  InternalBillingReservation `json:"data"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *InternalBillingReservationResource) UnmarshalJSON(b []byte) error {
	type plain InternalBillingReservationResource
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = InternalBillingReservationResource(out)
	return nil
}
func (v InternalBillingReservationResource) MarshalJSON() ([]byte, error) {
	type plain InternalBillingReservationResource
	return marshalUnknown(plain(v), v.Extra)
}

type InternalBillingAdmissionRequest struct {
	ProjectID   string `json:"project_id"`
	APIKeyID    string `json:"api_key_id"`
	WorkspaceID string `json:"workspace_id"`
	ChatID      string `json:"chat_id"`
}

type InternalBillingReserveRequest struct {
	InvocationID string           `json:"invocation_id"`
	OwnerToken   string           `json:"owner_token"`
	ProjectID    string           `json:"project_id"`
	APIKeyID     string           `json:"api_key_id"`
	WorkspaceID  string           `json:"workspace_id"`
	ChatID       Optional[string] `json:"chat_id,omitzero"`
	OriginType   string           `json:"origin_type"`
	OriginID     string           `json:"origin_id"`
	ClaimToken   Optional[string] `json:"claim_token,omitzero"`
	ClaimFence   Optional[int64]  `json:"claim_fence,omitzero"`
	Purpose      Optional[string] `json:"purpose,omitzero"`
	Operation    Optional[string] `json:"operation,omitzero"`
}

type InternalBillingSettleRequest struct {
	InvocationID  string           `json:"invocation_id"`
	OwnerToken    string           `json:"owner_token"`
	ProjectID     string           `json:"project_id"`
	APIKeyID      string           `json:"api_key_id"`
	WorkspaceID   string           `json:"workspace_id"`
	ChatID        Optional[string] `json:"chat_id,omitzero"`
	OriginType    string           `json:"origin_type"`
	OriginID      string           `json:"origin_id"`
	ClaimToken    Optional[string] `json:"claim_token,omitzero"`
	ClaimFence    Optional[int64]  `json:"claim_fence,omitzero"`
	Purpose       Optional[string] `json:"purpose,omitzero"`
	Operation     Optional[string] `json:"operation,omitzero"`
	ObservationID Optional[string] `json:"observation_id,omitzero"`
}

type InternalBillingReleaseRequest struct {
	InvocationID string           `json:"invocation_id"`
	OwnerToken   string           `json:"owner_token"`
	ProjectID    string           `json:"project_id"`
	APIKeyID     string           `json:"api_key_id"`
	WorkspaceID  string           `json:"workspace_id"`
	ChatID       Optional[string] `json:"chat_id,omitzero"`
	OriginType   string           `json:"origin_type"`
	OriginID     string           `json:"origin_id"`
	ClaimToken   Optional[string] `json:"claim_token,omitzero"`
	ClaimFence   Optional[int64]  `json:"claim_fence,omitzero"`
	Purpose      Optional[string] `json:"purpose,omitzero"`
	Operation    Optional[string] `json:"operation,omitzero"`
}

type InternalBillingAdmissionResource struct {
	Data  InternalBillingAdmissionResourceData `json:"data"`
	Extra map[string]json.RawMessage           `json:"-"`
}

func (v *InternalBillingAdmissionResource) UnmarshalJSON(b []byte) error {
	type plain InternalBillingAdmissionResource
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = InternalBillingAdmissionResource(out)
	return nil
}
func (v InternalBillingAdmissionResource) MarshalJSON() ([]byte, error) {
	type plain InternalBillingAdmissionResource
	return marshalUnknown(plain(v), v.Extra)
}

type InternalBillingSettleResource struct {
	Data  InternalBillingSettleResourceData `json:"data"`
	Extra map[string]json.RawMessage        `json:"-"`
}

func (v *InternalBillingSettleResource) UnmarshalJSON(b []byte) error {
	type plain InternalBillingSettleResource
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = InternalBillingSettleResource(out)
	return nil
}
func (v InternalBillingSettleResource) MarshalJSON() ([]byte, error) {
	type plain InternalBillingSettleResource
	return marshalUnknown(plain(v), v.Extra)
}

type InternalBillingReleaseResource struct {
	Data  InternalBillingReleaseResourceData `json:"data"`
	Extra map[string]json.RawMessage         `json:"-"`
}

func (v *InternalBillingReleaseResource) UnmarshalJSON(b []byte) error {
	type plain InternalBillingReleaseResource
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = InternalBillingReleaseResource(out)
	return nil
}
func (v InternalBillingReleaseResource) MarshalJSON() ([]byte, error) {
	type plain InternalBillingReleaseResource
	return marshalUnknown(plain(v), v.Extra)
}

type InternalBillingNotice = json.RawMessage

type InternalBillingCompanyNotice struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Reason   string `json:"reason"`
	Scope    string `json:"scope"`
	Currency string `json:"currency"`
}

type InternalBillingScopedNotice struct {
	ID        string           `json:"id"`
	Type      string           `json:"type"`
	Reason    string           `json:"reason"`
	Scope     string           `json:"scope"`
	Currency  string           `json:"currency"`
	Amount    string           `json:"amount"`
	Remaining string           `json:"remaining"`
	ResetAt   Optional[string] `json:"reset_at,omitzero"`
}

type InternalBillingError struct {
	ErrorEnvelope InternalBillingErrorErrorEnvelope `json:"error"`
}

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

type ChatCronConflictError struct {
	ErrorEnvelope ChatCronConflictErrorErrorEnvelope `json:"error"`
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
	ID              string                     `json:"id"`
	ProjectID       string                     `json:"project_id"`
	CreatedByUserID *string                    `json:"created_by_user_id"`
	Name            string                     `json:"name"`
	Metadata        WorkspaceMetadata          `json:"metadata"`
	CreatedAt       string                     `json:"created_at"`
	UpdatedAt       string                     `json:"updated_at"`
	Extra           map[string]json.RawMessage `json:"-"`
}

func (v *Workspace) UnmarshalJSON(b []byte) error {
	type plain Workspace
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"id", "project_id", "created_by_user_id", "name", "metadata", "created_at", "updated_at"})
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

type WorkspaceMetadata = json.RawMessage

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

type ProjectTool struct {
	ID             string                     `json:"id"`
	Name           string                     `json:"name"`
	Kind           string                     `json:"kind"`
	Description    *string                    `json:"description"`
	Parameters     map[string]json.RawMessage `json:"parameters"`
	TimeoutSeconds int64                      `json:"timeout_seconds"`
	CreatedAt      string                     `json:"created_at"`
	UpdatedAt      string                     `json:"updated_at"`
}

type ProjectToolResource struct {
	Data  ProjectTool                `json:"data"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *ProjectToolResource) UnmarshalJSON(b []byte) error {
	type plain ProjectToolResource
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ProjectToolResource(out)
	return nil
}
func (v ProjectToolResource) MarshalJSON() ([]byte, error) {
	type plain ProjectToolResource
	return marshalUnknown(plain(v), v.Extra)
}

type ProjectToolCollection struct {
	Data  []ProjectTool              `json:"data"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *ProjectToolCollection) UnmarshalJSON(b []byte) error {
	type plain ProjectToolCollection
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ProjectToolCollection(out)
	return nil
}
func (v ProjectToolCollection) MarshalJSON() ([]byte, error) {
	type plain ProjectToolCollection
	return marshalUnknown(plain(v), v.Extra)
}

type ProjectToolCreateRequest struct {
	Name           string                     `json:"name"`
	Kind           Optional[string]           `json:"kind,omitzero"`
	Description    Optional[string]           `json:"description,omitzero"`
	Parameters     map[string]json.RawMessage `json:"parameters"`
	TimeoutSeconds Optional[int64]            `json:"timeout_seconds,omitzero"`
}

type ProjectToolUpdateRequest struct {
	Description    Optional[string]                     `json:"description,omitzero"`
	Parameters     Optional[map[string]json.RawMessage] `json:"parameters,omitzero"`
	TimeoutSeconds Optional[int64]                      `json:"timeout_seconds,omitzero"`
}

type ProjectSkill struct {
	ID              string  `json:"id"`
	ProjectID       string  `json:"project_id"`
	CreatedByUserID *string `json:"created_by_user_id"`
	Name            string  `json:"name"`
	Description     *string `json:"description"`
	Content         string  `json:"content"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

type ProjectSkillResource struct {
	Data  ProjectSkill               `json:"data"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *ProjectSkillResource) UnmarshalJSON(b []byte) error {
	type plain ProjectSkillResource
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ProjectSkillResource(out)
	return nil
}
func (v ProjectSkillResource) MarshalJSON() ([]byte, error) {
	type plain ProjectSkillResource
	return marshalUnknown(plain(v), v.Extra)
}

type ProjectSkillCollection struct {
	Data  []ProjectSkill             `json:"data"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *ProjectSkillCollection) UnmarshalJSON(b []byte) error {
	type plain ProjectSkillCollection
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ProjectSkillCollection(out)
	return nil
}
func (v ProjectSkillCollection) MarshalJSON() ([]byte, error) {
	type plain ProjectSkillCollection
	return marshalUnknown(plain(v), v.Extra)
}

type ProjectSkillCreateRequest map[string]json.RawMessage

type ProjectSkillUpdateRequest map[string]json.RawMessage

type Chat struct {
	ID              string                     `json:"id"`
	WorkspaceID     string                     `json:"workspace_id"`
	ProjectID       string                     `json:"project_id"`
	CreatedByUserID *string                    `json:"created_by_user_id"`
	TeamID          string                     `json:"team_id"`
	Title           *string                    `json:"title"`
	CreatedAt       string                     `json:"created_at"`
	UpdatedAt       string                     `json:"updated_at"`
	Extra           map[string]json.RawMessage `json:"-"`
}

func (v *Chat) UnmarshalJSON(b []byte) error {
	type plain Chat
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"id", "workspace_id", "project_id", "created_by_user_id", "team_id", "title", "created_at", "updated_at"})
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

type ChatCron struct {
	ID              string                     `json:"id"`
	WorkspaceID     string                     `json:"workspace_id"`
	OriginChatID    string                     `json:"origin_chat_id"`
	TeamID          string                     `json:"team_id"`
	TargetMode      string                     `json:"target_mode"`
	Prompt          string                     `json:"prompt"`
	IntervalMinutes *int64                     `json:"interval_minutes"`
	DailyAt         *string                    `json:"daily_at"`
	Timezone        *string                    `json:"timezone"`
	StartAt         *string                    `json:"start_at"`
	RepeatCount     *int64                     `json:"repeat_count"`
	Until           *string                    `json:"until"`
	Status          string                     `json:"status"`
	ScheduledCount  int64                      `json:"scheduled_count"`
	DeliveryCount   int64                      `json:"delivery_count"`
	NextDueAt       *string                    `json:"next_due_at"`
	MaxUntil        string                     `json:"max_until"`
	CreatedAt       string                     `json:"created_at"`
	UpdatedAt       string                     `json:"updated_at"`
	Extra           map[string]json.RawMessage `json:"-"`
}

func (v *ChatCron) UnmarshalJSON(b []byte) error {
	type plain ChatCron
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"id", "workspace_id", "origin_chat_id", "team_id", "target_mode", "prompt", "interval_minutes", "daily_at", "timezone", "start_at", "repeat_count", "until", "status", "scheduled_count", "delivery_count", "next_due_at", "max_until", "created_at", "updated_at"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ChatCron(out)
	return nil
}
func (v ChatCron) MarshalJSON() ([]byte, error) {
	type plain ChatCron
	return marshalUnknown(plain(v), v.Extra)
}

type ChatCronResource struct {
	Data  ChatCron                   `json:"data"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *ChatCronResource) UnmarshalJSON(b []byte) error {
	type plain ChatCronResource
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ChatCronResource(out)
	return nil
}
func (v ChatCronResource) MarshalJSON() ([]byte, error) {
	type plain ChatCronResource
	return marshalUnknown(plain(v), v.Extra)
}

type ChatCronCollection struct {
	Data  []ChatCron                 `json:"data"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *ChatCronCollection) UnmarshalJSON(b []byte) error {
	type plain ChatCronCollection
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"data"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ChatCronCollection(out)
	return nil
}
func (v ChatCronCollection) MarshalJSON() ([]byte, error) {
	type plain ChatCronCollection
	return marshalUnknown(plain(v), v.Extra)
}

type ChatCronUpdateRequest struct {
	TargetMode      Optional[string]           `json:"target_mode,omitzero"`
	Prompt          Optional[string]           `json:"prompt,omitzero"`
	IntervalMinutes Optional[int64]            `json:"interval_minutes,omitzero"`
	DailyAt         Optional[string]           `json:"daily_at,omitzero"`
	Timezone        Optional[string]           `json:"timezone,omitzero"`
	StartAt         Optional[string]           `json:"start_at,omitzero"`
	RepeatCount     Optional[int64]            `json:"repeat_count,omitzero"`
	Until           Optional[string]           `json:"until,omitzero"`
	Extra           map[string]json.RawMessage `json:"-"`
}

func (v *ChatCronUpdateRequest) UnmarshalJSON(b []byte) error {
	type plain ChatCronUpdateRequest
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"target_mode", "prompt", "interval_minutes", "daily_at", "timezone", "start_at", "repeat_count", "until"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ChatCronUpdateRequest(out)
	return nil
}
func (v ChatCronUpdateRequest) MarshalJSON() ([]byte, error) {
	type plain ChatCronUpdateRequest
	return marshalUnknown(plain(v), v.Extra)
}

type ChatCronToolArguments = json.RawMessage

type ChatCronToolCreateArguments struct {
	Action          string           `json:"action"`
	Prompt          string           `json:"prompt"`
	TargetMode      Optional[string] `json:"target_mode,omitzero"`
	IntervalMinutes Optional[int64]  `json:"interval_minutes,omitzero"`
	DailyAt         Optional[string] `json:"daily_at,omitzero"`
	Timezone        Optional[string] `json:"timezone,omitzero"`
	StartAt         Optional[string] `json:"start_at,omitzero"`
	RepeatCount     Optional[int64]  `json:"repeat_count,omitzero"`
	Until           Optional[string] `json:"until,omitzero"`
}

type ChatCronToolListArguments struct {
	Action string `json:"action"`
}

type ChatCronToolCancelArguments struct {
	Action string `json:"action"`
	CronID string `json:"cron_id"`
}

type ChatUpdateRequest struct {
	Title Optional[string]           `json:"title,omitzero"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *ChatUpdateRequest) UnmarshalJSON(b []byte) error {
	type plain ChatUpdateRequest
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"title"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ChatUpdateRequest(out)
	return nil
}
func (v ChatUpdateRequest) MarshalJSON() ([]byte, error) {
	type plain ChatUpdateRequest
	return marshalUnknown(plain(v), v.Extra)
}

type Completion struct {
	ID                string                               `json:"id"`
	WorkspaceID       string                               `json:"workspace_id"`
	ChatID            string                               `json:"chat_id"`
	TeamID            string                               `json:"team_id"`
	ProjectID         string                               `json:"project_id"`
	APIKeyID          string                               `json:"api_key_id"`
	UserID            *string                              `json:"user_id"`
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
	extra, err := unknownFields(b, []string{"id", "workspace_id", "chat_id", "team_id", "project_id", "api_key_id", "user_id", "model_orchestrator", "models_subagents", "started_at", "finished_at", "duration_ms", "usage", "billing", "usage_cost", "terminal_snapshot", "cost_usd", "status", "finish_reason", "content_hash", "created_at", "updated_at"})
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
	ProjectID            string                     `json:"project_id"`
	CreatedByUserID      *string                    `json:"created_by_user_id"`
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
	extra, err := unknownFields(b, []string{"id", "project_id", "created_by_user_id", "name", "description", "system_prompt", "model", "mode", "tools_enabled", "external_tools_enabled", "permissions", "can_write_knowledge", "can_write_skills", "config", "created_at", "updated_at"})
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
	ID              string                     `json:"id"`
	ProjectID       string                     `json:"project_id"`
	CreatedByUserID *string                    `json:"created_by_user_id"`
	Name            string                     `json:"name"`
	Description     *string                    `json:"description"`
	OrchestratorID  string                     `json:"orchestrator_id"`
	Members         []TeamMember               `json:"members"`
	CreatedAt       string                     `json:"created_at"`
	UpdatedAt       string                     `json:"updated_at"`
	Extra           map[string]json.RawMessage `json:"-"`
}

func (v *Team) UnmarshalJSON(b []byte) error {
	type plain Team
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"id", "project_id", "created_by_user_id", "name", "description", "orchestrator_id", "members", "created_at", "updated_at"})
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
	ID              string                     `json:"id"`
	ProjectID       string                     `json:"project_id"`
	CreatedByUserID *string                    `json:"created_by_user_id"`
	Name            string                     `json:"name"`
	Scopes          []string                   `json:"scopes"`
	ExpiresAt       *string                    `json:"expires_at"`
	RevokedAt       *string                    `json:"revoked_at"`
	CreatedAt       string                     `json:"created_at"`
	Extra           map[string]json.RawMessage `json:"-"`
}

func (v *APIKey) UnmarshalJSON(b []byte) error {
	type plain APIKey
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"id", "project_id", "created_by_user_id", "name", "scopes", "expires_at", "revoked_at", "created_at"})
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
	Name       string                                   `json:"name"`
	Scopes     Optional[[]string]                       `json:"scopes,omitzero"`
	ExpiresAt  Optional[string]                         `json:"expires_at,omitzero"`
	SpendLimit Optional[BillingSpendLimitUpdateRequest] `json:"spend_limit,omitzero"`
	Extra      map[string]json.RawMessage               `json:"-"`
}

func (v *APIKeyCreate) UnmarshalJSON(b []byte) error {
	type plain APIKeyCreate
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"name", "scopes", "expires_at", "spend_limit"})
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

type ActivityDurationBucket struct {
	LowerMs float64  `json:"lower_ms"`
	UpperMs *float64 `json:"upper_ms"`
	Count   int64    `json:"count"`
}

type ActivityDurationDistribution struct {
	CandidateCount int64                    `json:"candidate_count"`
	SampleCount    int64                    `json:"sample_count"`
	MissingCount   int64                    `json:"missing_count"`
	InvalidCount   int64                    `json:"invalid_count"`
	ZeroCount      int64                    `json:"zero_count"`
	MinMs          *float64                 `json:"min_ms"`
	MaxMs          *float64                 `json:"max_ms"`
	AvgMs          *float64                 `json:"avg_ms"`
	P50Ms          *float64                 `json:"p50_ms"`
	P95Ms          *float64                 `json:"p95_ms"`
	Buckets        []ActivityDurationBucket `json:"buckets"`
}

type ActivityDailyCounts struct {
	Date               string `json:"date"`
	EventCount         int64  `json:"event_count"`
	ToolStartCount     int64  `json:"tool_start_count"`
	FileReadStartCount int64  `json:"file_read_start_count"`
	TurnFailureCount   int64  `json:"turn_failure_count"`
	AgentFailureCount  int64  `json:"agent_failure_count"`
	ToolFailureCount   int64  `json:"tool_failure_count"`
}

type ActivityEventTypeGroups struct {
	Items             []ActivityEventTypeGroupsItemsItem `json:"items"`
	OtherCount        int64                              `json:"other_count"`
	OmittedGroupCount int64                              `json:"omitted_group_count"`
}

type ActivityToolGroups struct {
	Items             []ActivityToolGroupsItemsItem `json:"items"`
	OtherCount        int64                         `json:"other_count"`
	OmittedGroupCount int64                         `json:"omitted_group_count"`
	MissingNameCount  int64                         `json:"missing_name_count"`
}

type ProjectActivityAnalytics struct {
	Window     ProjectActivityAnalyticsWindow          `json:"window"`
	Coverage   ProjectActivityAnalyticsCoverage        `json:"coverage"`
	Daily      []ActivityDailyCounts                   `json:"daily"`
	EventTypes ActivityEventTypeGroups                 `json:"event_types"`
	Tools      ActivityToolGroups                      `json:"tools"`
	Failures   ProjectActivityAnalyticsFailures        `json:"failures"`
	Durations  ProjectActivityAnalyticsDurations       `json:"durations"`
	FileReads  []ProjectActivityAnalyticsFileReadsItem `json:"file_reads"`
}

type ProjectActivityAnalyticsResource struct {
	Data ProjectActivityAnalytics `json:"data"`
}

type UserTool = ProjectTool

type UserToolResource = ProjectToolResource

type UserToolCollection = ProjectToolCollection

type UserToolCreateRequest = ProjectToolCreateRequest

type UserToolUpdateRequest = ProjectToolUpdateRequest

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

type ProjectActivityAnalyticsFileReadsItem struct {
	ToolName string `json:"tool_name"`
	Count    int64  `json:"count"`
}

type ProjectActivityAnalyticsDurations struct {
	ToolExecution ActivityDurationDistribution `json:"tool_execution"`
	TurnExecution ActivityDurationDistribution `json:"turn_execution"`
}

type ProjectActivityAnalyticsFailures struct {
	TurnEventCount               int64 `json:"turn_event_count"`
	AgentEventCount              int64 `json:"agent_event_count"`
	ToolEventCount               int64 `json:"tool_event_count"`
	ToolResultUnknownStatusCount int64 `json:"tool_result_unknown_status_count"`
}

type ProjectActivityAnalyticsCoverage struct {
	Source                   string  `json:"source"`
	Aggregation              string  `json:"aggregation"`
	ExecutionHistoryComplete bool    `json:"execution_history_complete"`
	ConfiguredRetentionDays  int64   `json:"configured_retention_days"`
	ObservedEventCount       int64   `json:"observed_event_count"`
	OldestObservedAt         *string `json:"oldest_observed_at"`
	NewestObservedAt         *string `json:"newest_observed_at"`
}

type ProjectActivityAnalyticsWindow struct {
	From     string  `json:"from"`
	To       string  `json:"to"`
	Timezone string  `json:"timezone"`
	APIKeyID *string `json:"api_key_id"`
	ChatID   *string `json:"chat_id"`
}

type ActivityToolGroupsItemsItem struct {
	ToolName string  `json:"tool_name"`
	Kind     *string `json:"kind"`
	Count    int64   `json:"count"`
}

type ActivityEventTypeGroupsItemsItem struct {
	Type  string `json:"type"`
	Count int64  `json:"count"`
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

type ChatCronConflictErrorErrorEnvelope struct {
	Type    string `json:"type"`
	Code    string `json:"code"`
	Message string `json:"message"`
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

type InternalBillingErrorErrorEnvelope struct {
	Type    string          `json:"type"`
	Code    string          `json:"code"`
	Details json.RawMessage `json:"details"`
}

type InternalBillingReleaseResourceData struct {
	ReservationID string `json:"reservation_id"`
	Status        string `json:"status"`
}

type InternalBillingSettleResourceData struct {
	ReservationID string                  `json:"reservation_id"`
	Status        string                  `json:"status"`
	Notices       []InternalBillingNotice `json:"notices"`
}

type InternalBillingAdmissionResourceData struct {
	Allowed bool `json:"allowed"`
}

type InternalAttachmentResolveResourceData struct {
	ProjectID     string                       `json:"project_id"`
	APIKeyID      string                       `json:"api_key_id"`
	InputID       string                       `json:"input_id"`
	TurnID        string                       `json:"turn_id"`
	Message       string                       `json:"message"`
	PreparedInput string                       `json:"prepared_input"`
	Manifest      []InternalAttachmentManifest `json:"manifest"`
	Accepted      bool                         `json:"accepted"`
	Extra         map[string]json.RawMessage   `json:"-"`
}

func (v *InternalAttachmentResolveResourceData) UnmarshalJSON(b []byte) error {
	type plain InternalAttachmentResolveResourceData
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"project_id", "api_key_id", "input_id", "turn_id", "message", "prepared_input", "manifest", "accepted"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = InternalAttachmentResolveResourceData(out)
	return nil
}
func (v InternalAttachmentResolveResourceData) MarshalJSON() ([]byte, error) {
	type plain InternalAttachmentResolveResourceData
	return marshalUnknown(plain(v), v.Extra)
}

type InternalWikiDispatchRequestActor struct {
	OperationID  string                     `json:"operation_id"`
	ClaimToken   string                     `json:"claim_token"`
	Fence        int64                      `json:"fence"`
	Phase        Optional[string]           `json:"phase,omitzero"`
	TurnID       Optional[string]           `json:"turn_id,omitzero"`
	ChatID       Optional[string]           `json:"chat_id,omitzero"`
	FileID       Optional[string]           `json:"file_id,omitzero"`
	AttemptID    Optional[string]           `json:"attempt_id,omitzero"`
	AttemptFence Optional[int64]            `json:"attempt_fence,omitzero"`
	Extra        map[string]json.RawMessage `json:"-"`
}

func (v *InternalWikiDispatchRequestActor) UnmarshalJSON(b []byte) error {
	type plain InternalWikiDispatchRequestActor
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"operation_id", "claim_token", "fence", "phase", "turn_id", "chat_id", "file_id", "attempt_id", "attempt_fence"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = InternalWikiDispatchRequestActor(out)
	return nil
}
func (v InternalWikiDispatchRequestActor) MarshalJSON() ([]byte, error) {
	type plain InternalWikiDispatchRequestActor
	return marshalUnknown(plain(v), v.Extra)
}

type ProjectLogCollectionMeta struct {
	CurrentPage Optional[int64]            `json:"current_page,omitzero"`
	From        Optional[int64]            `json:"from,omitzero"`
	LastPage    Optional[int64]            `json:"last_page,omitzero"`
	PerPage     Optional[int64]            `json:"per_page,omitzero"`
	To          Optional[int64]            `json:"to,omitzero"`
	Total       Optional[int64]            `json:"total,omitzero"`
	Extra       map[string]json.RawMessage `json:"-"`
}

func (v *ProjectLogCollectionMeta) UnmarshalJSON(b []byte) error {
	type plain ProjectLogCollectionMeta
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"current_page", "from", "last_page", "per_page", "to", "total"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ProjectLogCollectionMeta(out)
	return nil
}
func (v ProjectLogCollectionMeta) MarshalJSON() ([]byte, error) {
	type plain ProjectLogCollectionMeta
	return marshalUnknown(plain(v), v.Extra)
}

type ProjectLogCollectionLinks struct {
	First Optional[string]           `json:"first,omitzero"`
	Last  Optional[string]           `json:"last,omitzero"`
	Prev  Optional[string]           `json:"prev,omitzero"`
	Next  Optional[string]           `json:"next,omitzero"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *ProjectLogCollectionLinks) UnmarshalJSON(b []byte) error {
	type plain ProjectLogCollectionLinks
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"first", "last", "prev", "next"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ProjectLogCollectionLinks(out)
	return nil
}
func (v ProjectLogCollectionLinks) MarshalJSON() ([]byte, error) {
	type plain ProjectLogCollectionLinks
	return marshalUnknown(plain(v), v.Extra)
}

type UsageReportCollectionMeta struct {
	CurrentPage Optional[int64]            `json:"current_page,omitzero"`
	From        Optional[int64]            `json:"from,omitzero"`
	LastPage    Optional[int64]            `json:"last_page,omitzero"`
	PerPage     Optional[int64]            `json:"per_page,omitzero"`
	To          Optional[int64]            `json:"to,omitzero"`
	Total       Optional[int64]            `json:"total,omitzero"`
	Extra       map[string]json.RawMessage `json:"-"`
}

func (v *UsageReportCollectionMeta) UnmarshalJSON(b []byte) error {
	type plain UsageReportCollectionMeta
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"current_page", "from", "last_page", "per_page", "to", "total"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = UsageReportCollectionMeta(out)
	return nil
}
func (v UsageReportCollectionMeta) MarshalJSON() ([]byte, error) {
	type plain UsageReportCollectionMeta
	return marshalUnknown(plain(v), v.Extra)
}

type UsageReportCollectionLinks struct {
	First Optional[string]           `json:"first,omitzero"`
	Last  Optional[string]           `json:"last,omitzero"`
	Prev  Optional[string]           `json:"prev,omitzero"`
	Next  Optional[string]           `json:"next,omitzero"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *UsageReportCollectionLinks) UnmarshalJSON(b []byte) error {
	type plain UsageReportCollectionLinks
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"first", "last", "prev", "next"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = UsageReportCollectionLinks(out)
	return nil
}
func (v UsageReportCollectionLinks) MarshalJSON() ([]byte, error) {
	type plain UsageReportCollectionLinks
	return marshalUnknown(plain(v), v.Extra)
}

type ProjectMemberCollectionMeta struct {
	CurrentPage Optional[int64]            `json:"current_page,omitzero"`
	From        Optional[int64]            `json:"from,omitzero"`
	LastPage    Optional[int64]            `json:"last_page,omitzero"`
	PerPage     Optional[int64]            `json:"per_page,omitzero"`
	To          Optional[int64]            `json:"to,omitzero"`
	Total       Optional[int64]            `json:"total,omitzero"`
	Extra       map[string]json.RawMessage `json:"-"`
}

func (v *ProjectMemberCollectionMeta) UnmarshalJSON(b []byte) error {
	type plain ProjectMemberCollectionMeta
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"current_page", "from", "last_page", "per_page", "to", "total"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ProjectMemberCollectionMeta(out)
	return nil
}
func (v ProjectMemberCollectionMeta) MarshalJSON() ([]byte, error) {
	type plain ProjectMemberCollectionMeta
	return marshalUnknown(plain(v), v.Extra)
}

type ProjectMemberCollectionLinks struct {
	First Optional[string]           `json:"first,omitzero"`
	Last  Optional[string]           `json:"last,omitzero"`
	Prev  Optional[string]           `json:"prev,omitzero"`
	Next  Optional[string]           `json:"next,omitzero"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *ProjectMemberCollectionLinks) UnmarshalJSON(b []byte) error {
	type plain ProjectMemberCollectionLinks
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"first", "last", "prev", "next"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ProjectMemberCollectionLinks(out)
	return nil
}
func (v ProjectMemberCollectionLinks) MarshalJSON() ([]byte, error) {
	type plain ProjectMemberCollectionLinks
	return marshalUnknown(plain(v), v.Extra)
}

type BillingErrorErrorEnvelope struct {
	Type    string                     `json:"type"`
	Code    string                     `json:"code"`
	Message Optional[string]           `json:"message,omitzero"`
	Extra   map[string]json.RawMessage `json:"-"`
}

func (v *BillingErrorErrorEnvelope) UnmarshalJSON(b []byte) error {
	type plain BillingErrorErrorEnvelope
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"type", "code", "message"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = BillingErrorErrorEnvelope(out)
	return nil
}
func (v BillingErrorErrorEnvelope) MarshalJSON() ([]byte, error) {
	type plain BillingErrorErrorEnvelope
	return marshalUnknown(plain(v), v.Extra)
}

type BillingTransactionCollectionMeta struct {
	CurrentPage Optional[int64]            `json:"current_page,omitzero"`
	From        Optional[int64]            `json:"from,omitzero"`
	LastPage    Optional[int64]            `json:"last_page,omitzero"`
	PerPage     Optional[int64]            `json:"per_page,omitzero"`
	To          Optional[int64]            `json:"to,omitzero"`
	Total       Optional[int64]            `json:"total,omitzero"`
	Extra       map[string]json.RawMessage `json:"-"`
}

func (v *BillingTransactionCollectionMeta) UnmarshalJSON(b []byte) error {
	type plain BillingTransactionCollectionMeta
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"current_page", "from", "last_page", "per_page", "to", "total"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = BillingTransactionCollectionMeta(out)
	return nil
}
func (v BillingTransactionCollectionMeta) MarshalJSON() ([]byte, error) {
	type plain BillingTransactionCollectionMeta
	return marshalUnknown(plain(v), v.Extra)
}

type BillingTransactionCollectionLinks struct {
	First Optional[string]           `json:"first,omitzero"`
	Last  Optional[string]           `json:"last,omitzero"`
	Prev  Optional[string]           `json:"prev,omitzero"`
	Next  Optional[string]           `json:"next,omitzero"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *BillingTransactionCollectionLinks) UnmarshalJSON(b []byte) error {
	type plain BillingTransactionCollectionLinks
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"first", "last", "prev", "next"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = BillingTransactionCollectionLinks(out)
	return nil
}
func (v BillingTransactionCollectionLinks) MarshalJSON() ([]byte, error) {
	type plain BillingTransactionCollectionLinks
	return marshalUnknown(plain(v), v.Extra)
}

type BillingWalletCapability struct {
	Configured Optional[bool]   `json:"configured,omitzero"`
	Reason     Optional[string] `json:"reason,omitzero"`
	Version    Optional[string] `json:"version,omitzero"`
}

type ProjectCollectionMeta struct {
	CurrentPage Optional[int64]            `json:"current_page,omitzero"`
	From        Optional[int64]            `json:"from,omitzero"`
	LastPage    Optional[int64]            `json:"last_page,omitzero"`
	PerPage     Optional[int64]            `json:"per_page,omitzero"`
	To          Optional[int64]            `json:"to,omitzero"`
	Total       Optional[int64]            `json:"total,omitzero"`
	Extra       map[string]json.RawMessage `json:"-"`
}

func (v *ProjectCollectionMeta) UnmarshalJSON(b []byte) error {
	type plain ProjectCollectionMeta
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"current_page", "from", "last_page", "per_page", "to", "total"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ProjectCollectionMeta(out)
	return nil
}
func (v ProjectCollectionMeta) MarshalJSON() ([]byte, error) {
	type plain ProjectCollectionMeta
	return marshalUnknown(plain(v), v.Extra)
}

type ProjectCollectionLinks struct {
	First Optional[string]           `json:"first,omitzero"`
	Last  Optional[string]           `json:"last,omitzero"`
	Prev  Optional[string]           `json:"prev,omitzero"`
	Next  Optional[string]           `json:"next,omitzero"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *ProjectCollectionLinks) UnmarshalJSON(b []byte) error {
	type plain ProjectCollectionLinks
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"first", "last", "prev", "next"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ProjectCollectionLinks(out)
	return nil
}
func (v ProjectCollectionLinks) MarshalJSON() ([]byte, error) {
	type plain ProjectCollectionLinks
	return marshalUnknown(plain(v), v.Extra)
}

type CompanyInvitationCollectionMeta struct {
	CurrentPage Optional[int64]            `json:"current_page,omitzero"`
	From        Optional[int64]            `json:"from,omitzero"`
	LastPage    Optional[int64]            `json:"last_page,omitzero"`
	PerPage     Optional[int64]            `json:"per_page,omitzero"`
	To          Optional[int64]            `json:"to,omitzero"`
	Total       Optional[int64]            `json:"total,omitzero"`
	Extra       map[string]json.RawMessage `json:"-"`
}

func (v *CompanyInvitationCollectionMeta) UnmarshalJSON(b []byte) error {
	type plain CompanyInvitationCollectionMeta
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"current_page", "from", "last_page", "per_page", "to", "total"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = CompanyInvitationCollectionMeta(out)
	return nil
}
func (v CompanyInvitationCollectionMeta) MarshalJSON() ([]byte, error) {
	type plain CompanyInvitationCollectionMeta
	return marshalUnknown(plain(v), v.Extra)
}

type CompanyInvitationCollectionLinks struct {
	First Optional[string]           `json:"first,omitzero"`
	Last  Optional[string]           `json:"last,omitzero"`
	Prev  Optional[string]           `json:"prev,omitzero"`
	Next  Optional[string]           `json:"next,omitzero"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *CompanyInvitationCollectionLinks) UnmarshalJSON(b []byte) error {
	type plain CompanyInvitationCollectionLinks
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"first", "last", "prev", "next"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = CompanyInvitationCollectionLinks(out)
	return nil
}
func (v CompanyInvitationCollectionLinks) MarshalJSON() ([]byte, error) {
	type plain CompanyInvitationCollectionLinks
	return marshalUnknown(plain(v), v.Extra)
}

type CompanyMemberCollectionMeta struct {
	CurrentPage Optional[int64]            `json:"current_page,omitzero"`
	From        Optional[int64]            `json:"from,omitzero"`
	LastPage    Optional[int64]            `json:"last_page,omitzero"`
	PerPage     Optional[int64]            `json:"per_page,omitzero"`
	To          Optional[int64]            `json:"to,omitzero"`
	Total       Optional[int64]            `json:"total,omitzero"`
	Extra       map[string]json.RawMessage `json:"-"`
}

func (v *CompanyMemberCollectionMeta) UnmarshalJSON(b []byte) error {
	type plain CompanyMemberCollectionMeta
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"current_page", "from", "last_page", "per_page", "to", "total"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = CompanyMemberCollectionMeta(out)
	return nil
}
func (v CompanyMemberCollectionMeta) MarshalJSON() ([]byte, error) {
	type plain CompanyMemberCollectionMeta
	return marshalUnknown(plain(v), v.Extra)
}

type CompanyMemberCollectionLinks struct {
	First Optional[string]           `json:"first,omitzero"`
	Last  Optional[string]           `json:"last,omitzero"`
	Prev  Optional[string]           `json:"prev,omitzero"`
	Next  Optional[string]           `json:"next,omitzero"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *CompanyMemberCollectionLinks) UnmarshalJSON(b []byte) error {
	type plain CompanyMemberCollectionLinks
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"first", "last", "prev", "next"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = CompanyMemberCollectionLinks(out)
	return nil
}
func (v CompanyMemberCollectionLinks) MarshalJSON() ([]byte, error) {
	type plain CompanyMemberCollectionLinks
	return marshalUnknown(plain(v), v.Extra)
}

type CompanyCollectionMeta struct {
	CurrentPage Optional[int64]            `json:"current_page,omitzero"`
	From        Optional[int64]            `json:"from,omitzero"`
	LastPage    Optional[int64]            `json:"last_page,omitzero"`
	PerPage     Optional[int64]            `json:"per_page,omitzero"`
	To          Optional[int64]            `json:"to,omitzero"`
	Total       Optional[int64]            `json:"total,omitzero"`
	Extra       map[string]json.RawMessage `json:"-"`
}

func (v *CompanyCollectionMeta) UnmarshalJSON(b []byte) error {
	type plain CompanyCollectionMeta
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"current_page", "from", "last_page", "per_page", "to", "total"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = CompanyCollectionMeta(out)
	return nil
}
func (v CompanyCollectionMeta) MarshalJSON() ([]byte, error) {
	type plain CompanyCollectionMeta
	return marshalUnknown(plain(v), v.Extra)
}

type CompanyCollectionLinks struct {
	First Optional[string]           `json:"first,omitzero"`
	Last  Optional[string]           `json:"last,omitzero"`
	Prev  Optional[string]           `json:"prev,omitzero"`
	Next  Optional[string]           `json:"next,omitzero"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *CompanyCollectionLinks) UnmarshalJSON(b []byte) error {
	type plain CompanyCollectionLinks
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"first", "last", "prev", "next"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = CompanyCollectionLinks(out)
	return nil
}
func (v CompanyCollectionLinks) MarshalJSON() ([]byte, error) {
	type plain CompanyCollectionLinks
	return marshalUnknown(plain(v), v.Extra)
}

type MessageResourceData struct {
	Message string                     `json:"message"`
	Extra   map[string]json.RawMessage `json:"-"`
}

func (v *MessageResourceData) UnmarshalJSON(b []byte) error {
	type plain MessageResourceData
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"message"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = MessageResourceData(out)
	return nil
}
func (v MessageResourceData) MarshalJSON() ([]byte, error) {
	type plain MessageResourceData
	return marshalUnknown(plain(v), v.Extra)
}

type CsrfResourceData struct {
	CsrfToken string                     `json:"csrf_token"`
	Extra     map[string]json.RawMessage `json:"-"`
}

func (v *CsrfResourceData) UnmarshalJSON(b []byte) error {
	type plain CsrfResourceData
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"csrf_token"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = CsrfResourceData(out)
	return nil
}
func (v CsrfResourceData) MarshalJSON() ([]byte, error) {
	type plain CsrfResourceData
	return marshalUnknown(plain(v), v.Extra)
}

type AuthUserCollectionMeta struct {
	CurrentPage Optional[int64]            `json:"current_page,omitzero"`
	From        Optional[int64]            `json:"from,omitzero"`
	LastPage    Optional[int64]            `json:"last_page,omitzero"`
	PerPage     Optional[int64]            `json:"per_page,omitzero"`
	To          Optional[int64]            `json:"to,omitzero"`
	Total       Optional[int64]            `json:"total,omitzero"`
	Extra       map[string]json.RawMessage `json:"-"`
}

func (v *AuthUserCollectionMeta) UnmarshalJSON(b []byte) error {
	type plain AuthUserCollectionMeta
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"current_page", "from", "last_page", "per_page", "to", "total"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = AuthUserCollectionMeta(out)
	return nil
}
func (v AuthUserCollectionMeta) MarshalJSON() ([]byte, error) {
	type plain AuthUserCollectionMeta
	return marshalUnknown(plain(v), v.Extra)
}

type AuthUserCollectionLinks struct {
	First Optional[string]           `json:"first,omitzero"`
	Last  Optional[string]           `json:"last,omitzero"`
	Prev  Optional[string]           `json:"prev,omitzero"`
	Next  Optional[string]           `json:"next,omitzero"`
	Extra map[string]json.RawMessage `json:"-"`
}

func (v *AuthUserCollectionLinks) UnmarshalJSON(b []byte) error {
	type plain AuthUserCollectionLinks
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"first", "last", "prev", "next"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = AuthUserCollectionLinks(out)
	return nil
}
func (v AuthUserCollectionLinks) MarshalJSON() ([]byte, error) {
	type plain AuthUserCollectionLinks
	return marshalUnknown(plain(v), v.Extra)
}

type EmailVerificationResourceData struct {
	EmailVerified bool `json:"email_verified"`
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

type ChoiceClassificationRequestChoicesItem struct {
	Label       string `json:"label"`
	Description string `json:"description"`
}
