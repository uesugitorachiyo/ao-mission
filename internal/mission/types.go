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
	Evidence        EvidenceSummary    `json:"evidence,omitempty"`
}

type EvidenceSummary struct {
	AtlasWorkgraph    *NodeCounts              `json:"atlas_workgraph,omitempty"`
	FoundryRollup     *FoundryRollupCounts     `json:"foundry_rollup,omitempty"`
	SchedulerReadback *SchedulerEvidenceCounts `json:"scheduler_readback,omitempty"`
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
	Status       string `json:"status"`
	Scheduler    string `json:"scheduler"`
	EventLoop    bool   `json:"event_loop"`
	ExecutesWork bool   `json:"executes_work"`
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
type A2AAgentCard struct {
	Schema            string   `json:"schema"`
	Name              string   `json:"name"`
	Methods           []string `json:"methods"`
	MutationAuthority bool     `json:"mutation_authority"`
}
type A2ATask struct {
	Schema            string `json:"schema"`
	TaskID            string `json:"task_id"`
	Method            string `json:"method"`
	Status            string `json:"status"`
	MutationAuthority bool   `json:"mutation_authority"`
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

func now(clock func() time.Time) string {
	if clock == nil {
		clock = time.Now
	}
	return clock().UTC().Format(time.RFC3339)
}
