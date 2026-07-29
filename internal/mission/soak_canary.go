package mission

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	SoakCanaryAuthoritySchema      = "ao.mission.soak-canary-authority.v1"
	SoakCanaryCommandCatalogSchema = "ao.mission.soak-canary-command-catalog.v1"
	SoakCanaryActivationSchema     = "ao.mission.soak-canary-activation.v1"
	SoakCanaryAttemptSchema        = "ao.mission.soak-canary-attempt.v1"
	SoakCanaryCheckpointSchema     = "ao.mission.soak-canary-checkpoint.v1"
	SoakCanarySummarySchema        = "ao.mission.soak-canary-summary.v1"
	SoakCanaryValidationSchema     = "ao.mission.soak-canary-validation.v1"
	soakCanaryMaxInputBytes        = 1 << 20
	soakCanaryDefaultOutputBytes   = 1 << 20
)

type SoakCanarySafety struct {
	RepositoryMutationAllowed   bool `json:"repository_mutation_allowed"`
	RuntimeProviderCallsAllowed bool `json:"runtime_provider_calls_allowed"`
	PublicationAllowed          bool `json:"publication_allowed"`
	ReleaseAllowed              bool `json:"release_allowed"`
	DeploymentAllowed           bool `json:"deployment_allowed"`
	CredentialChangesAllowed    bool `json:"credential_changes_allowed"`
	PermissionChangesAllowed    bool `json:"permission_changes_allowed"`
	AuthorityExpansionAllowed   bool `json:"authority_expansion_allowed"`
	NetworkAccessAllowed        bool `json:"network_access_allowed"`
}

type SoakCanaryAuthority struct {
	Schema                      string           `json:"schema"`
	CampaignID                  string           `json:"campaign_id"`
	CanaryID                    string           `json:"canary_id"`
	MissionID                   string           `json:"mission_id"`
	HandoffPath                 string           `json:"handoff_path"`
	HandoffSHA256               string           `json:"handoff_sha256"`
	AuthorityClass              string           `json:"authority_class"`
	EvidenceRoot                string           `json:"evidence_root"`
	SourceHead                  string           `json:"source_head"`
	CreatedAtUTC                string           `json:"created_at_utc"`
	HardWallMS                  int64            `json:"hard_wall_ms"`
	MaximumPlannedNodes         int              `json:"maximum_planned_nodes"`
	MaximumAttempts             int              `json:"maximum_attempts"`
	MaximumChildProcessLaunches int              `json:"maximum_child_process_launches"`
	MaximumScaleLaunches        int              `json:"maximum_scale_launches"`
	MaximumRetryCount           int              `json:"maximum_retry_count"`
	LocalTestExecutionAllowed   bool             `json:"local_test_execution_allowed"`
	Safety                      SoakCanarySafety `json:"safety"`
	AuthorityRecordDigest       string           `json:"authority_record_digest"`
}

type SoakCanaryEnvironment struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type SoakCanaryCommand struct {
	TestID               string                  `json:"test_id"`
	TestName             string                  `json:"test_name"`
	ExecutablePath       string                  `json:"executable_path"`
	ExecutableSHA256     string                  `json:"executable_sha256"`
	Argv                 []string                `json:"argv"`
	WorkingDirectory     string                  `json:"working_directory"`
	Package              string                  `json:"package"`
	TestRegex            string                  `json:"test_regex"`
	Classification       string                  `json:"classification"`
	RequestedRepeatCount int                     `json:"requested_repeat_count"`
	EffectiveRepeatCount int                     `json:"effective_repeat_count"`
	TimeoutMS            int64                   `json:"timeout_ms"`
	Race                 bool                    `json:"race"`
	OutputFormat         string                  `json:"output_format"`
	Environment          []SoakCanaryEnvironment `json:"environment"`
}

type SoakCanaryCommandCatalog struct {
	Schema                 string              `json:"schema"`
	SourceHead             string              `json:"source_head"`
	ExecutionProfileID     string              `json:"execution_profile_id"`
	ExecutionProfileDigest string              `json:"execution_profile_digest"`
	Commands               []SoakCanaryCommand `json:"commands"`
	CommandCatalogDigest   string              `json:"command_catalog_digest"`
}

type SoakCanaryPartitionBinding struct {
	PartitionID          string              `json:"partition_id"`
	NodeID               string              `json:"node_id"`
	TestID               string              `json:"test_id"`
	Classification       string              `json:"classification"`
	RequestedRepeatCount int                 `json:"requested_repeat_count"`
	EffectiveRepeatCount int                 `json:"effective_repeat_count"`
	ScaleDimension       *SoakScaleDimension `json:"scale_dimension,omitempty"`
	EstimatedDurationMS  int64               `json:"estimated_duration_ms"`
	RetryAllowanceMS     int64               `json:"retry_allowance_ms"`
	PerAttemptTimeoutMS  int64               `json:"per_attempt_timeout_ms"`
	TotalNodeTimeoutMS   int64               `json:"total_node_timeout_ms"`
	NodeBudgetMS         int64               `json:"node_budget_ms"`
}

type SoakCanaryActivation struct {
	Schema                                string                       `json:"schema"`
	CanaryID                              string                       `json:"canary_id"`
	MissionID                             string                       `json:"mission_id"`
	PlanID                                string                       `json:"plan_id"`
	PlanFixtureSHA256                     string                       `json:"plan_fixture_sha256"`
	PlanInputDigest                       string                       `json:"plan_input_digest"`
	PolicyDigest                          string                       `json:"policy_digest"`
	SourceHead                            string                       `json:"source_head"`
	ExecutionProfileID                    string                       `json:"execution_profile_id"`
	ExecutionProfileDigest                string                       `json:"execution_profile_digest"`
	CommandCatalogDigest                  string                       `json:"command_catalog_digest"`
	AuthorityRecordDigest                 string                       `json:"authority_record_digest"`
	ActivationManifestDigest              string                       `json:"activation_manifest_digest"`
	ActivationState                       string                       `json:"activation_state"`
	PlannerActivationEligible             bool                         `json:"planner_activation_eligible"`
	CanaryExecutionAuthorized             bool                         `json:"canary_execution_authorized"`
	PolicyBoundBeforeActivation           bool                         `json:"policy_bound_before_activation"`
	ClassificationBoundBeforePartitioning bool                         `json:"classification_bound_before_partitioning"`
	PhaseStartUTC                         string                       `json:"phase_start_utc"`
	Partitions                            []SoakCanaryPartitionBinding `json:"partitions"`
	ControlledRetryNodeID                 string                       `json:"controlled_retry_node_id"`
	ControlledRetryReason                 string                       `json:"controlled_retry_reason"`
	Safety                                SoakCanarySafety             `json:"safety"`
}

type SoakCanaryOutputArtifact struct {
	Path      string `json:"path"`
	SHA256    string `json:"sha256"`
	Bytes     int    `json:"bytes"`
	Truncated bool   `json:"truncated"`
}

