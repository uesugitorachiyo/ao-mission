package mission

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestBuildSoakPlanSeparatesScaleFromRepeatedRegularWork(t *testing.T) {
	input := validSoakPlanInput()
	readback := buildValidSoakPlan(t, input)

	if !readback.ActivationAllowed || readback.ConflictCodes == nil || len(readback.ConflictCodes) != 0 {
		t.Fatalf("valid plan failed closed: %+v", readback)
	}
	if readback.InputDigest == "" || readback.PolicyDigest == "" {
		t.Fatalf("valid plan omitted digests: %+v", readback)
	}
	if len(readback.Partitions) != 2 {
		t.Fatalf("mixed work was not split: %+v", readback.Partitions)
	}
	scale, regular := readback.Partitions[0], readback.Partitions[1]
	if scale.Classification != "scale" || scale.RequestedRepeatCount != 1 || scale.EffectiveRepeatCount != 1 ||
		scale.AmplificationDecision != "scale_repeat_one_preserved" ||
		len(scale.Tests) != 1 || scale.Tests[0] != "scale-index-10000" {
		t.Fatalf("scale partition was amplified: %+v", scale)
	}
	if regular.Classification != "regular" || regular.RequestedRepeatCount != 3 || regular.EffectiveRepeatCount != 3 ||
		regular.AmplificationDecision != "bounded_regular_repeat_preserved" ||
		!reflect.DeepEqual(regular.Tests, []string{"regular-alpha", "regular-beta"}) {
		t.Fatalf("regular partition lost bounded repeats: %+v", regular)
	}
	if scale.EstimatedDurationMS != 1_500 || regular.EstimatedDurationMS != 1_260 {
		t.Fatalf("conservative estimates changed: scale=%d regular=%d", scale.EstimatedDurationMS, regular.EstimatedDurationMS)
	}
	if readback.DurationHistory.Estimator != "nearest_rank_p95_plus_partition_overhead_ms" ||
		readback.DurationHistory.SampleCount != 9 ||
		readback.LeaseBudget.TotalPlannedWithRetryMS != 5_520 ||
		readback.ExactNextAction != "Review this read-only plan and activate it only through a separately authorized execution system." {
		t.Fatalf("unexpected summary: %+v", readback)
	}
	assertSoakSafety(t, readback.SafetyBoundaries)
}

func TestBuildSoakPlanDurationEstimateIsIndependentOfSampleOrder(t *testing.T) {
	first := validSoakPlanInput()
	second := validSoakPlanInput()
	for i := range second.DurationHistory {
		sort.Slice(second.DurationHistory[i].Samples, func(a, b int) bool {
			return second.DurationHistory[i].Samples[a] > second.DurationHistory[i].Samples[b]
		})
	}
	first.PolicyDigest = soakPolicyDigest(first)
	second.PolicyDigest = soakPolicyDigest(second)
	gotFirst := buildValidSoakPlan(t, first)
	gotSecond := buildValidSoakPlan(t, second)
	if !reflect.DeepEqual(gotFirst, gotSecond) {
		t.Fatalf("history ordering changed semantic output:\nfirst=%+v\nsecond=%+v", gotFirst, gotSecond)
	}
}

