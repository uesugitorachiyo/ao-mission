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
	Schema          string               `json:"schema"`
	MissionID       string               `json:"mission_id"`
	Objective       string               `json:"objective"`
	ObjectiveDigest string               `json:"objective_digest"`
	Status          string               `json:"status"`
	CreatedAtUTC    string               `json:"created_at_utc"`
	UpdatedAtUTC    string               `json:"updated_at_utc"`
	CurrentRoute    string               `json:"current_route"`
	CurrentPhase    string               `json:"current_phase"`
	Blockers        []string             `json:"blockers"`
	ExactNextAction string               `json:"exact_next_action"`
	ArtifactRefs    []ArtifactRef        `json:"artifact_refs"`
	Steps           []ContinuationStep   `json:"steps"`
	RouteHistory    []RouteDecision      `json:"route_history,omitempty"`
	Evidence        EvidenceSummary      `json:"evidence,omitempty"`
	GoalLease       *GoalLease           `json:"goal_lease,omitempty"`
	Checkpoints     []MissionCheckpoint  `json:"checkpoints,omitempty"`
	ReturnGate      *ReturnGate          `json:"return_gate,omitempty"`
	Reconciliation  *RouteReconciliation `json:"route_reconciliation,omitempty"`
}

type EvidenceSummary struct {
	AtlasWorkgraph      *NodeCounts                        `json:"atlas_workgraph,omitempty"`
	AtlasRecommendation *AtlasRecommendationReadbackCounts `json:"atlas_recommendation,omitempty"`
	AtlasFinalSynthesis *AtlasFinalSynthesisReadbackCounts `json:"atlas_final_synthesis,omitempty"`
	FoundryRollup       *FoundryRollupCounts               `json:"foundry_rollup,omitempty"`
	SchedulerReadback   *SchedulerEvidenceCounts           `json:"scheduler_readback,omitempty"`
	SchedulerRecovery   *SchedulerRecoveryCounts           `json:"scheduler_recovery,omitempty"`
	LedgerCompaction    *LedgerCompactionCounts            `json:"ledger_compaction,omitempty"`
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

type AtlasRecommendationReadbackCounts struct {
	Status               string `json:"status"`
	TotalNodes           int    `json:"total_nodes"`
	CompletedNodes       int    `json:"completed_nodes"`
	ReadyNodes           int    `json:"ready_nodes"`
	CheckpointCount      int    `json:"checkpoint_count"`
	ElapsedMinutes       int    `json:"elapsed_minutes"`
	MinMinutesMet        bool   `json:"min_minutes_met"`
	LeaseTimeStatus      string `json:"lease_time_status"`
	ReturnGateStatus     string `json:"return_gate_status"`
	FinalResponseAllowed bool   `json:"final_response_allowed"`
	Blocker              string `json:"blocker,omitempty"`
	RSIRemainsDenied     bool   `json:"rsi_remains_denied,omitempty"`
	ExactNextAction      string `json:"exact_next_action"`
}

type AtlasFinalSynthesisReadbackCounts struct {
	ContractVersion      string `json:"contract_version"`
	Status               string `json:"status"`
	TotalNodes           int    `json:"total_nodes"`
	CompletedNodes       int    `json:"completed_nodes"`
	ReadyNodes           int    `json:"ready_nodes"`
	BlockedNodes         int    `json:"blocked_nodes"`
	MinimumNodes         int    `json:"minimum_nodes"`
	ReturnGateStatus     string `json:"return_gate_status"`
	FinalResponseAllowed bool   `json:"final_response_allowed"`
	FinalResponseReason  string `json:"final_response_reason"`
	AtlasWorkgraphStatus string `json:"atlas_workgraph_status"`
	FoundryRollup        string `json:"foundry_rollup"`
	PromoterStatus       string `json:"promoter_status"`
	CommandReadback      string `json:"command_readback"`
	EventSearchBound     bool   `json:"event_search_bound"`
	BranchCleanupBound   bool   `json:"branch_cleanup_bound"`
	RSIRemainsDenied     bool   `json:"rsi_remains_denied"`
	ExactNextAction      string `json:"exact_next_action"`
}

type GoalLease struct {
	Schema           string `json:"schema"`
	MinNodes         int    `json:"min_nodes"`
	MinMinutes       int    `json:"min_minutes"`
	MaxMinutes       int    `json:"max_minutes"`
	MaxIterations    int    `json:"max_iterations"`
	ReturnOnlyWhen   string `json:"return_only_when"`
	CheckpointPolicy string `json:"checkpoint_policy"`
	CreatedAtUTC     string `json:"created_at_utc"`
	UpdatedAtUTC     string `json:"updated_at_utc"`
}

type MissionCheckpoint struct {
	Schema          string `json:"schema"`
	MissionID       string `json:"mission_id"`
	Sequence        int    `json:"sequence"`
	Iteration       int    `json:"iteration"`
	Route           string `json:"route"`
	Phase           string `json:"phase"`
	Result          string `json:"result"`
	ExactNextAction string `json:"exact_next_action"`
	ResumeCommand   string `json:"resume_command"`
	GeneratedAtUTC  string `json:"generated_at_utc"`
}

type ReturnGate struct {
	Schema               string   `json:"schema"`
	MissionID            string   `json:"mission_id"`
	Status               string   `json:"status"`
	FinalResponseAllowed bool     `json:"final_response_allowed"`
	Reason               string   `json:"reason"`
	CompletedNodes       int      `json:"completed_nodes"`
	MinNodes             int      `json:"min_nodes"`
	ReadyNodesRemaining  int      `json:"ready_nodes_remaining"`
	HardBlocker          bool     `json:"hard_blocker"`
	ExactNextAction      string   `json:"exact_next_action"`
	Blockers             []string `json:"blockers,omitempty"`
	GeneratedAtUTC       string   `json:"generated_at_utc"`
}

type RouteReconciliation struct {
	Schema                string `json:"schema"`
	MissionID             string `json:"mission_id"`
	Status                string `json:"status"`
	CurrentRoute          string `json:"current_route"`
	LatestRoute           string `json:"latest_route"`
	FoundryTerminalStatus string `json:"foundry_terminal_status,omitempty"`
	AtlasReadyNodes       int    `json:"atlas_ready_nodes"`
	CommandReadbackBound  bool   `json:"command_readback_bound"`
	PromoterReadbackBound bool   `json:"promoter_readback_bound"`
	ExactNextAction       string `json:"exact_next_action"`
	GeneratedAtUTC        string `json:"generated_at_utc"`
}

type FeatureDepthRecommendation struct {
	ID              string `json:"id"`
	Owner           string `json:"owner"`
	Task            string `json:"task"`
	ExactNextAction string `json:"exact_next_action"`
}

type MissionCheckpointBundle struct {
	Schema              string             `json:"schema"`
	MissionID           string             `json:"mission_id"`
	Status              string             `json:"status"`
	CheckpointCount     int                `json:"checkpoint_count"`
	LatestCheckpoint    *MissionCheckpoint `json:"latest_checkpoint,omitempty"`
	ReturnGate          *ReturnGate        `json:"return_gate,omitempty"`
	ResumePrompt        string             `json:"resume_prompt"`
	SafeToExecute       bool               `json:"safe_to_execute"`
	ExecutesWork        bool               `json:"executes_work"`
	ApprovesWork        bool               `json:"approves_work"`
	MutatesRepositories bool               `json:"mutates_repositories"`
	GeneratedAtUTC      string             `json:"generated_at_utc"`
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
	SSERequested       bool   `json:"sse_requested,omitempty"`
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
	Schema              string                             `json:"schema"`
	MissionID           string                             `json:"mission_id"`
	Status              string                             `json:"status"`
	CurrentRoute        string                             `json:"current_route"`
	CurrentPhase        string                             `json:"current_phase"`
	ExactNextAction     string                             `json:"exact_next_action"`
	ReadOnly            bool                               `json:"read_only"`
	SafeToExecute       bool                               `json:"safe_to_execute"`
	ExecutesWork        bool                               `json:"executes_work"`
	ApprovesWork        bool                               `json:"approves_work"`
	MutatesRepositories bool                               `json:"mutates_repositories"`
	AtlasRecommendation *AtlasRecommendationReadbackCounts `json:"atlas_recommendation,omitempty"`
	Blockers            []string                           `json:"blockers"`
	GeneratedAtUTC      string                             `json:"generated_at_utc"`
}

type MissionFinalReconciliationPacket struct {
	Schema                    string `json:"schema"`
	MissionID                 string `json:"mission_id"`
	Status                    string `json:"status"`
	ArtifactsAgree            bool   `json:"artifacts_agree"`
	MissionStatus             string `json:"mission_status"`
	AtlasRecommendationStatus string `json:"atlas_recommendation_status,omitempty"`
	FoundryStatus             string `json:"foundry_status,omitempty"`
	CommandStatus             string `json:"command_status"`
	CompletedNodes            int    `json:"completed_nodes"`
	TotalNodes                int    `json:"total_nodes"`
	ReadyNodes                int    `json:"ready_nodes"`
	FinalResponseAllowed      bool   `json:"final_response_allowed"`
	ReturnGateStatus          string `json:"return_gate_status"`
	Blocker                   string `json:"blocker,omitempty"`
	PromotionClaimed          bool   `json:"promotion_claimed"`
	RSIRemainsDenied          bool   `json:"rsi_remains_denied"`
	ClaimsAuthorityAdvance    bool   `json:"claims_authority_advance"`
	SafeToExecute             bool   `json:"safe_to_execute"`
	ExecutesWork              bool   `json:"executes_work"`
	ApprovesWork              bool   `json:"approves_work"`
	MutatesRepositories       bool   `json:"mutates_repositories"`
	GeneratedAtUTC            string `json:"generated_at_utc"`
}

type AtlasWaveFinalSynthesis struct {
	Schema                                string                       `json:"schema"`
	Mission                               string                       `json:"mission"`
	Status                                string                       `json:"status"`
	MissionID                             string                       `json:"mission_id"`
	CompletedNodes                        int                          `json:"completed_nodes"`
	ReadyNodes                            int                          `json:"ready_nodes"`
	BlockedNodes                          int                          `json:"blocked_nodes"`
	MinimumNodes                          int                          `json:"minimum_nodes"`
	TargetMinutes                         int                          `json:"target_minutes"`
	MaxMinutes                            int                          `json:"max_minutes"`
	FinalResponseAllowed                  bool                         `json:"final_response_allowed"`
	AtlasWorkgraphStatus                  string                       `json:"atlas_workgraph_status"`
	FoundryRollup                         string                       `json:"foundry_rollup"`
	PromoterStatus                        string                       `json:"promoter_status"`
	CommandReadback                       string                       `json:"command_readback"`
	EventSearchBound                      bool                         `json:"event_search_bound"`
	BranchCleanupBoundThroughPreviousNode bool                         `json:"branch_cleanup_bound_through_previous_node"`
	MergedPRsFinal                        []int                        `json:"merged_prs_final,omitempty"`
	CurrentNodeBranch                     string                       `json:"current_node_branch"`
	CurrentNodePRPending                  bool                         `json:"current_node_pr_pending"`
	PromotionClaimed                      bool                         `json:"promotion_claimed"`
	ClaimsAuthorityAdvance                bool                         `json:"claims_authority_advance"`
	RSIRemainsDenied                      bool                         `json:"rsi_remains_denied"`
	SafeToExecute                         bool                         `json:"safe_to_execute"`
	ExecutesWork                          bool                         `json:"executes_work"`
	ApprovesWork                          bool                         `json:"approves_work"`
	MutatesRepositories                   bool                         `json:"mutates_repositories"`
	FeatureDepthRecommendations           []FeatureDepthRecommendation `json:"feature_depth_recommendations"`
	ExactNextAction                       string                       `json:"exact_next_action"`
	GeneratedAtUTC                        string                       `json:"generated_at_utc"`
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

type MissionEvent struct {
	Schema         string `json:"schema"`
	MissionID      string `json:"mission_id"`
	Kind           string `json:"kind"`
	Sequence       int    `json:"sequence"`
	Status         string `json:"status,omitempty"`
	Route          string `json:"route,omitempty"`
	Phase          string `json:"phase,omitempty"`
	ArtifactKind   string `json:"artifact_kind,omitempty"`
	Summary        string `json:"summary"`
	GeneratedAtUTC string `json:"generated_at_utc,omitempty"`
}

type MissionEventIndex struct {
	Schema              string         `json:"schema"`
	Status              string         `json:"status"`
	IndexVersion        string         `json:"index_version"`
	IndexDigest         string         `json:"index_digest"`
	SourceDigest        string         `json:"source_digest"`
	Root                string         `json:"root"`
	MissionCount        int            `json:"mission_count"`
	TotalEvents         int            `json:"total_events"`
	Events              []MissionEvent `json:"events"`
	SafeToExecute       bool           `json:"safe_to_execute"`
	ExecutesWork        bool           `json:"executes_work"`
	ApprovesWork        bool           `json:"approves_work"`
	MutatesRepositories bool           `json:"mutates_repositories"`
	GeneratedAtUTC      string         `json:"generated_at_utc"`
}

type MissionEventSearchReadback struct {
	Schema              string         `json:"schema"`
	Status              string         `json:"status"`
	Query               string         `json:"query,omitempty"`
	MissionID           string         `json:"mission_id,omitempty"`
	Kind                string         `json:"kind,omitempty"`
	TotalMatches        int            `json:"total_matches"`
	Events              []MissionEvent `json:"events"`
	SafeToExecute       bool           `json:"safe_to_execute"`
	ExecutesWork        bool           `json:"executes_work"`
	ApprovesWork        bool           `json:"approves_work"`
	MutatesRepositories bool           `json:"mutates_repositories"`
	GeneratedAtUTC      string         `json:"generated_at_utc"`
}

type MissionDoctorReadback struct {
	Schema              string   `json:"schema"`
	Status              string   `json:"status"`
	Root                string   `json:"root"`
	MissionCount        int      `json:"mission_count"`
	EventCount          int      `json:"event_count"`
	LeaseCount          int      `json:"lease_count"`
	FreshCheckpoints    int      `json:"fresh_checkpoints"`
	EarlyReturnRisks    int      `json:"early_return_risks"`
	StaleRoutes         int      `json:"stale_routes"`
	Checks              []string `json:"checks"`
	Blockers            []string `json:"blockers"`
	SafeToExecute       bool     `json:"safe_to_execute"`
	ExecutesWork        bool     `json:"executes_work"`
	ApprovesWork        bool     `json:"approves_work"`
	MutatesRepositories bool     `json:"mutates_repositories"`
	GeneratedAtUTC      string   `json:"generated_at_utc"`
}

type MissionReadinessBundleInput struct {
	Repo string
	Path string
}

type MissionReadinessRepoReadback struct {
	Repo   string `json:"repo"`
	Path   string `json:"path"`
	Status string `json:"status"`
	Score  string `json:"score,omitempty"`
	SHA256 string `json:"sha256"`
}

type MissionReadinessBundleReadback struct {
	Schema              string                         `json:"schema"`
	Status              string                         `json:"status"`
	RepoCount           int                            `json:"repo_count"`
	ReadyRepos          int                            `json:"ready_repos"`
	BlockedRepos        int                            `json:"blocked_repos"`
	Repos               []MissionReadinessRepoReadback `json:"repos"`
	SafeToExecute       bool                           `json:"safe_to_execute"`
	ExecutesWork        bool                           `json:"executes_work"`
	ApprovesWork        bool                           `json:"approves_work"`
	MutatesRepositories bool                           `json:"mutates_repositories"`
	ExactNextAction     string                         `json:"exact_next_action"`
	GeneratedAtUTC      string                         `json:"generated_at_utc"`
}

type GatewayReplayBundleInputs struct {
	TelegramConfigPath  string
	TelegramMatrixPath  string
	TelegramUpdatesPath string
	TelegramWebhookPath string
	A2AHTTPPath         string
	A2ALifecyclePath    string
	SchedulerPath       string
}

type GatewayReplayBundleReadback struct {
	Schema              string   `json:"schema"`
	Status              string   `json:"status"`
	TelegramReadbacks   int      `json:"telegram_readbacks"`
	A2AReadbacks        int      `json:"a2a_readbacks"`
	SchedulerReadbacks  int      `json:"scheduler_readbacks"`
	TotalIntents        int      `json:"total_intents"`
	Denied              int      `json:"denied"`
	Invalid             int      `json:"invalid"`
	ReplayRefs          []string `json:"replay_refs"`
	SafeToExecute       bool     `json:"safe_to_execute"`
	ExecutesWork        bool     `json:"executes_work"`
	ApprovesWork        bool     `json:"approves_work"`
	MutatesRepositories bool     `json:"mutates_repositories"`
	ExactNextAction     string   `json:"exact_next_action"`
	GeneratedAtUTC      string   `json:"generated_at_utc"`
}

type MissionDashboardReadback struct {
	Schema              string         `json:"schema"`
	Status              string         `json:"status"`
	MissionID           string         `json:"mission_id"`
	MissionStatus       string         `json:"mission_status"`
	CurrentPhase        string         `json:"current_phase"`
	CurrentRoute        string         `json:"current_route"`
	LatestRoute         string         `json:"latest_route"`
	EventCount          int            `json:"event_count"`
	EventIndexDigest    string         `json:"event_index_digest"`
	Compact             bool           `json:"compact"`
	RecentEvents        []MissionEvent `json:"recent_events"`
	SafeToExecute       bool           `json:"safe_to_execute"`
	ExecutesWork        bool           `json:"executes_work"`
	ApprovesWork        bool           `json:"approves_work"`
	MutatesRepositories bool           `json:"mutates_repositories"`
	ExactNextAction     string         `json:"exact_next_action"`
	GeneratedAtUTC      string         `json:"generated_at_utc"`
}

type MissionVerificationBundleOptions struct {
	ReadinessBundlePath     string
	GatewayReplayBundlePath string
}

type MissionVerificationBundleComponent struct {
	Name   string `json:"name"`
	Schema string `json:"schema"`
	Path   string `json:"path,omitempty"`
	Status string `json:"status"`
	SHA256 string `json:"sha256"`
}

type MissionVerificationBundleReadback struct {
	Schema              string                               `json:"schema"`
	Status              string                               `json:"status"`
	MissionID           string                               `json:"mission_id"`
	ComponentCount      int                                  `json:"component_count"`
	Components          []MissionVerificationBundleComponent `json:"components"`
	BundleDigest        string                               `json:"bundle_digest"`
	SafeToExecute       bool                                 `json:"safe_to_execute"`
	ExecutesWork        bool                                 `json:"executes_work"`
	ApprovesWork        bool                                 `json:"approves_work"`
	MutatesRepositories bool                                 `json:"mutates_repositories"`
	ExactNextAction     string                               `json:"exact_next_action"`
	GeneratedAtUTC      string                               `json:"generated_at_utc"`
}

func now(clock func() time.Time) string {
	if clock == nil {
		clock = time.Now
	}
	return clock().UTC().Format(time.RFC3339)
}