type SoakCanaryGoTestCounts struct {
	TotalEvents    int `json:"total_events"`
	MatchingPasses int `json:"matching_passes"`
}

type SoakCanaryAttempt struct {
	Schema                   string                   `json:"schema"`
	CanaryID                 string                   `json:"canary_id"`
	MissionID                string                   `json:"mission_id"`
	PlanID                   string                   `json:"plan_id"`
	PartitionID              string                   `json:"partition_id"`
	NodeID                   string                   `json:"node_id"`
	TestID                   string                   `json:"test_id"`
	AttemptNumber            int                      `json:"attempt_number"`
	PhaseStartUTC            string                   `json:"phase_start_utc"`
	SourceHead               string                   `json:"source_head"`
	PlanInputDigest          string                   `json:"plan_input_digest"`
	PolicyDigest             string                   `json:"policy_digest"`
	ExecutionProfileDigest   string                   `json:"execution_profile_digest"`
	CommandCatalogDigest     string                   `json:"command_catalog_digest"`
	AuthorityRecordDigest    string                   `json:"authority_record_digest"`
	ActivationManifestDigest string                   `json:"activation_manifest_digest"`
	CommandArgvDigest        string                   `json:"command_argv_digest"`
	RequestedRepeatCount     int                      `json:"requested_repeat_count"`
	EffectiveRepeatCount     int                      `json:"effective_repeat_count"`
	Classification           string                   `json:"classification"`
	ScaleDimension           *SoakScaleDimension      `json:"scale_dimension,omitempty"`
	OutcomeClass             string                   `json:"outcome_class"`
	ChildProcessLaunched     bool                     `json:"child_process_launched"`
	ChildPID                 int                      `json:"child_pid,omitempty"`
	StartedAtUTC             string                   `json:"started_at_utc"`
	CompletedAtUTC           string                   `json:"completed_at_utc"`
	ElapsedMS                int64                    `json:"elapsed_ms"`
	ExitCode                 int                      `json:"exit_code"`
	Signal                   string                   `json:"signal,omitempty"`
	Stdout                   SoakCanaryOutputArtifact `json:"stdout"`
	Stderr                   SoakCanaryOutputArtifact `json:"stderr"`
	GoTestEvents             SoakCanaryGoTestCounts   `json:"go_test_events"`
	CheckpointBeforeDigest   string                   `json:"checkpoint_before_digest"`
	CheckpointAfterSequence  int                      `json:"checkpoint_after_sequence"`
	Safety                   SoakCanarySafety         `json:"safety"`
	AttemptDigest            string                   `json:"attempt_digest"`
}

type SoakCanaryCheckpoint struct {
	Schema                   string              `json:"schema"`
	CanaryID                 string              `json:"canary_id"`
	MissionID                string              `json:"mission_id"`
	PlanID                   string              `json:"plan_id"`
	PhaseStartUTC            string              `json:"phase_start_utc"`
	CompletedAtUTC           string              `json:"completed_at_utc,omitempty"`
	SourceHead               string              `json:"source_head"`
	PlanInputDigest          string              `json:"plan_input_digest"`
	PolicyDigest             string              `json:"policy_digest"`
	CommandCatalogDigest     string              `json:"command_catalog_digest"`
	AuthorityRecordDigest    string              `json:"authority_record_digest"`
	ActivationManifestDigest string              `json:"activation_manifest_digest"`
	Sequence                 int                 `json:"sequence"`
	PriorCheckpointDigest    string              `json:"prior_checkpoint_digest"`
	Attempts                 []SoakCanaryAttempt `json:"attempts"`
	CompletedNodeIDs         []string            `json:"completed_node_ids"`
	ControlledRetryConsumed  bool                `json:"controlled_retry_consumed"`
	ScaleLaunchConsumed      bool                `json:"scale_launch_consumed"`
	CheckpointDigest         string              `json:"checkpoint_digest"`
}

type SoakCanarySummary struct {
	Schema                      string              `json:"schema"`
	Status                      string              `json:"status"`
	CanaryID                    string              `json:"canary_id"`
	MissionID                   string              `json:"mission_id"`
	PlanID                      string              `json:"plan_id"`
	SourceHead                  string              `json:"source_head"`
	PlanInputDigest             string              `json:"plan_input_digest"`
	PolicyDigest                string              `json:"policy_digest"`
	CommandCatalogDigest        string              `json:"command_catalog_digest"`
	AuthorityRecordDigest       string              `json:"authority_record_digest"`
	ActivationManifestDigest    string              `json:"activation_manifest_digest"`
	PlannedPartitions           int                 `json:"planned_partitions"`
	PlannedNodes                int                 `json:"planned_nodes"`
	CompletedNodes              int                 `json:"completed_nodes"`
	TotalAttempts               int                 `json:"total_attempts"`
	ChildProcessLaunches        int                 `json:"child_process_launches"`
	ScaleLaunches               int                 `json:"scale_launches"`
	ControlledRetryCount        int                 `json:"controlled_retry_count"`
	LocalTestExecutionPerformed bool                `json:"local_test_execution_performed"`
	PhaseStartUTC               string              `json:"phase_start_utc"`
	CompletedAtUTC              string              `json:"completed_at_utc"`
	PhaseElapsedMS              int64               `json:"phase_elapsed_ms"`
	LeaseMaximumMS              int64               `json:"lease_maximum_ms"`
	TerminalIndexDigest         string              `json:"terminal_index_digest"`
	ConflictCodes               []string            `json:"conflict_codes"`
	Attempts                    []SoakCanaryAttempt `json:"attempts"`
	Safety                      SoakCanarySafety    `json:"safety"`
	SummaryDigest               string              `json:"summary_digest"`
}

type SoakCanaryActivationReadback struct {
	Schema                    string   `json:"schema"`
	ActivationAllowed         bool     `json:"activation_allowed"`
	PlannerActivationEligible bool     `json:"planner_activation_eligible"`
	CanaryExecutionAuthorized bool     `json:"canary_execution_authorized"`
	ChildProcessLaunches      int      `json:"child_process_launches"`
	ConflictCodes             []string `json:"conflict_codes"`
	ExactNextAction           string   `json:"exact_next_action"`
}

type SoakCanaryRunRequest struct {
	PlanInput          SoakPlanInput
	PlanReadback       SoakPlanReadback
	PlanFixtureSHA256  string
	Authority          SoakCanaryAuthority
	Catalog            SoakCanaryCommandCatalog
	Activation         SoakCanaryActivation
	VerifiedSourceHead string
	RepositoryRoot     string
	EvidenceRoot       string
	CheckpointPath     string
	OutputLimitBytes   int
	Executor           SoakCanaryExecutor
	Clock              SoakCanaryClock
}

type soakCanaryApprovedTest struct {
	name           string
	regex          string
	classification string
}

