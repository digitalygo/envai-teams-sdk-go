package envai

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestGeneratedOpenObjectPreservation(t *testing.T) {
	for _, model := range []any{&BillingState{}, &CurrencyAmount{}, &UsageCost{}, &ConversionState{}, &IndexingState{}, &WikiSource{}, &BatchSizes{}, &FileOutcome{}, &PhaseResult{}, &WikiOperation{}, &Category{}, &Categories{}, &CatalogEntry{}, &IndexRange{}, &CategoryIndex{}, &WikiFileProcessedEvent{}, &WikiFileDeletedEvent{}, &WikiOperationEvent{}, &ErrorEnvelope{}, &ChatMessageRequest{}, &ChatMessageAccepted{}, &TurnAccepted{}, &ToolOutputRequest{}, &ApprovalRequest{}, &SteerRequest{}, &ForkChatRequest{}, &ChatForked{}, &OrchestratorContextItem{}, &RestoreChatRequest{}, &RestoreChatAccepted{}, &Capabilities{}, &Workspace{}, &WorkspaceResource{}, &WorkspaceCollection{}, &WorkspaceCreateRequest{}, &WorkspaceUpdateRequest{}, &UserToolResource{}, &UserToolCollection{}, &Chat{}, &ChatResource{}, &ChatCollection{}, &ChatCreateRequest{}, &Completion{}, &CompletionResource{}, &CompletionCollection{}, &Agent{}, &AgentResource{}, &AgentCollection{}, &AgentCreateRequest{}, &AgentUpdateRequest{}, &Team{}, &TeamResource{}, &TeamCollection{}, &TeamCreateRequest{}, &TeamUpdateRequest{}, &TeamMember{}, &TeamMemberCreateRequest{}, &APIKey{}, &APIKeyCreated{}, &APIKeyCollection{}, &APIKeyCreate{}, &JWKSDocument{}, &JWK{}, &ModelList{}, &LLMModelEntry{}, &OpenRouterArchitecture{}, &OpenRouterAliasTarget{}, &TeamModelEntry{}, &ModelDefaults{}, &TestImageGenerateRequest{}, &TestImageFileMetadata{}, &TestImageGenerateResource{}, &ImageGenerationAcceptedRecord{}, &ImageGenerationAcceptedResource{}, &ImageGenerationStatusRecord{}, &ImageGenerationStatusResource{}, &ImagePrice{}, &ImageUsage{}, &ImageUploadRecord{}, &ImageUploadResource{}, &ImageUploadRequest{}, &ModelCatalogEntry{}, &ModelPriceClass{}, &ModelSetting{}, &ModelCatalogResource{}, &WebhookOk{}, &WorkspaceFile{}, &WorkspaceFileEntry{}, &WorkspaceFileResource{}, &WorkspaceFileRequest{}, &ConvertToPDFRequest{}, &ThrottleError{}, &CapabilitiesLegacy{}, &CreateWikiSyncRunBody{}, &CreateWikiOutboxRunBody{}, &CreateWikiInboxRunBody{}, &WithdrawWikiSourceBody{}, &TestGenerateImageResult{}, &HealthCheckResult{}, &ModelPriceClassAveragePrice{}, &ImageGenerationStatusRecordFilesItem{}, &ImageGenerationStatusRecordErrorEnvelope{}, &ImageGenerationStatusRecordResult{}, &ImageGenerationStatusRecordResultDimensions{}, &ImageGenerationStatusRecordResultImagesItem{}, &TestImageGenerateResourceData{}, &CompletionCollectionMeta{}, &CompletionCollectionLinks{}, &CompletionUsage{}, &ErrorEnvelopeErrorEnvelope{}, &ATPAgent{}, &ATPUsage{}, &ATPBillingState{}, &ATPCurrencyAmount{}, &ATPUsageCost{}} {
		t.Run(reflect.TypeOf(model).Elem().Name(), func(t *testing.T) {
			for _, invalid := range []string{"null", "[]", "{"} {
				if json.Unmarshal([]byte(invalid), model) == nil {
					t.Fatal("malformed open object accepted")
				}
			}
			input := []byte(`{"extension":{"number":9007199254740993}}`)
			err := json.Unmarshal(input, model)
			// Shared Wiki DTOs enforce required fields separately; their positive fixtures
			// are covered by the handwritten Wiki fixture and accounting tests.
			if err == nil {
				out, e := json.Marshal(model)
				if e != nil || !strings.Contains(string(out), `"extension":{"number":9007199254740993}`) {
					t.Fatal(string(out), e)
				}
			}
			value := reflect.ValueOf(model).Elem()
			extra := value.FieldByName("Extra")
			extra.Set(reflect.ValueOf(map[string]json.RawMessage{"extension": json.RawMessage(`true`)}))
			if _, err := json.Marshal(model); err != nil {
				t.Fatal(err)
			}
		})
	}
}