func TestBuildSoakPlanFailsClosedWithExactConflictCodes(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*SoakPlanInput)
		want   []string
	}{
		{name: "missing classification", mutate: func(in *SoakPlanInput) {
			in.TestCatalog[0].Classification = ""
		}, want: []string{"classification_missing"}},
		{name: "unknown classification", mutate: func(in *SoakPlanInput) {
			in.TestCatalog[0].Classification = "expensive"
		}, want: []string{"classification_unknown"}},
		{name: "classification after partitioning", mutate: func(in *SoakPlanInput) {
			in.ClassificationBoundBeforePartitioning = false
		}, want: []string{"classification_after_partitioning"}},
		{name: "contradictory classification", mutate: func(in *SoakPlanInput) {
			in.TestCatalog[0].ScaleDimension = &SoakScaleDimension{Unit: "records", Value: 10}
		}, want: []string{"classification_contradictory"}},
		{name: "execution mode missing", mutate: func(in *SoakPlanInput) {
			in.ExecutionProfile.Mode = ""
		}, want: []string{"execution_profile_mode_invalid"}},
		{name: "scale repeat", mutate: func(in *SoakPlanInput) {
			in.TestCatalog[2].RequestedRepeatCount = 2
		}, want: []string{"scale_repeat_amplification"}},
		{name: "mixed repeat inheritance", mutate: func(in *SoakPlanInput) {
			in.RepeatPolicy.ApplyPartitionRepeatToAll = true
		}, want: []string{"mixed_partition_scale_amplification"}},
		{name: "regular repeat above maximum", mutate: func(in *SoakPlanInput) {
			in.TestCatalog[0].RequestedRepeatCount = 5
		}, want: []string{"repeat_limit_exceeded"}},
		{name: "history missing", mutate: func(in *SoakPlanInput) {
			in.DurationHistory = in.DurationHistory[1:]
		}, want: []string{"duration_history_missing"}},
		{name: "history empty", mutate: func(in *SoakPlanInput) {
			in.DurationHistory[0].Samples = nil
		}, want: []string{"duration_history_empty"}},
		{name: "history insufficient", mutate: func(in *SoakPlanInput) {
			in.DurationHistory[0].Samples = []int64{100, 110}
		}, want: []string{"duration_history_insufficient"}},
		{name: "history non positive", mutate: func(in *SoakPlanInput) {
			in.DurationHistory[0].Samples[0] = 0
		}, want: []string{"duration_history_non_positive"}},
		{name: "history excessive", mutate: func(in *SoakPlanInput) {
			in.DurationHistory[0].Samples = make([]int64, soakPlanMaxHistorySamples+1)
			for i := range in.DurationHistory[0].Samples {
				in.DurationHistory[0].Samples[i] = 1
			}
		}, want: []string{"duration_history_sample_limit_exceeded"}},
		{name: "history overflow risk", mutate: func(in *SoakPlanInput) {
			in.DurationHistory[0].Samples[0] = soakPlanMaxDurationMS + 1
		}, want: []string{"duration_history_overflow_risk"}},
		{name: "wrong history head", mutate: func(in *SoakPlanInput) {
			in.DurationHistory[0].SourceHead = strings.Repeat("b", 40)
		}, want: []string{"duration_history_source_head_mismatch"}},
		{name: "wrong history profile", mutate: func(in *SoakPlanInput) {
			in.DurationHistory[0].ExecutionProfileDigest = "sha256:" + strings.Repeat("b", 64)
		}, want: []string{"duration_history_profile_mismatch"}},
		{name: "wrong history unit", mutate: func(in *SoakPlanInput) {
			in.DurationHistory[0].Unit = "seconds"
		}, want: []string{"duration_history_unit_mismatch"}},
		{name: "node estimate above budget", mutate: func(in *SoakPlanInput) {
			in.PartitionBudgets[0].NodeBudgetMS = 1_000
		}, want: []string{"partition_node_budget_exceeded", "retry_node_budget_exceeded"}},
		{name: "attempt timeout below estimate", mutate: func(in *SoakPlanInput) {
			in.TimeoutPolicy.PerAttemptTimeoutMS = 1_000
		}, want: []string{"timeout_below_estimate"}},
		{name: "retry allowance above total timeout", mutate: func(in *SoakPlanInput) {
			in.TimeoutPolicy.TotalNodeTimeoutMS = 2_000
		}, want: []string{"retry_total_timeout_exceeded"}},
		{name: "retry allowance above lease", mutate: func(in *SoakPlanInput) {
			in.Lease.MaximumMS = 5_000
			in.Lease.TargetMS = 4_000
		}, want: []string{"lease_budget_exceeded"}},
		{name: "missing retry policy", mutate: func(in *SoakPlanInput) {
			in.RetryPolicy = nil
		}, want: []string{"retry_policy_missing"}},
		{name: "activation policy not bound", mutate: func(in *SoakPlanInput) {
			in.Activation.PolicyBoundBeforeActivation = false
		}, want: []string{"activation_predates_policy"}},
		{name: "activated digest changed", mutate: func(in *SoakPlanInput) {
			in.Activation.BoundPolicyDigest = "sha256:" + strings.Repeat("0", 64)
		}, want: []string{"activation_policy_digest_mismatch"}},
		{name: "retry changes node identity", mutate: func(in *SoakPlanInput) {
			in.RetryPolicy.PreservesNodeIdentity = false
		}, want: []string{"retry_node_identity_changed"}},
		{name: "retry changes test set", mutate: func(in *SoakPlanInput) {
			in.RetryPolicy.PreservesTestSet = false
		}, want: []string{"retry_test_set_changed"}},
		{name: "retry changes scale", mutate: func(in *SoakPlanInput) {
			in.RetryPolicy.PreservesScaleDimension = false
		}, want: []string{"retry_scale_dimension_changed"}},
		{name: "retry changes repeat", mutate: func(in *SoakPlanInput) {
			in.RetryPolicy.PreservesRepeatCount = false
		}, want: []string{"retry_repeat_count_changed"}},
		{name: "retry changes head", mutate: func(in *SoakPlanInput) {
			in.RetryPolicy.PreservesSourceHead = false
		}, want: []string{"retry_source_head_changed"}},
		{name: "retry changes profile", mutate: func(in *SoakPlanInput) {
			in.RetryPolicy.PreservesExecutionProfile = false
		}, want: []string{"retry_execution_profile_changed"}},
		{name: "retry changes phase start", mutate: func(in *SoakPlanInput) {
			in.RetryPolicy.PreservesPhaseStart = false
		}, want: []string{"retry_phase_start_changed"}},
		{name: "retry widens authority", mutate: func(in *SoakPlanInput) {
			in.RetryPolicy.PreservesAuthorityBoundaries = false
		}, want: []string{"retry_authority_broadened"}},
		{name: "retry widens side effects", mutate: func(in *SoakPlanInput) {
			in.RetryPolicy.PreservesSideEffectBoundaries = false
		}, want: []string{"retry_side_effects_broadened"}},
		{name: "retry resets phase clock", mutate: func(in *SoakPlanInput) {
			in.RetryPolicy.PhaseClockResetProhibited = false
		}, want: []string{"retry_phase_clock_reset"}},
		{name: "retry evidence missing", mutate: func(in *SoakPlanInput) {
			in.RetryPolicy.EvidenceRequiredAfterFailure = ""
		}, want: []string{"retry_policy_incomplete"}},
		{name: "unsafe authority", mutate: func(in *SoakPlanInput) {
			in.SafetyBoundaries.CallsProviders = true
		}, want: []string{"unsafe_authority_boundary"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := validSoakPlanInput()
			test.mutate(&input)
			input.PolicyDigest = soakPolicyDigest(input)
			if test.name == "activated digest changed" {
				input.Activation.BoundPolicyDigest = "sha256:" + strings.Repeat("0", 64)
			} else {
				input.Activation.BoundPolicyDigest = input.PolicyDigest
			}
			readback, err := BuildSoakPlan(input)
			if err != nil {
				t.Fatal(err)
			}
			if readback.ActivationAllowed {
				t.Fatalf("invalid plan was activation eligible: %+v", readback)
			}
			for _, code := range test.want {
				if !containsSoakConflict(readback.ConflictCodes, code) {
					t.Fatalf("conflicts=%v want %q", readback.ConflictCodes, code)
				}
			}
			if readback.ExactNextAction == "" {
				t.Fatal("fail-closed readback omitted exact next action")
			}
			assertSoakSafety(t, readback.SafetyBoundaries)
		})
	}
}

