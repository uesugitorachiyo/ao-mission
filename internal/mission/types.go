package mission

import "time"

const (
	RecordSchema            = "ao.mission.record.v0.1"
	IntentSchema            = "ao.mission.operator-intent.v0.1"
	ResultSchema            = "ao.mission.operator-result.v0.1"
	RouteSchema             = "ao.mission.route-decision.v0.1"
	SnapshotSchema          = "ao.mission.governance-snapshot.v0.1"
	StepSchema              = "ao.mission.continuation-step.v0.1"
	SchedulerRequestSchema  = "ao.mission.scheduler-request.v0.1"
	SchedulerReadbackSchema = "ao.mission.scheduler-readback.v0.1"
	EventLoopDecisionSchema = "ao.mission.event-loop-decision.v0.1"
	KillSwitchSchema        = "ao.mission.kill-switch.v0.1"
	TelegramCommandSchema   = "ao.mission.telegram-command.v0.1"
	TelegramReadbackSchema  = "ao.mission.telegram-readback.v0.1"
	A2AAgentCardSchema      = "ao.mission.a2a-agent-card.v0.1"
	A2ATaskSchema           = "ao.mission.a2a-task.v0.1"
	ArtifactRefSchema       = "ao.mission.artifact-ref.v0.1"
	ErrorSchema             = "ao.mission.error.v0.1"
)

type ArtifactRef struct {
	Schema string `json:"schema"`
	Ref    string `json:"ref"`
	Digest string `json:"digest,omitempty"`
	Kind   string `json:"kind,omitempty"`
}

type Record struct {
	Schema          string             `json:"schema"`
	MissionID       string             `json:"mission_id"`
	Objective       string             `json:"objective"`
	ObjectiveDigest string             `json:"objective_digest"`
	Status          string             `json:"status"`
	CreatedAtUTC    string             `json:"created_at_utc"`
	UpdatedAtUTC    string             `json:"updated_at_utc"`
	CurrentRoute    string             `json:"current_route"`
	CurrentPhase    string             `json:"current_phase"`
	Blockers        []string           `json:"blockers"`
	ExactNextAction string             `json:"exact_next_action"`
	ArtifactRefs    []ArtifactRef      `json:"artifact_refs"`
	Steps           []ContinuationStep `json:"steps"`
	RouteHistory    []RouteDecision    `json:"route_history,omitempty"`
	Evidence        EvidenceSummary    `json:"evidence,omitempty"`
}

type EvidenceSummary struct {
	AtlasWorkgraph    *NodeCounts              `json:"atlas_workgraph,omitempty"`
	FoundryRollup     *FoundryRollupCounts     `json:"foundry_rollup,omitempty"`
	SchedulerReadback *SchedulerEvidenceCounts `json:"scheduler_readback,omitempty"`
	SchedulerRecovery *SchedulerRecoveryCounts `json:"scheduler_recovery,omitempty"`
	LedgerCompaction  *LedgerCompactionCounts  `json:"ledger_compaction,omitempty"`
}

type NodeCounts struct {
	Total     int `json:"total"`
	Ready     int `json:"ready"`
	Blocked   int `json:"blocked"`
	Completed int `json:"completed"`
	Failed    int `json:"failed"`
}

type FoundryRollupCounts struct {
	Status         string `json:"status"`
	CompletedNodes int    `json:"completed_nodes"`
	TotalNodes     int    `json:"total_nodes"`
}

type SchedulerEvidenceCounts struct {
	Status          string `json:"status"`
	Scheduler       string `json:"scheduler"`
	EventLoop       bool   `json:"event_loop"`
	FreshnessStatus string `json:"freshness_status"`
	ExecutesWork    bool   `json:"executes_work"`
}

type SchedulerRecoveryCounts struct {
	Status        string `json:"status"`
	RecoveryMode  string `json:"recovery_mode"`
	MissedWakeups int    `json:"missed_wakeups"`
	ExecutesWork  bool   `json:"executes_work"`
}

