package envai

import "encoding/json"

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
	OperationID string                     `json:"operation_id"`
	WorkspaceID string                     `json:"workspace_id"`
	Kind        string                     `json:"kind"`
	Status      string                     `json:"status"`
	Phase       *string                    `json:"phase"`
	Phases      WikiOperationPhases        `json:"phases"`
	StartedAt   *string                    `json:"started_at"`
	FinishedAt  *string                    `json:"finished_at"`
	BatchSizes  BatchSizes                 `json:"batch_sizes"`
	Billing     BillingState               `json:"billing"`
	UsageCost   UsageCost                  `json:"usage_cost"`
	Extra       map[string]json.RawMessage `json:"-"`
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
	extra, err := unknownFields(b, []string{"operation_id", "workspace_id", "kind", "status", "phase", "phases", "started_at", "finished_at", "batch_sizes", "billing", "usage_cost"})
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
}

type ChatMessageRequest struct {
	Message string `json:"message"`
}

type ChatMessageAccepted struct {
	Status string `json:"status"`
	TurnID string `json:"turn_id"`
}

type TurnAccepted struct {
	TurnID string `json:"turn_id"`
}

type ToolOutputRequest struct {
	CallID  string         `json:"call_id"`
	Output  string         `json:"output"`
	IsError Optional[bool] `json:"is_error,omitzero"`
}

type ApprovalRequest struct {
	ApprovalID string           `json:"approval_id"`
	Approve    bool             `json:"approve"`
	Reason     Optional[string] `json:"reason,omitzero"`
}

type SteerRequest struct {
	Text          string           `json:"text"`
	TargetAgentID Optional[string] `json:"target_agent_id,omitzero"`
}

type ForkChatRequest struct {
	DestinationChatID string `json:"destination_chat_id"`
}

type ChatForked struct {
	SourceChatID      string `json:"source_chat_id"`
	DestinationChatID string `json:"destination_chat_id"`
}

type OrchestratorContextItem struct {
	Kind    string           `json:"kind"`
	TurnID  Optional[string] `json:"turn_id,omitzero"`
	Content string           `json:"content"`
}

type RestoreChatRequest struct {
	OrchestratorContext []OrchestratorContextItem `json:"orchestrator_context"`
}

type RestoreChatAccepted struct {
	Status string  `json:"status"`
	TurnID *string `json:"turn_id"`
}

type Capabilities struct {
	ProtocolVersion           int64    `json:"protocol_version"`
	Families                  []string `json:"families"`
	EventTypes                []string `json:"event_types"`
	ToolKinds                 []string `json:"tool_kinds"`
	Features                  []string `json:"features"`
	ContractVersion           int64    `json:"contract_version"`
	SupportedContractVersions []int64  `json:"supported_contract_versions"`
	LegacyV2Mode              string   `json:"legacy_v2_mode"`
}

type Workspace struct {
	ID        string            `json:"id"`
	UserID    string            `json:"user_id"`
	Name      string            `json:"name"`
	Metadata  WorkspaceMetadata `json:"metadata"`
	CreatedAt string            `json:"created_at"`
	UpdatedAt string            `json:"updated_at"`
}

type WorkspaceResource struct {
	Data Workspace `json:"data"`
}

type WorkspaceCollection struct {
	Data []Workspace `json:"data"`
}

type WorkspaceMetadata map[string]json.RawMessage

type WorkspaceCreateRequest struct {
	Name     string                      `json:"name"`
	Metadata Optional[WorkspaceMetadata] `json:"metadata,omitzero"`
}

type WorkspaceUpdateRequest struct {
	Name     Optional[string]            `json:"name,omitzero"`
	Metadata Optional[WorkspaceMetadata] `json:"metadata,omitzero"`
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
	Data UserTool `json:"data"`
}