func TestLoadSoakPlanInputRejectsUnsafeFixtureTransportAndJSON(t *testing.T) {
	input := validSoakPlanInput()
	input.PolicyDigest = soakPolicyDigest(input)
	input.Activation.BoundPolicyDigest = input.PolicyDigest
	body, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	validPath := filepath.Join(dir, "valid.json")
	if err := os.WriteFile(validPath, body, 0o600); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		path string
		want string
	}{
		{name: "malformed", path: writeSoakTestFile(t, dir, "malformed.json", []byte("{")), want: "invalid JSON"},
		{name: "duplicate", path: writeSoakTestFile(t, dir, "duplicate.json", bytes.Replace(body, []byte(`"schema":`), []byte(`"schema":"duplicate","schema":`), 1)), want: "duplicate JSON key"},
		{name: "trailing", path: writeSoakTestFile(t, dir, "trailing.json", append(append([]byte{}, body...), []byte(` {}`)...)), want: "trailing JSON"},
		{name: "oversized", path: writeSoakTestFile(t, dir, "oversized.json", bytes.Repeat([]byte("x"), soakPlanMaxFixtureBytes+1)), want: "size limit"},
		{name: "traversal", path: filepath.Join(dir, "..", "outside.json"), want: "safe relative fixture path"},
	}
	symlinkPath := filepath.Join(dir, "symlink.json")
	if err := os.Symlink(validPath, symlinkPath); err != nil {
		t.Fatal(err)
	}
	tests = append(tests, struct {
		name string
		path string
		want string
	}{name: "symlink", path: symlinkPath, want: "regular non-symlink file"})

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := LoadSoakPlanInput(dir, test.path); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error=%v want substring %q", err, test.want)
			}
			if test.name != "traversal" {
				var stdout, stderr bytes.Buffer
				if code := Run([]string{"qualification", "soak-plan", "--fixture", test.path, "--json"}, &stdout, &stderr); code == 0 ||
					!strings.Contains(stderr.String(), test.want) {
					t.Fatalf("CLI code=%d stdout=%s stderr=%s want substring %q", code, stdout.String(), stderr.String(), test.want)
				}
			}
		})
	}
}