type RouteDecision struct {
	Schema          string `json:"schema"`
	MissionID       string `json:"mission_id"`
	Route           string `json:"route"`
	Reason          string `json:"reason"`
	SafeToRequest   bool   `json:"safe_to_request"`
	SafeToExecute   bool   `json:"safe_to_execute"`
	SafeToPromote   bool   `json:"safe_to_promote"`
	ExactNextAction string `json:"exact_next_action"`
	GeneratedAtUTC  string `json:"generated_at_utc,omitempty"`
}

type ContinuationStep struct {
	Schema          string `json:"schema"`
	MissionID       string `json:"mission_id"`
	Iteration       int    `json:"iteration"`
	Route           string `json:"route"`
	Result          string `json:"result"`
	ExactNextAction string `json:"exact_next_action"`
	GeneratedAtUTC  string `json:"generated_at_utc"`
}

type GovernanceSnapshot struct {
	Schema                  string        `json:"schema"`
	MissionID               string        `json:"mission_id"`
	ObjectiveDigest         string        `json:"objective_digest"`
	CurrentOwner            string        `json:"current_owner"`
	CurrentPhase            string        `json:"current_phase"`
	CurrentRoute            string        `json:"current_route"`
	HighestProvenLiveClass  string        `json:"highest_proven_live_class"`
	NextDeniedClass         string        `json:"next_denied_class"`
	SafeToRequest           bool          `json:"safe_to_request"`
	SafeToExecute           bool          `json:"safe_to_execute"`
	SafeToPromote           bool          `json:"safe_to_promote"`
	SchedulesWork           bool          `json:"schedules_work"`
	ExecutesWork            bool          `json:"executes_work"`
	ApprovesWork            bool          `json:"approves_work"`
	MutatesRepositories     bool          `json:"mutates_repositories"`
	ProviderCalls           bool          `json:"provider_calls"`
	ReleaseOrPublish        bool          `json:"release_or_publish"`
	CredentialUse           bool          `json:"credential_use"`
	DirectMainMutation      bool          `json:"direct_main_mutation"`
	ConcurrentMutation      bool          `json:"concurrent_mutation"`
	SentinelStatus          string        `json:"sentinel_status"`
	PromoterStatus          string        `json:"promoter_status"`
	CovenantStatus          string        `json:"covenant_status"`
	RollbackStatus          string        `json:"rollback_status"`
	CIStatus                string        `json:"ci_status"`
	RepoHygieneStatus       string        `json:"repo_hygiene_status"`
	EvidenceFreshnessStatus string        `json:"evidence_freshness_status"`
	KillSwitchStatus        string        `json:"kill_switch_status"`
	Blockers                []string      `json:"blockers"`
	ExactNextAction         string        `json:"exact_next_action"`
	ArtifactRefs            []ArtifactRef `json:"artifact_refs"`
	GeneratedAtUTC          string        `json:"generated_at_utc"`
}

type SchedulerReadback struct {
	Schema         string `json:"schema"`
	MissionID      string `json:"mission_id"`
	Status         string `json:"status"`
	Scheduler      string `json:"scheduler"`
	EventLoop      bool   `json:"event_loop"`
	Reason         string `json:"reason,omitempty"`
	GeneratedAtUTC string `json:"generated_at_utc"`
}

type EventLoopDecision struct {
	Schema              string `json:"schema"`
	MissionID           string `json:"mission_id"`
	Iteration           int    `json:"iteration"`
	Status              string `json:"status"`
	Route               string `json:"route"`
	ExactNextAction     string `json:"exact_next_action"`
	ExecutesWork        bool   `json:"executes_work"`
	ApprovesWork        bool   `json:"approves_work"`
	MutatesRepositories bool   `json:"mutates_repositories"`
	GeneratedAtUTC      string `json:"generated_at_utc"`
}

type TelegramCommand struct {
	Schema  string `json:"schema"`
	ChatID  string `json:"chat_id"`
	Command string `json:"command"`
	Role    string `json:"role"`
}
type TelegramReadback struct {
	Schema            string `json:"schema"`
	Status            string `json:"status"`
	Message           string `json:"message"`
	MutationAuthority bool   `json:"mutation_authority"`
}
type TelegramCommandMatrix struct {
	Schema   string                       `json:"schema"`
	Commands []TelegramCommandMatrixEntry `json:"commands"`
}
type TelegramCommandMatrixEntry struct {
	Command        string `json:"command"`
	Role           string `json:"role"`
	ExpectedStatus string `json:"expected_status"`
}