var soakCanaryApprovedTests = map[string]soakCanaryApprovedTest{
	"scale-event-index-10000": {
		name:           "TestMissionEventIndexScaleMetricsExposeReadAndEventCounts/10000",
		regex:          "^TestMissionEventIndexScaleMetricsExposeReadAndEventCounts$/^10000$",
		classification: "scale",
	},
	"regular-issue-reader-special-files": {
		name:           "TestIssueRepairRequestReaderRejectsSymlinkAndFIFO",
		regex:          "^TestIssueRepairRequestReaderRejectsSymlinkAndFIFO$",
		classification: "regular",
	},
	"regular-correlated-import-binding": {
		name:           "TestCorrelatedImportRequiresArtifactRoleAndDigestWithoutMutation",
		regex:          "^TestCorrelatedImportRequiresArtifactRoleAndDigestWithoutMutation$",
		classification: "regular",
	},
	"regular-archive-duplicate-authority": {
		name:           "TestMissionArchiveEntryPointsRejectDuplicateAuthorityAtAnyDepth",
		regex:          "^TestMissionArchiveEntryPointsRejectDuplicateAuthorityAtAnyDepth$",
		classification: "regular",
	},
	"regular-checkpoint-doctor": {
		name:           "TestContinueWritesCheckpointBundleAndDoctorSupervisorHealth",
		regex:          "^TestContinueWritesCheckpointBundleAndDoctorSupervisorHealth$",
		classification: "regular",
	},
	"regular-final-reconciliation-shape": {
		name:           "TestFinalReconciliationEventSearchFixturePreservesReadbackShape",
		regex:          "^TestFinalReconciliationEventSearchFixturePreservesReadbackShape$",
		classification: "regular",
	},
	"regular-final-rollup-ready-denial": {
		name:           "TestFinalRollupReadyNodeDenialFixtureValidatesSchema",
		regex:          "^TestFinalRollupReadyNodeDenialFixtureValidatesSchema$",
		classification: "regular",
	},
	"regular-ledger-compaction": {
		name:           "TestMissionLedgerCompactionTrimsHistoryAndRecordsEvidence",
		regex:          "^TestMissionLedgerCompactionTrimsHistoryAndRecordsEvidence$",
		classification: "regular",
	},
	"regular-dashboard-corrupt-unrelated": {
		name:           "TestMissionDashboardUsesOneMissionReadPathWithCorruptUnrelatedRecord",
		regex:          "^TestMissionDashboardUsesOneMissionReadPathWithCorruptUnrelatedRecord$",
		classification: "regular",
	},
	"regular-objective-pending-blueprint": {
		name:           "TestObjectiveWorkflowRoutesPendingBlueprint",
		regex:          "^TestObjectiveWorkflowRoutesPendingBlueprint$",
		classification: "regular",
	},
}

func LoadSoakCanaryAuthority(path string) (SoakCanaryAuthority, error) {
	return loadSoakCanaryJSON[SoakCanaryAuthority](path)
}

func LoadSoakCanaryCommandCatalog(path string) (SoakCanaryCommandCatalog, error) {
	return loadSoakCanaryJSON[SoakCanaryCommandCatalog](path)
}

func LoadSoakCanaryActivation(path string) (SoakCanaryActivation, error) {
	return loadSoakCanaryJSON[SoakCanaryActivation](path)
}

func loadSoakCanaryJSON[T any](path string) (T, error) {
	var value T
	body, err := readBoundedRegularFile(path, soakCanaryMaxInputBytes)
	if err != nil {
		return value, err
	}
	return decodeSoakCanaryJSON[T](body)
}

func decodeSoakCanaryJSON[T any](body []byte) (T, error) {
	var value T
	if err := validateNoDuplicateJSONKeys(body); err != nil {
		if strings.Contains(err.Error(), "duplicate JSON key") {
			return value, err
		}
		return value, fmt.Errorf("invalid JSON: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return value, fmt.Errorf("invalid JSON: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return value, errors.New("soak canary input contains trailing JSON")
		}
		return value, fmt.Errorf("invalid JSON: %w", err)
	}
	return value, nil
}

func BuildSoakCanaryActivation(
	input SoakPlanInput,
	plan SoakPlanReadback,
	planFixtureSHA256 string,
	authority SoakCanaryAuthority,
	catalog SoakCanaryCommandCatalog,
	phaseStartUTC, controlledRetryNodeID string,
) (SoakCanaryActivation, error) {
	if _, err := time.Parse(time.RFC3339, phaseStartUTC); err != nil {
		return SoakCanaryActivation{}, errors.New("soak canary phase start must be RFC3339")
	}
	activation := SoakCanaryActivation{
		Schema:   SoakCanaryActivationSchema,
		CanaryID: authority.CanaryID, MissionID: input.MissionID, PlanID: input.PlanID,
		PlanFixtureSHA256: planFixtureSHA256, PlanInputDigest: plan.InputDigest,
		PolicyDigest: plan.PolicyDigest, SourceHead: input.SourceHead,
		ExecutionProfileID:                    input.ExecutionProfile.ID,
		ExecutionProfileDigest:                input.ExecutionProfile.Digest,
		CommandCatalogDigest:                  catalog.CommandCatalogDigest,
		AuthorityRecordDigest:                 authority.AuthorityRecordDigest,
		ActivationState:                       "activated",
		PlannerActivationEligible:             plan.ActivationAllowed,
		CanaryExecutionAuthorized:             authority.LocalTestExecutionAllowed,
		PolicyBoundBeforeActivation:           input.Activation.PolicyBoundBeforeActivation,
		ClassificationBoundBeforePartitioning: input.ClassificationBoundBeforePartitioning,
		PhaseStartUTC:                         phaseStartUTC,
		ControlledRetryNodeID:                 controlledRetryNodeID,
		ControlledRetryReason:                 "transient_infrastructure",
		Safety:                                authority.Safety,
	}
	scaleDimensions := map[string]*SoakScaleDimension{}
	for _, test := range input.TestCatalog {
		if test.ScaleDimension != nil {
			copyDimension := *test.ScaleDimension
			scaleDimensions[test.ID] = &copyDimension
		}
	}
	for _, partition := range plan.Partitions {
		if len(partition.Tests) != 1 {
			return SoakCanaryActivation{}, errors.New("soak canary requires exactly one test per planned partition")
		}
		activation.Partitions = append(activation.Partitions, SoakCanaryPartitionBinding{
			PartitionID: partition.PartitionID, NodeID: partition.NodeID,
			TestID: partition.Tests[0], Classification: partition.Classification,
			RequestedRepeatCount: partition.RequestedRepeatCount,
			EffectiveRepeatCount: partition.EffectiveRepeatCount,
			ScaleDimension:       scaleDimensions[partition.Tests[0]],
			EstimatedDurationMS:  partition.EstimatedDurationMS,
			RetryAllowanceMS:     partition.RetryAllowanceMS,
			PerAttemptTimeoutMS:  input.TimeoutPolicy.PerAttemptTimeoutMS,
			TotalNodeTimeoutMS:   input.TimeoutPolicy.TotalNodeTimeoutMS,
			NodeBudgetMS:         partition.NodeBudgetMS,
		})
	}
	signSoakCanaryActivation(&activation)
	return activation, nil
}