type UserToolCollection struct {
	Data []UserTool `json:"data"`
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
	ID          string  `json:"id"`
	WorkspaceID string  `json:"workspace_id"`
	UserID      string  `json:"user_id"`
	TeamID      string  `json:"team_id"`
	Title       *string `json:"title"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type ChatResource struct {
	Data Chat `json:"data"`
}

type ChatCollection struct {
	Data []Chat `json:"data"`
}

type ChatCreateRequest struct {
	TeamID string `json:"team_id"`
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
	Data Completion `json:"data"`
}

type CompletionCollection struct {
	Data  []Completion              `json:"data"`
	Links CompletionCollectionLinks `json:"links"`
	Meta  CompletionCollectionMeta  `json:"meta"`
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
}

type AgentResource struct {
	Data Agent `json:"data"`
}

type AgentCollection struct {
	Data []Agent `json:"data"`
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
}

type AgentTool map[string]json.RawMessage

type Team struct {
	ID             string       `json:"id"`
	UserID         string       `json:"user_id"`
	Name           string       `json:"name"`
	Description    *string      `json:"description"`
	OrchestratorID string       `json:"orchestrator_id"`
	Members        []TeamMember `json:"members"`
	CreatedAt      string       `json:"created_at"`
	UpdatedAt      string       `json:"updated_at"`
}

type TeamResource struct {
	Data Team `json:"data"`
}

type TeamCollection struct {
	Data []Team `json:"data"`
}

type TeamCreateRequest struct {
	Name           string                              `json:"name"`
	Description    Optional[string]                    `json:"description,omitzero"`
	OrchestratorID string                              `json:"orchestrator_id"`
	Members        Optional[[]TeamMemberCreateRequest] `json:"members,omitzero"`
}

type TeamUpdateRequest struct {
	Name           Optional[string]                    `json:"name,omitzero"`
	Description    Optional[string]                    `json:"description,omitzero"`
	OrchestratorID Optional[string]                    `json:"orchestrator_id,omitzero"`
	Members        Optional[[]TeamMemberCreateRequest] `json:"members,omitzero"`
}

type TeamMember struct {
	AgentDefinitionID string  `json:"agent_definition_id"`
	Name              string  `json:"name"`
	Position          int64   `json:"position"`
	Role              *string `json:"role"`
}

type TeamMemberCreateRequest struct {
	AgentDefinitionID string           `json:"agent_definition_id"`
	Role              Optional[string] `json:"role,omitzero"`
}

type APIKey struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Scopes    []string `json:"scopes"`
	ExpiresAt *string  `json:"expires_at"`
	RevokedAt *string  `json:"revoked_at"`
	CreatedAt string   `json:"created_at"`
}

type APIKeyCreated struct {
	Data json.RawMessage `json:"data"`
}

type APIKeyCollection struct {
	Data []APIKey `json:"data"`
}

type APIKeyCreate struct {
	Name      string             `json:"name"`
	Scopes    Optional[[]string] `json:"scopes,omitzero"`
	ExpiresAt Optional[string]   `json:"expires_at,omitzero"`
}

type JWKSDocument struct {
	Keys []JWK `json:"keys"`
}

type JWK struct {
	Kty string `json:"kty"`
	Crv string `json:"crv"`
	X   string `json:"x"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
}

type ModelList struct {
	Object   string           `json:"object"`
	Data     []ModelListEntry `json:"data"`
	Defaults ModelDefaults    `json:"defaults"`
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
}

type OpenRouterPricing map[string]json.RawMessage

type OpenRouterArchitecture struct {
	Modality         Optional[string]   `json:"modality,omitzero"`
	InputModalities  Optional[[]string] `json:"input_modalities,omitzero"`
	OutputModalities Optional[[]string] `json:"output_modalities,omitzero"`
	Tokenizer        Optional[string]   `json:"tokenizer,omitzero"`
	InstructType     Optional[string]   `json:"instruct_type,omitzero"`
}

type OpenRouterAliasTarget struct {
	Name Optional[string] `json:"name,omitzero"`
	Slug Optional[string] `json:"slug,omitzero"`
}