type GatewayReplayResult struct {
	RequestID         string `json:"request_id,omitempty"`
	ResponseID        string `json:"response_id,omitempty"`
	Command           string `json:"command,omitempty"`
	Method            string `json:"method,omitempty"`
	ExpectedStatus    string `json:"expected_status"`
	ActualStatus      string `json:"actual_status"`
	MutationAuthority bool   `json:"mutation_authority"`
}

type GatewayReplayReadback struct {
	Schema            string                `json:"schema"`
	Gateway           string                `json:"gateway"`
	Status            string                `json:"status"`
	Total             int                   `json:"total"`
	IntentRecorded    int                   `json:"intent_recorded"`
	Denied            int                   `json:"denied"`
	Invalid           int                   `json:"invalid"`
	Duplicates        int                   `json:"duplicates,omitempty"`
	Fresh             int                   `json:"fresh,omitempty"`
	Stale             int                   `json:"stale,omitempty"`
	UnknownFreshness  int                   `json:"unknown_freshness,omitempty"`
	FreshnessStatus   string                `json:"freshness_status,omitempty"`
	CorrelationID     string                `json:"correlation_id,omitempty"`
	Results           []GatewayReplayResult `json:"results"`
	MutationAuthority bool                  `json:"mutation_authority"`
	ExecutesWork      bool                  `json:"executes_work"`
	ApprovesWork      bool                  `json:"approves_work"`
	GeneratedAtUTC    string                `json:"generated_at_utc"`
}

type GatewayIntentRecord struct {
	Schema            string `json:"schema"`
	MissionID         string `json:"mission_id"`
	Gateway           string `json:"gateway"`
	Command           string `json:"command,omitempty"`
	Method            string `json:"method,omitempty"`
	Status            string `json:"status"`
	ExpectedStatus    string `json:"expected_status,omitempty"`
	MutationAuthority bool   `json:"mutation_authority"`
	ExecutesWork      bool   `json:"executes_work"`
	ApprovesWork      bool   `json:"approves_work"`
	GeneratedAtUTC    string `json:"generated_at_utc"`
}

type GatewayIntentLedger struct {
	Schema            string                `json:"schema"`
	MissionID         string                `json:"mission_id"`
	Status            string                `json:"status"`
	Total             int                   `json:"total"`
	IntentRecorded    int                   `json:"intent_recorded"`
	Denied            int                   `json:"denied"`
	Invalid           int                   `json:"invalid"`
	Intents           []GatewayIntentRecord `json:"intents"`
	MutationAuthority bool                  `json:"mutation_authority"`
	ExecutesWork      bool                  `json:"executes_work"`
	ApprovesWork      bool                  `json:"approves_work"`
	GeneratedAtUTC    string                `json:"generated_at_utc"`
}

type GatewayReplaySuiteReadback struct {
	Schema            string                    `json:"schema"`
	Status            string                    `json:"status"`
	TelegramReplays   int                       `json:"telegram_replays"`
	A2AReplays        int                       `json:"a2a_replays"`
	Total             int                       `json:"total"`
	IntentRecorded    int                       `json:"intent_recorded"`
	Denied            int                       `json:"denied"`
	Invalid           int                       `json:"invalid"`
	ArtifactReadbacks int                       `json:"artifact_readbacks"`
	CorrelationID     string                    `json:"correlation_id,omitempty"`
	ReplayRefs        []string                  `json:"replay_refs"`
	Replays           []GatewayReplayReadback   `json:"replays"`
	A2ALifecycle      *A2ATaskLifecycleReadback `json:"a2a_lifecycle,omitempty"`
	MutationAuthority bool                      `json:"mutation_authority"`
	ExecutesWork      bool                      `json:"executes_work"`
	ApprovesWork      bool                      `json:"approves_work"`
	GeneratedAtUTC    string                    `json:"generated_at_utc"`
}