func ValidateSoakCanaryActivation(request SoakCanaryRunRequest) SoakCanaryActivationReadback {
	conflicts := map[string]bool{}
	add := func(code string) { conflicts[code] = true }

	recomputedPlan, err := BuildSoakPlan(request.PlanInput)
	if err != nil || !reflect.DeepEqual(recomputedPlan, request.PlanReadback) {
		add("planner_readback_mismatch")
	}
	if !request.PlanReadback.ActivationAllowed {
		add("planner_activation_not_allowed")
	}
	if request.PlanReadback.InputDigest != request.Activation.PlanInputDigest {
		add("plan_input_digest_mismatch")
	}
	if request.PlanReadback.PolicyDigest != request.Activation.PolicyDigest ||
		request.PlanInput.PolicyDigest != request.Activation.PolicyDigest {
		add("policy_digest_mismatch")
	}
	if request.PlanFixtureSHA256 != request.Activation.PlanFixtureSHA256 ||
		!validSoakHexDigest(request.PlanFixtureSHA256, 71, "sha256:") {
		add("plan_fixture_digest_mismatch")
	}
	validateSoakCanaryAuthority(request, add)
	validateSoakCanaryCatalog(request, add)
	validateSoakCanaryActivationRecord(request, add)
	validateSoakCanaryPaths(request, add)

	codes := make([]string, 0, len(conflicts))
	for code := range conflicts {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	readback := SoakCanaryActivationReadback{
		Schema:                    SoakCanaryValidationSchema,
		ActivationAllowed:         len(codes) == 0,
		PlannerActivationEligible: request.PlanReadback.ActivationAllowed,
		CanaryExecutionAuthorized: request.Authority.LocalTestExecutionAllowed && !soakCanarySafetyEnabled(request.Authority.Safety),
		ChildProcessLaunches:      0,
		ConflictCodes:             codes,
	}
	if readback.ActivationAllowed {
		readback.ExactNextAction = "Execute only the digest-bound bounded local soak canary."
	} else {
		readback.ExactNextAction = "Correct the listed conflicts; no child process is authorized."
	}
	return readback
}

func validateSoakCanaryAuthority(request SoakCanaryRunRequest, add func(string)) {
	authority := request.Authority
	unsigned := authority
	signSoakCanaryAuthority(&unsigned)
	if authority.Schema != SoakCanaryAuthoritySchema ||
		authority.AuthorityRecordDigest != unsigned.AuthorityRecordDigest {
		add("authority_record_digest_mismatch")
	}
	if authority.CanaryID == "" || authority.MissionID != request.PlanInput.MissionID ||
		authority.SourceHead != request.PlanInput.SourceHead ||
		authority.AuthorityClass != "bounded_local_operational_canary" ||
		!filepath.IsAbs(authority.HandoffPath) ||
		!validSoakHexDigest(authority.HandoffSHA256, 71, "sha256:") ||
		!filepath.IsAbs(authority.EvidenceRoot) {
		add("authority_identity_mismatch")
	}
	handoffBody, handoffErr := readBoundedRegularFile(authority.HandoffPath, soakCanaryMaxInputBytes)
	if handoffErr != nil || digestBytes(handoffBody) != authority.HandoffSHA256 {
		add("handoff_digest_mismatch")
	}
	if _, err := time.Parse(time.RFC3339, authority.CreatedAtUTC); err != nil ||
		authority.HardWallMS != 45*60*1000 {
		add("authority_time_window_invalid")
	}
	if authority.MaximumPlannedNodes != 10 || authority.MaximumAttempts != 11 ||
		authority.MaximumChildProcessLaunches != 10 || authority.MaximumScaleLaunches != 1 ||
		authority.MaximumRetryCount != 1 || !authority.LocalTestExecutionAllowed {
		add("authority_scope_broadened")
	}
	if soakCanarySafetyEnabled(authority.Safety) {
		add("unsafe_authority_boundary")
	}
}

func validateSoakCanaryCatalog(request SoakCanaryRunRequest, add func(string)) {
	catalog := request.Catalog
	unsigned := catalog
	signSoakCanaryCommandCatalog(&unsigned)
	if catalog.Schema != SoakCanaryCommandCatalogSchema ||
		catalog.CommandCatalogDigest != unsigned.CommandCatalogDigest {
		add("command_catalog_digest_mismatch")
	}
	if catalog.SourceHead != request.PlanInput.SourceHead ||
		catalog.ExecutionProfileID != request.PlanInput.ExecutionProfile.ID ||
		catalog.ExecutionProfileDigest != request.PlanInput.ExecutionProfile.Digest {
		add("execution_profile_digest_mismatch")
	}
	if len(catalog.Commands) != len(soakCanaryApprovedTests) {
		add("command_catalog_count_mismatch")
	}
	seen := map[string]bool{}
	for _, command := range catalog.Commands {
		approved, exists := soakCanaryApprovedTests[command.TestID]
		if !exists || seen[command.TestID] {
			add("unplanned_test_id")
			continue
		}
		seen[command.TestID] = true
		if command.TestName != approved.name || command.Classification != approved.classification {
			add("command_test_identity_mismatch")
		}
		if command.Package != "./internal/mission" || len(command.Argv) < 2 || command.Argv[1] != "./internal/mission" {
			add("unapproved_package")
		}
		if command.WorkingDirectory != "." || filepath.IsAbs(command.WorkingDirectory) ||
			filepath.Clean(command.WorkingDirectory) != command.WorkingDirectory {
			add("working_directory_traversal")
		}
		if command.TestRegex != approved.regex || len(command.Argv) < 5 || command.Argv[4] != approved.regex {
			add("unanchored_test_regex")
		}
		if !validSoakCanaryExecutable(command) {
			add("unapproved_executable")
		}
		expectedRepeat := 3
		if command.Classification == "scale" {
			expectedRepeat = 1
		}
		if command.EffectiveRepeatCount != expectedRepeat ||
			command.RequestedRepeatCount != expectedRepeat {
			if command.Classification == "scale" {
				add("scale_repeat_above_one")
			} else {
				add("regular_repeat_above_approved")
			}
		}
		expectedArgv := []string{
			"test", "./internal/mission", "-race", "-run", approved.regex,
			"-count=" + strconv.Itoa(expectedRepeat), "-json", "-timeout",
			strconv.FormatInt(command.TimeoutMS, 10) + "ms",
		}
		if !reflect.DeepEqual(command.Argv, expectedArgv) || soakCanaryArgvHasShell(command.Argv) {
			add("free_form_shell_command")
		}
		if !command.Race || command.OutputFormat != "go-test-json" ||
			command.TimeoutMS <= 0 || !validSoakCanaryEnvironment(command.Environment) {
			add("environment_injection")
		}
		if command.TimeoutMS != request.PlanInput.TimeoutPolicy.PerAttemptTimeoutMS {
			add("command_timeout_mismatch")
		}
	}
}

func validateSoakCanaryActivationRecord(request SoakCanaryRunRequest, add func(string)) {
	activation := request.Activation
	unsigned := activation
	signSoakCanaryActivation(&unsigned)
	if activation.Schema != SoakCanaryActivationSchema ||
		activation.ActivationManifestDigest != unsigned.ActivationManifestDigest {
		add("activation_manifest_digest_mismatch")
	}
	if activation.CanaryID != request.Authority.CanaryID ||
		activation.MissionID != request.PlanInput.MissionID ||
		activation.PlanID != request.PlanInput.PlanID ||
		activation.ActivationState != "activated" {
		add("activation_identity_mismatch")
	}
	if activation.SourceHead != request.PlanInput.SourceHead ||
		request.VerifiedSourceHead != request.PlanInput.SourceHead {
		add("source_head_mismatch")
	}
	if activation.ExecutionProfileID != request.PlanInput.ExecutionProfile.ID ||
		activation.ExecutionProfileDigest != request.PlanInput.ExecutionProfile.Digest {
		add("execution_profile_digest_mismatch")
	}
	if activation.CommandCatalogDigest != request.Catalog.CommandCatalogDigest {
		add("command_catalog_digest_mismatch")
	}
	if activation.AuthorityRecordDigest != request.Authority.AuthorityRecordDigest {
		add("authority_record_digest_mismatch")
	}
	if !activation.PlannerActivationEligible || !activation.CanaryExecutionAuthorized ||
		!activation.PolicyBoundBeforeActivation ||
		!activation.ClassificationBoundBeforePartitioning {
		add("activation_predates_binding")
	}
	if _, err := time.Parse(time.RFC3339, activation.PhaseStartUTC); err != nil {
		add("phase_start_invalid")
	}
	if activation.PhaseStartUTC != request.Authority.CreatedAtUTC {
		add("phase_clock_reset")
	}
	if soakCanarySafetyEnabled(activation.Safety) ||
		!reflect.DeepEqual(activation.Safety, request.Authority.Safety) {
		add("unsafe_authority_boundary")
	}
	if len(activation.Partitions) != len(request.PlanReadback.Partitions) ||
		len(activation.Partitions) != 10 {
		add("activation_partition_mismatch")
	} else {
		for index, planned := range request.PlanReadback.Partitions {
			bound := activation.Partitions[index]
			if len(planned.Tests) != 1 ||
				bound.PartitionID != planned.PartitionID ||
				bound.NodeID != planned.NodeID ||
				bound.TestID != planned.Tests[0] ||
				bound.Classification != planned.Classification ||
				bound.RequestedRepeatCount != planned.RequestedRepeatCount ||
				bound.EffectiveRepeatCount != planned.EffectiveRepeatCount ||
				bound.EstimatedDurationMS != planned.EstimatedDurationMS ||
				bound.RetryAllowanceMS != planned.RetryAllowanceMS ||
				bound.NodeBudgetMS != planned.NodeBudgetMS ||
				bound.PerAttemptTimeoutMS != request.PlanInput.TimeoutPolicy.PerAttemptTimeoutMS ||
				bound.TotalNodeTimeoutMS != request.PlanInput.TimeoutPolicy.TotalNodeTimeoutMS {
				add("activation_partition_mismatch")
			}
		}
	}
	retryFound := false
	for _, partition := range activation.Partitions {
		if partition.NodeID == activation.ControlledRetryNodeID {
			retryFound = partition.Classification == "regular"
		}
	}
	if !retryFound || activation.ControlledRetryReason != "transient_infrastructure" {
		add("controlled_retry_binding_invalid")
	}
}

func validateSoakCanaryPaths(request SoakCanaryRunRequest, add func(string)) {
	if !sameCleanPath(request.EvidenceRoot, request.Authority.EvidenceRoot) {
		add("evidence_root_mismatch")
	}
	if !filepath.IsAbs(request.RepositoryRoot) {
		add("repository_root_invalid")
	}
	if !pathWithin(request.EvidenceRoot, request.CheckpointPath) {
		add("checkpoint_path_outside_evidence")
	}
}

func validSoakCanaryExecutable(command SoakCanaryCommand) bool {
	if !filepath.IsAbs(command.ExecutablePath) || filepath.Base(command.ExecutablePath) != "go" ||
		!validSoakHexDigest(command.ExecutableSHA256, 71, "sha256:") {
		return false
	}
	body, err := readBoundedRegularFile(command.ExecutablePath, 256<<20)
	return err == nil && digestBytes(body) == command.ExecutableSHA256
}

func validSoakCanaryEnvironment(environment []SoakCanaryEnvironment) bool {
	expected := []SoakCanaryEnvironment{
		{Name: "GOTOOLCHAIN", Value: "local"},
		{Name: "GOPROXY", Value: "off"},
		{Name: "GOSUMDB", Value: "off"},
		{Name: "GOVCS", Value: "*:off"},
	}
	return reflect.DeepEqual(environment, expected)
}

func soakCanaryArgvHasShell(argv []string) bool {
	for _, arg := range argv {
		if strings.ContainsAny(arg, ";&|`") || strings.Contains(arg, "$(") ||
			strings.Contains(arg, "${") || strings.ContainsAny(arg, "\r\n\x00") {
			return true
		}
	}
	return false
}

func soakCanarySafetyEnabled(safety SoakCanarySafety) bool {
	return safety.RepositoryMutationAllowed || safety.RuntimeProviderCallsAllowed ||
		safety.PublicationAllowed || safety.ReleaseAllowed || safety.DeploymentAllowed ||
		safety.CredentialChangesAllowed || safety.PermissionChangesAllowed ||
		safety.AuthorityExpansionAllowed || safety.NetworkAccessAllowed
}

func signSoakCanaryAuthority(authority *SoakCanaryAuthority) {
	authority.AuthorityRecordDigest = ""
	body, _ := json.Marshal(*authority)
	authority.AuthorityRecordDigest = digestBytes(body)
}

func signSoakCanaryCommandCatalog(catalog *SoakCanaryCommandCatalog) {
	catalog.CommandCatalogDigest = ""
	body, _ := json.Marshal(*catalog)
	catalog.CommandCatalogDigest = digestBytes(body)
}

func signSoakCanaryActivation(activation *SoakCanaryActivation) {
	activation.ActivationManifestDigest = ""
	body, _ := json.Marshal(*activation)
	activation.ActivationManifestDigest = digestBytes(body)
}

func signSoakCanaryAttempt(attempt *SoakCanaryAttempt) {
	attempt.AttemptDigest = ""
	body, _ := json.Marshal(*attempt)
	attempt.AttemptDigest = digestBytes(body)
}

func signSoakCanaryCheckpoint(checkpoint *SoakCanaryCheckpoint) {
	checkpoint.CheckpointDigest = ""
	body, _ := json.Marshal(*checkpoint)
	checkpoint.CheckpointDigest = digestBytes(body)
}

func signSoakCanarySummary(summary *SoakCanarySummary) {
	summary.SummaryDigest = ""
	body, _ := json.Marshal(*summary)
	summary.SummaryDigest = digestBytes(body)
}

func LoadSoakCanaryCheckpoint(path string) (SoakCanaryCheckpoint, error) {
	checkpoint, err := loadSoakCanaryJSON[SoakCanaryCheckpoint](path)
	if err != nil {
		return checkpoint, err
	}
	if checkpoint.Schema != SoakCanaryCheckpointSchema {
		return checkpoint, errors.New("soak canary checkpoint schema is invalid")
	}
	unsigned := checkpoint
	signSoakCanaryCheckpoint(&unsigned)
	if checkpoint.CheckpointDigest != unsigned.CheckpointDigest {
		return checkpoint, errors.New("soak canary checkpoint digest mismatch")
	}
	seenAttempts := map[string]bool{}
	completed := map[string]bool{}
	for _, attempt := range checkpoint.Attempts {
		unsignedAttempt := attempt
		signSoakCanaryAttempt(&unsignedAttempt)
		if attempt.Schema != SoakCanaryAttemptSchema ||
			attempt.AttemptDigest != unsignedAttempt.AttemptDigest ||
			seenAttempts[attempt.AttemptDigest] {
			return checkpoint, errors.New("soak canary attempt chain is invalid")
		}
		seenAttempts[attempt.AttemptDigest] = true
		if attempt.OutcomeClass == "passed" {
			completed[attempt.NodeID] = true
		}
	}
	expectedCompleted := sortedSoakCanaryKeys(completed)
	if !reflect.DeepEqual(checkpoint.CompletedNodeIDs, expectedCompleted) {
		return checkpoint, errors.New("soak canary checkpoint completion state is contradictory")
	}
	return checkpoint, nil
}

func writeSoakCanaryCheckpoint(path string, checkpoint SoakCanaryCheckpoint) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	body, err := json.MarshalIndent(checkpoint, "", "  ")
	if err != nil {
		return err
	}
	return writeAtomicFile(path, append(body, '\n'), 0o600)
}