func TestSoakPlanCLIPrintsDeterministicReadbackAndFailsClosed(t *testing.T) {
	dir := t.TempDir()
	input := validSoakPlanInput()
	input.PolicyDigest = soakPolicyDigest(input)
	input.Activation.BoundPolicyDigest = input.PolicyDigest
	validPath := writeSoakInputFile(t, dir, "valid.json", input)

	var stdout, stderr bytes.Buffer
	if code := Run([]string{"qualification", "soak-plan", "--fixture", validPath, "--json"}, &stdout, &stderr); code != 0 {
		t.Fatalf("valid CLI code=%d stderr=%s", code, stderr.String())
	}
	var valid SoakPlanReadback
	if err := json.Unmarshal(stdout.Bytes(), &valid); err != nil {
		t.Fatal(err)
	}
	if !valid.ActivationAllowed || len(valid.Partitions) != 2 || !valid.ReadOnly {
		t.Fatalf("unexpected valid CLI readback: %+v", valid)
	}

	input.TestCatalog[2].RequestedRepeatCount = 2
	input.PolicyDigest = soakPolicyDigest(input)
	input.Activation.BoundPolicyDigest = input.PolicyDigest
	invalidPath := writeSoakInputFile(t, dir, "invalid.json", input)
	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"qualification", "soak-plan", "--fixture", invalidPath, "--json"}, &stdout, &stderr); code != 0 {
		t.Fatalf("fail-closed contract should emit readback, code=%d stderr=%s", code, stderr.String())
	}
	var invalid SoakPlanReadback
	if err := json.Unmarshal(stdout.Bytes(), &invalid); err != nil {
		t.Fatal(err)
	}
	if invalid.ActivationAllowed || !reflect.DeepEqual(invalid.ConflictCodes, []string{"scale_repeat_amplification"}) {
		t.Fatalf("invalid CLI did not expose exact conflict: %+v", invalid)
	}

	stdout.Reset()
	stderr.Reset()
	traversal := dir + string(filepath.Separator) + "nested" + string(filepath.Separator) + ".." + string(filepath.Separator) + "valid.json"
	if code := Run([]string{"qualification", "soak-plan", "--fixture", traversal, "--json"}, &stdout, &stderr); code == 0 ||
		!strings.Contains(stderr.String(), "traversal") {
		t.Fatalf("traversal code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
}

func TestSoakPlanRepositoryFixturesExposeExactActivationDecisions(t *testing.T) {
	tests := []struct {
		path         string
		allowed      bool
		wantConflict string
	}{
		{path: filepath.Join("..", "..", "examples", "valid", "soak-plan-mixed.json"), allowed: true},
		{path: filepath.Join("..", "..", "examples", "invalid", "soak-plan-scale-amplification.json"), wantConflict: "scale_repeat_amplification"},
		{path: filepath.Join("..", "..", "examples", "invalid", "soak-plan-timeout-below-estimate.json"), wantConflict: "timeout_below_estimate"},
		{path: filepath.Join("..", "..", "examples", "invalid", "soak-plan-unsafe-authority.json"), wantConflict: "unsafe_authority_boundary"},
	}
	for _, test := range tests {
		t.Run(filepath.Base(test.path), func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if code := Run([]string{"qualification", "soak-plan", "--fixture", test.path, "--json"}, &stdout, &stderr); code != 0 {
				t.Fatalf("code=%d stderr=%s", code, stderr.String())
			}
			var readback SoakPlanReadback
			if err := json.Unmarshal(stdout.Bytes(), &readback); err != nil {
				t.Fatal(err)
			}
			if readback.ActivationAllowed != test.allowed {
				t.Fatalf("allowed=%t want=%t conflicts=%v", readback.ActivationAllowed, test.allowed, readback.ConflictCodes)
			}
			if test.wantConflict != "" && !reflect.DeepEqual(readback.ConflictCodes, []string{test.wantConflict}) {
				t.Fatalf("conflicts=%v want exactly %q", readback.ConflictCodes, test.wantConflict)
			}
		})
	}
}