type TeamModelEntry struct {
	ID         string          `json:"id"`
	Object     string          `json:"object"`
	Type       string          `json:"type"`
	OwnedBy    string          `json:"owned_by"`
	TeamID     string          `json:"team_id"`
	Created    Optional[int64] `json:"created,omitzero"`
	Deprecated bool            `json:"deprecated"`
	SunsetAt   *string         `json:"sunset_at"`
}

type ModelDefaults struct {
	Orchestrator string `json:"orchestrator"`
	Subagent     string `json:"subagent"`
}

type TestImageGenerateRequest struct {
	Files     []string `json:"files"`
	Prompt    string   `json:"prompt"`
	ModelCode string   `json:"model_code"`
}

type TestImageFileMetadata struct {
	OriginalName string `json:"original_name"`
	MimeType     string `json:"mime_type"`
	Size         int64  `json:"size"`
}

type TestImageGenerateResource struct {
	Data TestImageGenerateResourceData `json:"data"`
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
}

type ImageGenerationLifecycleStatus string

type ImageGenerationAcceptedResource struct {
	Data ImageGenerationAcceptedRecord `json:"data"`
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
}

type ImageGenerationStatusResource struct {
	Data ImageGenerationStatusRecord `json:"data"`
}

type ImagePrice struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
	Source   string `json:"source"`
}

type ImageUsage struct {
	SchemaVersion  int64 `json:"schema_version"`
	AttemptsTotal  int64 `json:"attempts_total"`
	AttemptsPriced int64 `json:"attempts_priced"`
}

type ImageUploadRecord struct {
	ID               string `json:"id"`
	MimeType         string `json:"mime_type"`
	Size             int64  `json:"size"`
	OriginalFilename string `json:"original_filename"`
	ExpiresAt        string `json:"expires_at"`
	CreatedAt        string `json:"created_at"`
}

type ImageUploadResource struct {
	Data ImageUploadRecord `json:"data"`
}

type ImageUploadRequest struct {
	File string `json:"file"`
}

type ModelCatalogEntry struct {
	Model         string            `json:"model"`
	Providers     []string          `json:"providers"`
	PriceClasses  []ModelPriceClass `json:"price_classes"`
	ModelSettings []ModelSetting    `json:"model_settings"`
}

type ModelPriceClass struct {
	Key            string                      `json:"key"`
	Label          string                      `json:"label"`
	ProvidersCount int64                       `json:"providers_count"`
	AveragePrice   ModelPriceClassAveragePrice `json:"average_price"`
}

type ModelSetting struct {
	Key                string                               `json:"key"`
	Type               string                               `json:"type"`
	Providers          []string                             `json:"providers"`
	Default            Optional[json.RawMessage]            `json:"default,omitzero"`
	Enum               Optional[[]string]                   `json:"enum,omitzero"`
	Range              Optional[[]int64]                    `json:"range,omitzero"`
	DefaultsByProvider Optional[map[string]json.RawMessage] `json:"defaults_by_provider,omitzero"`
}

type ModelCatalogResource struct {
	Data []ModelCatalogEntry `json:"data"`
}

type ProviderWebhookPayload map[string]json.RawMessage

type WebhookOk struct {
	Status string `json:"status"`
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
}

type WorkspaceFileEntry struct {
	Path       string  `json:"path"`
	Name       string  `json:"name"`
	Type       string  `json:"type"`
	Size       int64   `json:"size"`
	ModifiedAt *string `json:"modified_at"`
	MimeType   *string `json:"mime_type"`
}

type WorkspaceFileResource struct {
	Data WorkspaceFile `json:"data"`
}

type WorkspaceFileRequest struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type ConvertToPDFRequest struct {
	Path string `json:"path"`
}

type ThrottleError struct {
	ErrorEnvelope ThrottleErrorErrorEnvelope `json:"error"`
}

type CapabilitiesLegacy struct {
	ProtocolVersion int64    `json:"protocol_version"`
	Families        []string `json:"families"`
	EventTypes      []string `json:"event_types"`
	ToolKinds       []string `json:"tool_kinds"`
	Features        []string `json:"features"`
}