func ReconcileSoakCanary(request SoakCanaryRunRequest, checkpoint SoakCanaryCheckpoint, completedAt time.Time) (SoakCanarySummary, error) {
	summary := SoakCanarySummary{
		Schema: SoakCanarySummarySchema, Status: "completed",
		CanaryID: request.Activation.CanaryID, MissionID: request.Activation.MissionID,
		PlanID: request.Activation.PlanID, SourceHead: request.Activation.SourceHead,
		PlanInputDigest:          request.Activation.PlanInputDigest,
		PolicyDigest:             request.Activation.PolicyDigest,
		CommandCatalogDigest:     request.Activation.CommandCatalogDigest,
		AuthorityRecordDigest:    request.Activation.AuthorityRecordDigest,
		ActivationManifestDigest: request.Activation.ActivationManifestDigest,
		PlannedPartitions:        len(request.Activation.Partitions),
		PlannedNodes:             len(request.Activation.Partitions),
		CompletedNodes:           len(checkpoint.CompletedNodeIDs),
		TotalAttempts:            len(checkpoint.Attempts),
		PhaseStartUTC:            request.Activation.PhaseStartUTC,
		CompletedAtUTC:           completedAt.UTC().Format(time.RFC3339Nano),
		LeaseMaximumMS:           request.PlanInput.Lease.MaximumMS,
		Attempts:                 append([]SoakCanaryAttempt(nil), checkpoint.Attempts...),
		ConflictCodes:            []string{},
		Safety:                   request.Activation.Safety,
	}
	phaseStart, err := time.Parse(time.RFC3339, request.Activation.PhaseStartUTC)
	if err != nil {
		return summary, err
	}
	summary.PhaseElapsedMS = completedAt.Sub(phaseStart).Milliseconds()
	if err := VerifySoakCanaryEvidence(request, checkpoint); err != nil {
		summary.Status = "failed"
		summary.ConflictCodes = []string{"attempt_output_digest_mismatch"}
		signSoakCanarySummary(&summary)
		return summary, err
	}
	conflicts := map[string]bool{}
	completions := map[string]int{}
	nodeElapsed := map[string]int64{}
	var aggregateElapsed int64
	for _, attempt := range checkpoint.Attempts {
		if attempt.ChildProcessLaunched {
			summary.ChildProcessLaunches++
			if attempt.Classification == "scale" {
				summary.ScaleLaunches++
			}
			aggregateElapsed = checkedSoakCanaryAdd(aggregateElapsed, attempt.ElapsedMS, conflicts)
		}
		if attempt.OutcomeClass == "transient_infrastructure" {
			summary.ControlledRetryCount++
		}
		if attempt.OutcomeClass == "passed" {
			completions[attempt.NodeID]++
		}
		nodeElapsed[attempt.NodeID] = checkedSoakCanaryAdd(nodeElapsed[attempt.NodeID], attempt.ElapsedMS, conflicts)
	}
	for _, partition := range request.Activation.Partitions {
		if completions[partition.NodeID] != 1 {
			conflicts["attempt_omission_or_duplicate_completion"] = true
		}
		if nodeElapsed[partition.NodeID] > partition.TotalNodeTimeoutMS {
			conflicts["total_node_timeout_exceeded"] = true
		}
		if nodeElapsed[partition.NodeID] > partition.NodeBudgetMS {
			conflicts["node_budget_exceeded"] = true
		}
	}
	if summary.PlannedNodes != 10 || summary.CompletedNodes != 10 ||
		summary.TotalAttempts != 11 || summary.ChildProcessLaunches != 10 ||
		summary.ScaleLaunches != 1 || summary.ControlledRetryCount != 1 {
		conflicts["canary_count_mismatch"] = true
	}
	if aggregateElapsed > request.PlanReadback.LeaseBudget.TotalPlannedWithRetryMS {
		conflicts["aggregate_duration_exceeded"] = true
	}
	if summary.PhaseElapsedMS < 0 || summary.PhaseElapsedMS > request.PlanInput.Lease.MaximumMS ||
		summary.PhaseElapsedMS > request.Authority.HardWallMS {
		conflicts["lease_maximum_exceeded"] = true
	}
	for code := range conflicts {
		summary.ConflictCodes = append(summary.ConflictCodes, code)
	}
	sort.Strings(summary.ConflictCodes)
	if len(summary.ConflictCodes) > 0 {
		summary.Status = "failed"
	}
	summary.LocalTestExecutionPerformed = summary.ChildProcessLaunches > 0
	index, _, err := buildSoakCanaryTerminalIndex(summary)
	if err != nil {
		return summary, err
	}
	summary.TerminalIndexDigest = index.Digest
	signSoakCanarySummary(&summary)
	if summary.Status != "completed" {
		return summary, fmt.Errorf("soak canary reconciliation failed: %s", strings.Join(summary.ConflictCodes, ","))
	}
	return summary, nil
}