func validSoakPlanInput() SoakPlanInput {
	sourceHead := strings.Repeat("a", 40)
	profileDigest := "sha256:" + strings.Repeat("1", 64)
	input := SoakPlanInput{
		Schema:     SoakPlanInputSchema,
		PlanID:     "plan-fixture-001",
		MissionID:  "mission-fixture-001",
		SourceHead: sourceHead,
		ExecutionProfile: SoakExecutionProfile{
			ID: "darwin-arm64-go1.26.4-race", Digest: profileDigest, Mode: "race", Race: true,
		},
		ClassificationBoundBeforePartitioning: true,
		TestCatalog: []SoakTestEntry{
			{ID: "regular-alpha", Classification: "regular", RequestedRepeatCount: 3},
			{ID: "regular-beta", Classification: "regular", RequestedRepeatCount: 3},
			{ID: "scale-index-10000", Classification: "scale", RequestedRepeatCount: 1, ScaleDimension: &SoakScaleDimension{Unit: "records", Value: 10_000}},
		},
		DurationHistory: []SoakDurationHistory{
			{TestID: "regular-alpha", SourceHead: sourceHead, ExecutionProfileDigest: profileDigest, Unit: "milliseconds", Samples: []int64{100, 120, 110}},
			{TestID: "regular-beta", SourceHead: sourceHead, ExecutionProfileDigest: profileDigest, Unit: "milliseconds", Samples: []int64{200, 180, 190}},
			{TestID: "scale-index-10000", SourceHead: sourceHead, ExecutionProfileDigest: profileDigest, Unit: "milliseconds", Samples: []int64{1_100, 1_200, 1_150}},
		},
		Partitions: []SoakPartitionRequest{{
			PartitionID: "partition-01",
			NodeID:      "node-01",
			TestIDs:     []string{"regular-beta", "scale-index-10000", "regular-alpha"},
		}},
		PartitionBudgets: []SoakPartitionBudget{{
			PartitionID: "partition-01", NodeBudgetMS: 10_000,
		}},
		Budgets: SoakPlanBudgets{
			MaximumTests: 16, MaximumPartitions: 16, SetupOverheadMS: 100, SafetyOverheadMS: 200,
		},
		RepeatPolicy: SoakRepeatPolicy{
			MaximumRegularRepeatCount: 3, ScaleRepeatCount: 1, ApplyPartitionRepeatToAll: false,
		},
		RetryPolicy: &SoakRetryPolicy{
			MaximumAttempts:               2,
			RetryableOutcomeClasses:       []string{"timeout", "transient_infrastructure"},
			NonRetryableOutcomeClasses:    []string{"test_failure", "contract_violation"},
			CheckpointBehavior:            "preserve_failed_attempt_evidence",
			PreservesNodeIdentity:         true,
			PreservesTestSet:              true,
			PreservesScaleDimension:       true,
			PreservesRepeatCount:          true,
			PreservesSourceHead:           true,
			PreservesExecutionProfile:     true,
			PreservesPhaseStart:           true,
			PreservesAuthorityBoundaries:  true,
			PreservesSideEffectBoundaries: true,
			PhaseClockResetProhibited:     true,
			EvidenceRequiredAfterFailure:  "exit_code_stdout_stderr_duration_and_fixture_digest",
		},
		TimeoutPolicy: SoakTimeoutPolicy{
			PerAttemptTimeoutMS: 2_000, TotalNodeTimeoutMS: 4_000,
		},
		Lease: SoakLeaseBudget{
			MinimumMS: 1_000, TargetMS: 7_000, MaximumMS: 8_000,
		},
		Activation: SoakActivationBinding{
			State: "pre_activation", PolicyBoundBeforeActivation: true,
		},
		SafetyBoundaries: SoakSafetyBoundaries{
			SafeToExecute: false, ExecutesWork: false, ApprovesWork: false,
			MutatesRepositories: false, CallsProviders: false, Publishes: false,
			Releases: false, Deploys: false, AdvancesAuthority: false,
			RSIRemainsDenied: true,
		},
	}
	input.PolicyDigest = soakPolicyDigest(input)
	input.Activation.BoundPolicyDigest = input.PolicyDigest
	return input
}

func buildValidSoakPlan(t *testing.T, input SoakPlanInput) SoakPlanReadback {
	t.Helper()
	input.PolicyDigest = soakPolicyDigest(input)
	input.Activation.BoundPolicyDigest = input.PolicyDigest
	readback, err := BuildSoakPlan(input)
	if err != nil {
		t.Fatal(err)
	}
	return readback
}

func writeSoakTestFile(t *testing.T, dir, name string, body []byte) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, body, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeSoakInputFile(t *testing.T, dir, name string, input SoakPlanInput) string {
	t.Helper()
	body, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return writeSoakTestFile(t, dir, name, append(body, '\n'))
}

func containsSoakConflict(codes []string, want string) bool {
	for _, code := range codes {
		if code == want {
			return true
		}
	}
	return false
}

func assertSoakSafety(t *testing.T, safety SoakSafetyBoundaries) {
	t.Helper()
	if safety.SafeToExecute || safety.ExecutesWork || safety.ApprovesWork ||
		safety.MutatesRepositories || safety.CallsProviders || safety.Publishes ||
		safety.Releases || safety.Deploys || safety.AdvancesAuthority ||
		!safety.RSIRemainsDenied {
		t.Fatalf("safety boundary widened: %+v", safety)
	}
}