type A2ATaskLifecycleReadback struct {
	Schema            string    `json:"schema"`
	Status            string    `json:"status"`
	Total             int       `json:"total"`
	IntentRecorded    int       `json:"intent_recorded"`
	CancelRequested   int       `json:"cancel_requested"`
	Cancelled         int       `json:"cancelled"`
	ResumeRequested   int       `json:"resume_requested"`
	Resumed           int       `json:"resumed"`
	ArtifactReadbacks int       `json:"artifact_readbacks"`
	Tasks             []A2ATask `json:"tasks"`
	MutationAuthority bool      `json:"mutation_authority"`
	ExecutesWork      bool      `json:"executes_work"`
	ApprovesWork      bool      `json:"approves_work"`
	GeneratedAtUTC    string    `json:"generated_at_utc"`
}

type A2ACompatibilityReadback struct {
	Schema            string `json:"schema"`
	Status            string `json:"status"`
	ProtocolVersion   string `json:"protocol_version"`
	AgentCardSkills   int    `json:"agent_card_skills"`
	Methods           int    `json:"methods"`
	HTTPRequests      int    `json:"http_requests"`
	LifecycleTasks    int    `json:"lifecycle_tasks"`
	ArtifactReadbacks int    `json:"artifact_readbacks"`
	MutationAuthority bool   `json:"mutation_authority"`
	ExecutesWork      bool   `json:"executes_work"`
	ApprovesWork      bool   `json:"approves_work"`
	GeneratedAtUTC    string `json:"generated_at_utc"`
}

type A2AStreamingDenialReadback struct {
	Schema             string `json:"schema"`
	Status             string `json:"status"`
	StreamingRequested bool   `json:"streaming_requested"`
	PushRequested      bool   `json:"push_notifications_requested"`
	DeniedCapability   string `json:"denied_capability"`
	MutationAuthority  bool   `json:"mutation_authority"`
	ExecutesWork       bool   `json:"executes_work"`
	ApprovesWork       bool   `json:"approves_work"`
	ExactNextAction    string `json:"exact_next_action"`
	GeneratedAtUTC     string `json:"generated_at_utc"`
}

type A2ACancellationReplayReadback struct {
	Schema            string `json:"schema"`
	Status            string `json:"status"`
	Total             int    `json:"total"`
	CancelRequested   int    `json:"cancel_requested"`
	Cancelled         int    `json:"cancelled"`
	MutationAuthority bool   `json:"mutation_authority"`
	ExecutesWork      bool   `json:"executes_work"`
	ApprovesWork      bool   `json:"approves_work"`
	ExactNextAction   string `json:"exact_next_action"`
	GeneratedAtUTC    string `json:"generated_at_utc"`
}

type A2AAgentCard struct {
	Schema             string          `json:"schema"`
	Name               string          `json:"name"`
	ProtocolVersion    string          `json:"protocol_version"`
	Description        string          `json:"description"`
	Endpoint           string          `json:"endpoint"`
	Methods            []string        `json:"methods"`
	Capabilities       []string        `json:"capabilities"`
	CapabilitiesDetail map[string]bool `json:"capabilities_detail,omitempty"`
	Skills             []A2AAgentSkill `json:"skills,omitempty"`
	MutationAuthority  bool            `json:"mutation_authority"`
}
type A2AAgentSkill struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}
type A2ATask struct {
	Schema            string        `json:"schema"`
	TaskID            string        `json:"task_id"`
	Method            string        `json:"method"`
	Status            string        `json:"status"`
	ArtifactRefs      []ArtifactRef `json:"artifact_refs,omitempty"`
	MutationAuthority bool          `json:"mutation_authority"`
}

type A2AJSONRPCResponse struct {
	JSONRPC string  `json:"jsonrpc"`
	ID      any     `json:"id,omitempty"`
	Result  A2ATask `json:"result"`
}

type ArtifactManifest struct {
	Schema         string        `json:"schema"`
	MissionID      string        `json:"mission_id"`
	ArtifactRefs   []ArtifactRef `json:"artifact_refs"`
	ManifestDigest string        `json:"manifest_digest"`
	Signature      string        `json:"signature"`
	SafeToExecute  bool          `json:"safe_to_execute"`
	ExecutesWork   bool          `json:"executes_work"`
	ApprovesWork   bool          `json:"approves_work"`
	GeneratedAtUTC string        `json:"generated_at_utc"`
}