func VerifySoakCanaryEvidence(request SoakCanaryRunRequest, checkpoint SoakCanaryCheckpoint) error {
	seenPaths := map[string]bool{}
	for _, attempt := range checkpoint.Attempts {
		for _, output := range []struct {
			name     string
			artifact SoakCanaryOutputArtifact
		}{
			{name: "stdout", artifact: attempt.Stdout},
			{name: "stderr", artifact: attempt.Stderr},
		} {
			if !attempt.ChildProcessLaunched {
				if output.artifact.Path != "" || output.artifact.Bytes != 0 ||
					output.artifact.SHA256 != digestBytes(nil) {
					return fmt.Errorf("%s evidence exists for an unlaunched attempt", output.name)
				}
				continue
			}
			if output.artifact.Path == "" || filepath.IsAbs(output.artifact.Path) ||
				filepath.Clean(output.artifact.Path) != output.artifact.Path ||
				seenPaths[output.artifact.Path] {
				return fmt.Errorf("%s evidence path is invalid", output.name)
			}
			seenPaths[output.artifact.Path] = true
			path := filepath.Join(request.EvidenceRoot, output.artifact.Path)
			if !pathWithin(request.EvidenceRoot, path) {
				return fmt.Errorf("%s evidence path escaped root", output.name)
			}
			body, err := readBoundedRegularFile(path, int64(request.OutputLimitBytes))
			if err != nil {
				return fmt.Errorf("%s evidence: %w", output.name, err)
			}
			if len(body) != output.artifact.Bytes || digestBytes(body) != output.artifact.SHA256 {
				return fmt.Errorf("%s digest mismatch", output.name)
			}
		}
	}
	return nil
}

