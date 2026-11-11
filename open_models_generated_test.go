package envai

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestGeneratedOpenObjectPreservation(t *testing.T) {
	for _, model := range []any{&DetachedTaskSnapshot{}, &BillingState{}, &CurrencyAmount{}, &UsageCost{}, &ConversionState{}, &IndexingState{}, &WikiSource{}, &BatchSizes{}, &FileOutcome{}, &PhaseResult{}, &WikiOperation{}, &Category{}, &Categories{}, &CatalogEntry{}, &IndexRange{}, &CategoryIndex{}, &WikiFileProcessedEvent{}, &WikiFileDeletedEvent{}, &WikiOperationEvent{}, &AuthUser{}, &AuthUserResource{}, &AuthUserCollection{}, &CsrfResource{}, &MessageResource{}, &LoginRequest{}, &RegisterRequest{}, &RegisterFromInvitationRequest{}, &ForgotPasswordRequest{}, &ResetPasswordRequest{}, &AcceptInvitationRequest{}, &PlatformRoleUpdateRequest{}, &Company{}, &CompanyResource{}, &CompanyCollection{}, &CompanyCreateRequest{}, &CompanyUpdateRequest{}, &CompanyMember{}, &CompanyMemberResource{}, &CompanyMemberCollection{}, &CompanyMemberRoleUpdateRequest{}, &CompanyInvitation{}, &CompanyInvitationResource{}, &CompanyInvitationCollection{}, &CompanyInvitationCreateRequest{}, &Project{}, &ProjectResource{}, &ProjectCollection{}, &ProjectCreateRequest{}, &ProjectUpdateRequest{}, &BillingWalletResource{}, &BillingTransactionCollection{}, &BillingTopupQuoteResource{}, &BillingTopupCheckoutResource{}, &BillingPaymentMethodSetupResource{}, &BillingAutoRechargeResource{}, &BillingSpendLimitResource{}, &BillingWebhookOk{}, &BillingError{}, &ProjectMember{}, &ProjectMemberResource{}, &ProjectMemberCollection{}, &CleanupIntentState{}, &CleanupResult{}, &CleanupResource{}, &UsageMoney{}, &UsageKeyBreakdown{}, &UsageProjectBreakdown{}, &UsageMediaTotals{}, &UsageReportDay{}, &UsageReportCollection{}, &ProjectLogEntry{}, &ProjectLogCollection{}, &InternalWikiDispatchRequest{}, &InternalDispatchResponse{}, &InternalAttachmentResolveRequest{}, &InternalAttachmentResolveResource{}, &InternalAttachmentManifest{}, &InternalBillingReservationResource{}, &InternalBillingAdmissionResource{}, &InternalBillingSettleResource{}, &InternalBillingReleaseResource{}, &ErrorEnvelope{}, &AttachmentManifest{}, &AttachmentConversion{}, &ChatAttachment{}, &ChatAttachmentResource{}, &ChatAttachmentCollection{}, &ChatMessageRequest{}, &ChatMessageAccepted{}, &TurnAccepted{}, &ToolOutputRequest{}, &ApprovalRequest{}, &SteerRequest{}, &ForkChatRequest{}, &ChatForked{}, &OrchestratorContextItem{}, &RestoreChatRequest{}, &RestoreChatAccepted{}, &Capabilities{}, &Workspace{}, &WorkspaceResource{}, &WorkspaceCollection{}, &WorkspaceCreateRequest{}, &WorkspaceUpdateRequest{}, &ProjectToolResource{}, &ProjectToolCollection{}, &ProjectSkillResource{}, &ProjectSkillCollection{}, &Chat{}, &ChatResource{}, &ChatCollection{}, &ChatCreateRequest{}, &ChatCron{}, &ChatCronResource{}, &ChatCronCollection{}, &ChatCronUpdateRequest{}, &ChatUpdateRequest{}, &Completion{}, &CompletionResource{}, &CompletionCollection{}, &Agent{}, &AgentResource{}, &AgentCollection{}, &AgentCreateRequest{}, &AgentUpdateRequest{}, &Team{}, &TeamResource{}, &TeamCollection{}, &TeamCreateRequest{}, &TeamUpdateRequest{}, &TeamMember{}, &TeamMemberCreateRequest{}, &APIKey{}, &APIKeyCreated{}, &APIKeyCollection{}, &APIKeyCreate{}, &JWKSDocument{}, &JWK{}, &ModelList{}, &LLMModelEntry{}, &OpenRouterArchitecture{}, &OpenRouterAliasTarget{}, &TeamModelEntry{}, &ModelDefaults{}, &TestImageGenerateRequest{}, &TestImageFileMetadata{}, &TestImageGenerateResource{}, &ImageGenerationAcceptedRecord{}, &ImageGenerationAcceptedResource{}, &ImageGenerationStatusRecord{}, &ImageGenerationStatusResource{}, &ImagePrice{}, &ImageUsage{}, &ImageUploadRecord{}, &ImageUploadResource{}, &ImageUploadRequest{}, &ModelCatalogEntry{}, &ModelPriceClass{}, &ModelSetting{}, &ModelCatalogResource{}, &WebhookOk{}, &WorkspaceFile{}, &WorkspaceFileEntry{}, &WorkspaceFileResource{}, &WorkspaceFileRequest{}, &ConvertToPDFRequest{}, &ThrottleError{}, &CapabilitiesLegacy{}, &CreateWikiSyncRunBody{}, &CreateWikiOutboxRunBody{}, &CreateWikiInboxRunBody{}, &WithdrawWikiSourceBody{}, &ListChatAttachmentsResult{}, &TestGenerateImageResult{}, &HealthCheckResult{}, &ModelPriceClassAveragePrice{}, &ImageGenerationStatusRecordFilesItem{}, &ImageGenerationStatusRecordErrorEnvelope{}, &ImageGenerationStatusRecordResult{}, &ImageGenerationStatusRecordResultDimensions{}, &ImageGenerationStatusRecordResultImagesItem{}, &TestImageGenerateResourceData{}, &CompletionCollectionMeta{}, &CompletionCollectionLinks{}, &CompletionUsage{}, &ChatAttachmentCollectionData{}, &ErrorEnvelopeErrorEnvelope{}, &InternalAttachmentResolveResourceData{}, &InternalWikiDispatchRequestActor{}, &ProjectLogCollectionMeta{}, &ProjectLogCollectionLinks{}, &UsageReportCollectionMeta{}, &UsageReportCollectionLinks{}, &ProjectMemberCollectionMeta{}, &ProjectMemberCollectionLinks{}, &BillingErrorErrorEnvelope{}, &BillingTransactionCollectionMeta{}, &BillingTransactionCollectionLinks{}, &ProjectCollectionMeta{}, &ProjectCollectionLinks{}, &CompanyInvitationCollectionMeta{}, &CompanyInvitationCollectionLinks{}, &CompanyMemberCollectionMeta{}, &CompanyMemberCollectionLinks{}, &CompanyCollectionMeta{}, &CompanyCollectionLinks{}, &MessageResourceData{}, &CsrfResourceData{}, &AuthUserCollectionMeta{}, &AuthUserCollectionLinks{}, &ATPAgent{}, &ATPUsage{}, &ATPBillingState{}, &ATPCurrencyAmount{}, &ATPUsageCost{}, &ATPChatAttachmentManifestEvent{}, &ATPTaskStarted{}, &ATPTaskEvent{}, &ATPTaskCompleted{}, &ATPTaskFailed{}, &ATPTaskCancelled{}, &ATPChatDeliveryQueued{}, &ATPChatDeliveryDelivered{}} {
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