type CreateWikiSyncRunBody struct {
	WebhookURL string `json:"webhook_url"`
}

type CreateWikiOutboxRunBody struct {
	WebhookURL string `json:"webhook_url"`
}

type CreateWikiInboxRunBody struct {
	WebhookURL string `json:"webhook_url"`
}

type WithdrawWikiSourceBody struct {
	FileID string `json:"file_id"`
}

type TestGenerateImageResult struct {
	Status     string                  `json:"status"`
	Prompt     string                  `json:"prompt"`
	ModelCode  string                  `json:"model_code"`
	FilesCount int64                   `json:"files_count"`
	Files      []TestImageFileMetadata `json:"files"`
	ReceivedAt string                  `json:"received_at"`
}

type HealthCheckResult struct {
	Status    Optional[string] `json:"status,omitzero"`
	Timestamp Optional[string] `json:"timestamp,omitzero"`
	App       Optional[string] `json:"app,omitzero"`
	Env       Optional[string] `json:"env,omitzero"`
}

type ThrottleErrorErrorEnvelope struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type ModelPriceClassAveragePrice struct {
	Amount   *string `json:"amount"`
	Currency *string `json:"currency"`
}

type ImageGenerationStatusRecordFilesItem struct {
	ArtifactID  string  `json:"artifact_id"`
	MimeType    string  `json:"mime_type"`
	Size        int64   `json:"size"`
	DownloadURL *string `json:"download_url"`
}

type ImageGenerationStatusRecordErrorEnvelope struct {
	Code    Optional[string]                     `json:"code,omitzero"`
	Message Optional[string]                     `json:"message,omitzero"`
	Meta    Optional[map[string]json.RawMessage] `json:"meta,omitzero"`
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
}

type ImageGenerationStatusRecordResultDimensions struct {
	Width  Optional[int64] `json:"width,omitzero"`
	Height Optional[int64] `json:"height,omitzero"`
}

type ImageGenerationStatusRecordResultImagesItem struct {
	ProviderAssetID Optional[string] `json:"provider_asset_id,omitzero"`
	MimeType        Optional[string] `json:"mime_type,omitzero"`
	SourceURL       Optional[string] `json:"source_url,omitzero"`
	Format          Optional[string] `json:"format,omitzero"`
}

type TestImageGenerateResourceData struct {
	Status     string                  `json:"status"`
	Prompt     string                  `json:"prompt"`
	ModelCode  string                  `json:"model_code"`
	FilesCount int64                   `json:"files_count"`
	Files      []TestImageFileMetadata `json:"files"`
	ReceivedAt string                  `json:"received_at"`
}

type CompletionCollectionMeta struct {
	CurrentPage Optional[int64] `json:"current_page,omitzero"`
	From        Optional[int64] `json:"from,omitzero"`
	LastPage    Optional[int64] `json:"last_page,omitzero"`
	PerPage     Optional[int64] `json:"per_page,omitzero"`
	To          Optional[int64] `json:"to,omitzero"`
	Total       Optional[int64] `json:"total,omitzero"`
}

type CompletionCollectionLinks struct {
	First Optional[string] `json:"first,omitzero"`
	Last  Optional[string] `json:"last,omitzero"`
	Prev  Optional[string] `json:"prev,omitzero"`
	Next  Optional[string] `json:"next,omitzero"`
}

type CompletionUsage struct {
	PromptTokens     Optional[int64] `json:"prompt_tokens,omitzero"`
	CompletionTokens Optional[int64] `json:"completion_tokens,omitzero"`
	SpriteCpuMs      Optional[int64] `json:"sprite_cpu_ms,omitzero"`
	SpriteRamGbMs    Optional[int64] `json:"sprite_ram_gb_ms,omitzero"`
}

type ErrorEnvelopeErrorEnvelope struct {
	Type    string                               `json:"type"`
	Code    Optional[string]                     `json:"code,omitzero"`
	Message string                               `json:"message"`
	Details Optional[map[string]json.RawMessage] `json:"details,omitzero"`
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