type soakCanaryLeaseArtifact struct {
	Schema         string `json:"schema"`
	MissionID      string `json:"mission_id"`
	MinimumNodes   int    `json:"minimum_nodes"`
	MinimumMinutes int    `json:"minimum_minutes"`
	TargetMinutes  int    `json:"target_minutes"`
	MaximumMinutes int    `json:"maximum_minutes"`
}

type soakCanaryProgressArtifact struct {
	Schema               string              `json:"schema"`
	MissionID            string              `json:"mission_id"`
	CanaryID             string              `json:"canary_id"`
	PlanID               string              `json:"plan_id"`
	SourceHead           string              `json:"source_head"`
	ActivationDigest     string              `json:"activation_manifest_digest"`
	CommandCatalogDigest string              `json:"command_catalog_digest"`
	CompletedNodes       int                 `json:"completed_nodes"`
	ReadyNodes           int                 `json:"ready_nodes"`
	BlockedNodes         int                 `json:"blocked_nodes"`
	FailedNodes          int                 `json:"failed_nodes"`
	ElapsedMinutes       int                 `json:"elapsed_minutes,omitempty"`
	LeaseTimeStatus      string              `json:"lease_time_status,omitempty"`
	FinalResponseAllowed bool                `json:"final_response_allowed"`
	ExactNextAction      string              `json:"exact_next_action"`
	SafetyBoundaries     TerminalIndexSafety `json:"safety_boundaries"`
}

func WriteSoakCanaryTerminalBundle(root string, summary SoakCanarySummary) (string, error) {
	if !filepath.IsAbs(root) {
		return "", errors.New("soak canary terminal root must be absolute")
	}
	index, artifacts, err := buildSoakCanaryTerminalIndex(summary)
	if err != nil {
		return "", err
	}
	if summary.TerminalIndexDigest != index.Digest {
		return "", errors.New("soak canary terminal index digest does not match summary")
	}
	if err := os.MkdirAll(root, 0o700); err != nil {
		return "", err
	}
	for path, body := range artifacts {
		if err := writeAtomicFile(filepath.Join(root, path), body, 0o600); err != nil {
			return "", err
		}
	}
	indexBody, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return "", err
	}
	indexPath := filepath.Join(root, "canonical-terminal-index.json")
	if err := writeAtomicFile(indexPath, append(indexBody, '\n'), 0o600); err != nil {
		return "", err
	}
	if err := VerifyTerminalIndex(root, index); err != nil {
		return "", err
	}
	return indexPath, nil
}

func PersistSoakCanaryCompletion(evidenceRoot string, summary SoakCanarySummary) error {
	if !filepath.IsAbs(evidenceRoot) {
		return errors.New("soak canary evidence root must be absolute")
	}
	unsigned := summary
	signSoakCanarySummary(&unsigned)
	if summary.Schema != SoakCanarySummarySchema ||
		summary.Status != "completed" ||
		summary.SummaryDigest != unsigned.SummaryDigest {
		return errors.New("soak canary summary is not a valid completion")
	}
	if err := os.MkdirAll(evidenceRoot, 0o700); err != nil {
		return err
	}
	body, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	if err := writeAtomicFile(
		filepath.Join(evidenceRoot, "run-summary.json"),
		append(body, '\n'),
		0o600,
	); err != nil {
		return err
	}
	terminalRoot := filepath.Join(evidenceRoot, "terminal")
	indexPath, err := WriteSoakCanaryTerminalBundle(terminalRoot, summary)
	if err != nil {
		return err
	}
	imported, err := ImportTerminalIndex(
		terminalRoot,
		indexPath,
		filepath.Join(terminalRoot, "import-state.json"),
	)
	if err != nil {
		return err
	}
	surfaces := []string{"inspect", "checkpoint", "event-index", "command-readback"}
	readbacks := make([]TerminalIndexImportReadback, 0, len(surfaces))
	for _, surface := range surfaces {
		readback := imported
		readback.Surface = surface
		signTerminalIndexImport(&readback)
		readbacks = append(readbacks, readback)
	}
	if err := ValidateTerminalSurfaceAgreement(readbacks); err != nil {
		return err
	}
	body, err = json.MarshalIndent(readbacks, "", "  ")
	if err != nil {
		return err
	}
	return writeAtomicFile(
		filepath.Join(evidenceRoot, "terminal-surface-ledger.json"),
		append(body, '\n'),
		0o600,
	)
}