type ArtifactManifestValidation struct {
	Schema         string `json:"schema"`
	Status         string `json:"status"`
	MissionID      string `json:"mission_id"`
	ArtifactCount  int    `json:"artifact_count"`
	ManifestDigest string `json:"manifest_digest"`
	ExecutesWork   bool   `json:"executes_work"`
	ApprovesWork   bool   `json:"approves_work"`
	GeneratedAtUTC string `json:"generated_at_utc"`
}

type SchedulerReplayReadback struct {
	Schema         string `json:"schema"`
	Status         string `json:"status"`
	Total          int    `json:"total"`
	Fresh          int    `json:"fresh"`
	Stale          int    `json:"stale"`
	Unknown        int    `json:"unknown"`
	ExecutesWork   bool   `json:"executes_work"`
	ApprovesWork   bool   `json:"approves_work"`
	EvaluatedAtUTC string `json:"evaluated_at_utc"`
	GeneratedAtUTC string `json:"generated_at_utc"`
}

type SchedulerAlertSummary struct {
	Schema         string   `json:"schema"`
	Status         string   `json:"status"`
	Total          int      `json:"total"`
	Fresh          int      `json:"fresh"`
	Stale          int      `json:"stale"`
	Unknown        int      `json:"unknown"`
	Alerts         []string `json:"alerts"`
	ExecutesWork   bool     `json:"executes_work"`
	ApprovesWork   bool     `json:"approves_work"`
	GeneratedAtUTC string   `json:"generated_at_utc"`
}

type SchedulerRecoveryReadback struct {
	Schema          string `json:"schema"`
	MissionID       string `json:"mission_id"`
	Status          string `json:"status"`
	RecoveryMode    string `json:"recovery_mode"`
	MissedWakeups   int    `json:"missed_wakeups"`
	Fresh           int    `json:"fresh"`
	Stale           int    `json:"stale"`
	Unknown         int    `json:"unknown"`
	ExactNextAction string `json:"exact_next_action"`
	ExecutesWork    bool   `json:"executes_work"`
	ApprovesWork    bool   `json:"approves_work"`
	GeneratedAtUTC  string `json:"generated_at_utc"`
}

type LedgerCompactionOptions struct {
	KeepRouteHistory int
	KeepSteps        int
	DryRun           bool
}

type LedgerCompactionCounts struct {
	RouteHistoryBefore int `json:"route_history_before"`
	RouteHistoryAfter  int `json:"route_history_after"`
	StepsBefore        int `json:"steps_before"`
	StepsAfter         int `json:"steps_after"`
}

type LedgerCompactionReadback struct {
	Schema              string `json:"schema"`
	MissionID           string `json:"mission_id"`
	Status              string `json:"status"`
	RouteHistoryBefore  int    `json:"route_history_before"`
	RouteHistoryAfter   int    `json:"route_history_after"`
	StepsBefore         int    `json:"steps_before"`
	StepsAfter          int    `json:"steps_after"`
	ExactNextAction     string `json:"exact_next_action"`
	ExecutesWork        bool   `json:"executes_work"`
	ApprovesWork        bool   `json:"approves_work"`
	MutatesRepositories bool   `json:"mutates_repositories"`
	GeneratedAtUTC      string `json:"generated_at_utc"`
}

type TimelineCompactionReadback struct {
	Schema              string `json:"schema"`
	MissionID           string `json:"mission_id"`
	Status              string `json:"status"`
	RouteHistoryBefore  int    `json:"route_history_before"`
	RouteHistoryAfter   int    `json:"route_history_after"`
	StepsBefore         int    `json:"steps_before"`
	StepsAfter          int    `json:"steps_after"`
	TimelineDigest      string `json:"timeline_digest"`
	ExactNextAction     string `json:"exact_next_action"`
	ExecutesWork        bool   `json:"executes_work"`
	ApprovesWork        bool   `json:"approves_work"`
	MutatesRepositories bool   `json:"mutates_repositories"`
	GeneratedAtUTC      string `json:"generated_at_utc"`
}

type CommandStatus struct {
	Schema              string   `json:"schema"`
	MissionID           string   `json:"mission_id"`
	Status              string   `json:"status"`
	CurrentRoute        string   `json:"current_route"`
	CurrentPhase        string   `json:"current_phase"`
	ExactNextAction     string   `json:"exact_next_action"`
	ReadOnly            bool     `json:"read_only"`
	SafeToExecute       bool     `json:"safe_to_execute"`
	ExecutesWork        bool     `json:"executes_work"`
	ApprovesWork        bool     `json:"approves_work"`
	MutatesRepositories bool     `json:"mutates_repositories"`
	Blockers            []string `json:"blockers"`
	GeneratedAtUTC      string   `json:"generated_at_utc"`
}

type GovernanceSnapshotDiff struct {
	Schema         string   `json:"schema"`
	Status         string   `json:"status"`
	MissionID      string   `json:"mission_id"`
	ChangedFields  int      `json:"changed_fields"`
	Fields         []string `json:"fields"`
	SafeToExecute  bool     `json:"safe_to_execute"`
	ExecutesWork   bool     `json:"executes_work"`
	ApprovesWork   bool     `json:"approves_work"`
	GeneratedAtUTC string   `json:"generated_at_utc"`
}

type TelegramRoleEntry struct {
	ChatID string `json:"chat_id"`
	Role   string `json:"role"`
}

type TelegramRoleMatrixReadback struct {
	Schema            string              `json:"schema"`
	Status            string              `json:"status"`
	ChatCount         int                 `json:"chat_count"`
	AdminCount        int                 `json:"admin_count"`
	UserCount         int                 `json:"user_count"`
	Roles             []TelegramRoleEntry `json:"roles"`
	MutationAuthority bool                `json:"mutation_authority"`
	ExecutesWork      bool                `json:"executes_work"`
	ApprovesWork      bool                `json:"approves_work"`
	GeneratedAtUTC    string              `json:"generated_at_utc"`
}

type MissionArchive struct {
	Schema         string             `json:"schema"`
	MissionID      string             `json:"mission_id"`
	Record         Record             `json:"record"`
	Snapshot       GovernanceSnapshot `json:"snapshot"`
	FinalRollup    FinalRollup        `json:"final_rollup"`
	ArtifactCount  int                `json:"artifact_count"`
	ArchiveDigest  string             `json:"archive_digest"`
	SafeToExecute  bool               `json:"safe_to_execute"`
	ExecutesWork   bool               `json:"executes_work"`
	ApprovesWork   bool               `json:"approves_work"`
	GeneratedAtUTC string             `json:"generated_at_utc"`
}

type MissionArchiveValidation struct {
	Schema         string `json:"schema"`
	Status         string `json:"status"`
	MissionID      string `json:"mission_id"`
	ArchiveDigest  string `json:"archive_digest"`
	ArtifactCount  int    `json:"artifact_count"`
	SafeToExecute  bool   `json:"safe_to_execute"`
	ExecutesWork   bool   `json:"executes_work"`
	ApprovesWork   bool   `json:"approves_work"`
	GeneratedAtUTC string `json:"generated_at_utc"`
}

type MissionArchiveImportReadback struct {
	Schema         string `json:"schema"`
	Status         string `json:"status"`
	MissionID      string `json:"mission_id"`
	ArchiveDigest  string `json:"archive_digest"`
	SafeToExecute  bool   `json:"safe_to_execute"`
	ExecutesWork   bool   `json:"executes_work"`
	ApprovesWork   bool   `json:"approves_work"`
	GeneratedAtUTC string `json:"generated_at_utc"`
}

type GatewayReadinessRollup struct {
	Schema              string   `json:"schema"`
	MissionID           string   `json:"mission_id,omitempty"`
	Status              string   `json:"status"`
	CorrelationID       string   `json:"correlation_id,omitempty"`
	ReadbackCount       int      `json:"readback_count"`
	BlockedReadbacks    int      `json:"blocked_readbacks"`
	ReadbackRefs        []string `json:"readback_refs"`
	SafeToExecute       bool     `json:"safe_to_execute"`
	ExecutesWork        bool     `json:"executes_work"`
	ApprovesWork        bool     `json:"approves_work"`
	MutatesRepositories bool     `json:"mutates_repositories"`
	ExactNextAction     string   `json:"exact_next_action"`
	GeneratedAtUTC      string   `json:"generated_at_utc"`
}

func now(clock func() time.Time) string {
	if clock == nil {
		clock = time.Now
	}
	return clock().UTC().Format(time.RFC3339)
}