func buildSoakCanaryTerminalIndex(summary SoakCanarySummary) (CanonicalTerminalIndex, map[string][]byte, error) {
	if summary.Status != "completed" || len(summary.ConflictCodes) != 0 ||
		summary.PlannedNodes != 10 || summary.CompletedNodes != 10 {
		return CanonicalTerminalIndex{}, nil, errors.New("soak canary terminal index requires a completed ten-node summary")
	}
	minimumMinutes := 1
	targetMinutes := int((summary.LeaseMaximumMS/2 + 59_999) / 60_000)
	if targetMinutes < minimumMinutes {
		targetMinutes = minimumMinutes
	}
	maximumMinutes := int((summary.LeaseMaximumMS + 59_999) / 60_000)
	if maximumMinutes < targetMinutes {
		maximumMinutes = targetMinutes
	}
	elapsedMinutes := int((summary.PhaseElapsedMS + 59_999) / 60_000)
	if elapsedMinutes < minimumMinutes {
		elapsedMinutes = minimumMinutes
	}
	lease := soakCanaryLeaseArtifact{
		Schema: "ao.mission.soak-canary-lease.v1", MissionID: summary.MissionID,
		MinimumNodes: summary.PlannedNodes, MinimumMinutes: minimumMinutes,
		TargetMinutes: targetMinutes, MaximumMinutes: maximumMinutes,
	}
	root := soakCanaryProgressArtifact{
		Schema:    "ao.mission.soak-canary-progress.v1",
		MissionID: summary.MissionID, CanaryID: summary.CanaryID, PlanID: summary.PlanID,
		SourceHead:           summary.SourceHead,
		ActivationDigest:     summary.ActivationManifestDigest,
		CommandCatalogDigest: summary.CommandCatalogDigest,
		ReadyNodes:           summary.PlannedNodes,
		ExactNextAction:      "Run only the activated bounded local soak canary.",
		SafetyBoundaries:     TerminalIndexSafety{},
	}
	terminal := soakCanaryProgressArtifact{
		Schema:    "ao.mission.soak-canary-terminal.v1",
		MissionID: summary.MissionID, CanaryID: summary.CanaryID, PlanID: summary.PlanID,
		SourceHead:           summary.SourceHead,
		ActivationDigest:     summary.ActivationManifestDigest,
		CommandCatalogDigest: summary.CommandCatalogDigest,
		CompletedNodes:       summary.CompletedNodes,
		ElapsedMinutes:       elapsedMinutes, LeaseTimeStatus: "within_window",
		FinalResponseAllowed: true, ExactNextAction: "none",
		SafetyBoundaries: TerminalIndexSafety{},
	}
	values := []struct {
		role, state, path string
		value             any
	}{
		{role: "lease", state: "lease_authority", path: "lease.json", value: lease},
		{role: "root", state: "initial_snapshot", path: "root.json", value: root},
		{role: "terminal", state: "terminal_candidate", path: "terminal.json", value: terminal},
	}
	index := CanonicalTerminalIndex{
		ContractVersion: TerminalIndexContract,
		SchemaDigest:    digestBytes([]byte(TerminalIndexContract)),
		MissionID:       summary.MissionID, EvidenceRoot: ".",
		GeneratedAtUTC: summary.CompletedAtUTC, TerminalReference: "terminal.json",
		Counts: TerminalIndexCounts{
			Total: summary.PlannedNodes, Minimum: summary.PlannedNodes,
			Completed: summary.CompletedNodes,
		},
		Lease: TerminalIndexLease{
			MinimumMinutes: minimumMinutes, TargetMinutes: targetMinutes,
			MaximumMinutes: maximumMinutes, ElapsedMinutes: elapsedMinutes,
			Status: "within_window",
		},
		CompletionObserved: true, CanonicalEvidenceAgreement: true,
		ReadinessPassed: true, ReturnGateStatus: "final_response_allowed",
		FinalResponseAllowed: true, ConflictCodes: []string{},
		ConflictSummaries: []string{}, ExactNextAction: "none",
		SafetyBoundaries: TerminalIndexSafety{},
	}
	artifacts := map[string][]byte{}
	for sequence, value := range values {
		body, err := json.MarshalIndent(value.value, "", "  ")
		if err != nil {
			return CanonicalTerminalIndex{}, nil, err
		}
		body = append(body, '\n')
		artifacts[value.path] = body
		var raw map[string]any
		if err := json.Unmarshal(body, &raw); err != nil {
			return CanonicalTerminalIndex{}, nil, err
		}
		index.Artifacts = append(index.Artifacts, TerminalIndexArtifact{
			Role: value.role, Sequence: sequence, Path: value.path,
			Schema: terminalFirstString(raw, "schema"),
			SHA256: digestBytes(body), State: value.state,
		})
		if sequence > 0 {
			index.Lineage = append(index.Lineage, TerminalIndexLineage{
				FromSequence: sequence - 1, ToSequence: sequence, Relation: "precedes",
			})
		}
	}
	signCanonicalTerminalIndex(&index)
	return index, artifacts, nil
}

func BuildSoakCanaryTerminalReadbacks(summary SoakCanarySummary) ([]TerminalIndexImportReadback, error) {
	if summary.Status != "completed" || len(summary.ConflictCodes) != 0 ||
		summary.CompletedNodes != summary.PlannedNodes {
		return nil, errors.New("soak canary terminal readback requires a completed summary")
	}
	minimumMinutes := 1
	targetMinutes := int((summary.LeaseMaximumMS/2 + 59_999) / 60_000)
	if targetMinutes < minimumMinutes {
		targetMinutes = minimumMinutes
	}
	maximumMinutes := int((summary.LeaseMaximumMS + 59_999) / 60_000)
	if maximumMinutes < targetMinutes {
		maximumMinutes = targetMinutes
	}
	elapsedMinutes := int((summary.PhaseElapsedMS + 59_999) / 60_000)
	if elapsedMinutes < minimumMinutes {
		elapsedMinutes = minimumMinutes
	}
	base := TerminalIndexImportReadback{
		Schema: TerminalIndexImportSchema, MissionID: summary.MissionID,
		IndexDigest:    summary.TerminalIndexDigest,
		GeneratedAtUTC: summary.CompletedAtUTC, Status: "reconciled",
		Counts: TerminalIndexCounts{
			Total: summary.PlannedNodes, Minimum: summary.PlannedNodes,
			Completed: summary.CompletedNodes,
		},
		Lease: TerminalIndexLease{
			MinimumMinutes: minimumMinutes, TargetMinutes: targetMinutes,
			MaximumMinutes: maximumMinutes, ElapsedMinutes: elapsedMinutes,
			Status: "within_window",
		},
		CompletionObserved: true, TimingCompliant: true,
		CanonicalEvidenceAgreement: true, ReadinessPassed: true,
		ReturnGateStatus: "final_response_allowed", FinalResponseAllowed: true,
		ConflictCodes:   []string{},
		ExactNextAction: "none", ReadOnly: true,
	}
	surfaces := []string{"inspect", "checkpoint", "event-index", "command-readback"}
	readbacks := make([]TerminalIndexImportReadback, 0, len(surfaces))
	for _, surface := range surfaces {
		readback := base
		readback.Surface = surface
		signTerminalIndexImport(&readback)
		readbacks = append(readbacks, readback)
	}
	return readbacks, nil
}

func checkedSoakCanaryAdd(left, right int64, conflicts map[string]bool) int64 {
	if left < 0 || right < 0 || right > math.MaxInt64-left {
		conflicts["duration_arithmetic_overflow"] = true
		return math.MaxInt64
	}
	return left + right
}

func sortedSoakCanaryKeys(values map[string]bool) []string {
	keys := make([]string, 0, len(values))
	for key, present := range values {
		if present {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}

func sameCleanPath(left, right string) bool {
	leftAbs, leftErr := filepath.Abs(left)
	rightAbs, rightErr := filepath.Abs(right)
	return leftErr == nil && rightErr == nil && filepath.Clean(leftAbs) == filepath.Clean(rightAbs)
}

func pathWithin(root, path string) bool {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return false
	}
	pathAbs, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	relative, err := filepath.Rel(rootAbs, pathAbs)
	return err == nil && relative != ".." &&
		!strings.HasPrefix(relative, ".."+string(filepath.Separator))
}
