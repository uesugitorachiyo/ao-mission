package mission

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type SoakCanaryClock interface {
	Now() time.Time
}

type SoakCanaryExecRequest struct {
	TestID            string
	ExpectedTestName  string
	ExpectedPassCount int
	ExecutablePath    string
	Argv              []string
	RepositoryRoot    string
	WorkingDirectory  string
	Environment       []SoakCanaryEnvironment
	TimeoutMS         int64
	OutputLimitBytes  int
}

type SoakCanaryProcessResult struct {
	ExitCode        int
	Signal          string
	ElapsedMS       int64
	Stdout          []byte
	Stderr          []byte
	StdoutTruncated bool
	StderrTruncated bool
}

type SoakCanaryExecutor interface {
	Start(context.Context, SoakCanaryExecRequest) (SoakCanaryProcess, error)
}

type SoakCanaryProcess interface {
	PID() int
	Wait() SoakCanaryProcessResult
}

type realSoakCanaryClock struct{}

func (realSoakCanaryClock) Now() time.Time {
	return time.Now().UTC()
}

type OSExecSoakCanaryExecutor struct{}

type osExecSoakCanaryProcess struct {
	command *exec.Cmd
	started time.Time
	stdout  *soakCanaryBoundedBuffer
	stderr  *soakCanaryBoundedBuffer
	once    sync.Once
	result  SoakCanaryProcessResult
}

func (OSExecSoakCanaryExecutor) Start(ctx context.Context, request SoakCanaryExecRequest) (SoakCanaryProcess, error) {
	if !filepath.IsAbs(request.ExecutablePath) || request.WorkingDirectory != "." ||
		!validSoakCanaryEnvironment(request.Environment) || request.OutputLimitBytes <= 0 {
		return nil, errors.New("soak canary executor received an unvalidated request")
	}
	command := exec.CommandContext(ctx, request.ExecutablePath, request.Argv...)
	command.Dir = filepath.Join(request.RepositoryRoot, request.WorkingDirectory)
	command.Env = soakCanarySanitizedEnvironment(request.Environment)
	stdout := &soakCanaryBoundedBuffer{limit: request.OutputLimitBytes}
	stderr := &soakCanaryBoundedBuffer{limit: request.OutputLimitBytes}
	command.Stdout = stdout
	command.Stderr = stderr
	started := time.Now()
	if err := command.Start(); err != nil {
		return nil, err
	}
	return &osExecSoakCanaryProcess{
		command: command, started: started, stdout: stdout, stderr: stderr,
	}, nil
}

func (process *osExecSoakCanaryProcess) PID() int {
	if process.command.Process == nil {
		return 0
	}
	return process.command.Process.Pid
}

func (process *osExecSoakCanaryProcess) Wait() SoakCanaryProcessResult {
	process.once.Do(func() {
		err := process.command.Wait()
		process.result = SoakCanaryProcessResult{
			ExitCode:  soakCanaryExitCode(err),
			ElapsedMS: time.Since(process.started).Milliseconds(),
			Stdout:    process.stdout.Bytes(), Stderr: process.stderr.Bytes(),
			StdoutTruncated: process.stdout.truncated,
			StderrTruncated: process.stderr.truncated,
		}
		if exit, ok := err.(*exec.ExitError); ok && exit.ProcessState != nil {
			process.result.Signal = soakCanaryProcessSignal(exit.ProcessState.String())
		}
	})
	return process.result
}

type soakCanaryBoundedBuffer struct {
	buffer    bytes.Buffer
	limit     int
	truncated bool
}

func (buffer *soakCanaryBoundedBuffer) Write(body []byte) (int, error) {
	originalLength := len(body)
	remaining := buffer.limit - buffer.buffer.Len()
	if remaining <= 0 {
		buffer.truncated = buffer.truncated || originalLength > 0
		return originalLength, nil
	}
	if len(body) > remaining {
		buffer.truncated = true
		body = body[:remaining]
	}
	_, _ = buffer.buffer.Write(body)
	return originalLength, nil
}

func (buffer *soakCanaryBoundedBuffer) Bytes() []byte {
	return append([]byte(nil), buffer.buffer.Bytes()...)
}

func RunSoakCanary(ctx context.Context, request SoakCanaryRunRequest) (SoakCanarySummary, error) {
	validation := ValidateSoakCanaryActivation(request)
	if !validation.ActivationAllowed {
		summary := rejectedSoakCanarySummary(request, validation.ConflictCodes)
		return summary, fmt.Errorf("soak canary activation rejected: %s", strings.Join(validation.ConflictCodes, ","))
	}
	if request.Executor == nil {
		summary := rejectedSoakCanarySummary(request, []string{"executor_missing"})
		return summary, errors.New("soak canary executor is required")
	}
	if request.Clock == nil {
		request.Clock = realSoakCanaryClock{}
	}
	if request.OutputLimitBytes <= 0 {
		request.OutputLimitBytes = soakCanaryDefaultOutputBytes
	}
	now := request.Clock.Now().UTC()
	phaseStart, _ := time.Parse(time.RFC3339, request.Activation.PhaseStartUTC)
	if now.Before(phaseStart) || now.Sub(phaseStart).Milliseconds() > request.Authority.HardWallMS {
		summary := rejectedSoakCanarySummary(request, []string{"authority_time_window_expired"})
		return summary, errors.New("soak canary authority time window expired")
	}

	checkpoint, err := loadOrCreateSoakCanaryCheckpoint(request)
	if err != nil {
		summary := rejectedSoakCanarySummary(request, []string{"checkpoint_invalid"})
		return summary, err
	}
	if len(checkpoint.CompletedNodeIDs) == len(request.Activation.Partitions) {
		completedAt, err := time.Parse(time.RFC3339Nano, checkpoint.CompletedAtUTC)
		if err != nil {
			return rejectedSoakCanarySummary(request, []string{"checkpoint_completion_time_invalid"}), err
		}
		return ReconcileSoakCanary(request, checkpoint, completedAt)
	}

	commands := map[string]SoakCanaryCommand{}
	for _, command := range request.Catalog.Commands {
		commands[command.TestID] = command
	}
	for _, partition := range request.Activation.Partitions {
		if soakCanaryStringPresent(checkpoint.CompletedNodeIDs, partition.NodeID) {
			continue
		}
		if request.Clock.Now().UTC().Sub(phaseStart).Milliseconds() >= request.Authority.HardWallMS {
			return rejectedSoakCanarySummary(request, []string{"hard_wall_reached"}), errors.New("soak canary hard wall reached")
		}
		command := commands[partition.TestID]
		if partition.NodeID == request.Activation.ControlledRetryNodeID &&
			!checkpoint.ControlledRetryConsumed {
			attempt := newSoakCanaryAttempt(request, checkpoint, partition, command)
			attempt.AttemptNumber = 1
			attempt.OutcomeClass = "transient_infrastructure"
			attempt.StartedAtUTC = request.Clock.Now().UTC().Format(time.RFC3339Nano)
			attempt.CompletedAtUTC = attempt.StartedAtUTC
			attempt.ExitCode = -1
			attempt.Stdout = emptySoakCanaryOutput()
			attempt.Stderr = emptySoakCanaryOutput()
			if err := appendSoakCanaryAttempt(request, &checkpoint, attempt); err != nil {
				return rejectedSoakCanarySummary(request, []string{"checkpoint_write_failed"}), err
			}
		}

		attemptNumber := soakCanaryNodeAttemptCount(checkpoint.Attempts, partition.NodeID) + 1
		if partition.Classification == "scale" && attemptNumber != 1 {
			return rejectedSoakCanarySummary(request, []string{"scale_retry_requested"}), errors.New("soak canary scale retry is prohibited")
		}
		if attemptNumber > 2 {
			return rejectedSoakCanarySummary(request, []string{"second_retry_requested"}), errors.New("soak canary second retry is prohibited")
		}
		attempt := newSoakCanaryAttempt(request, checkpoint, partition, command)
		attempt.AttemptNumber = attemptNumber
		attempt.StartedAtUTC = request.Clock.Now().UTC().Format(time.RFC3339Nano)
		execRequest := SoakCanaryExecRequest{
			TestID: partition.TestID, ExpectedTestName: command.TestName,
			ExpectedPassCount: partition.EffectiveRepeatCount,
			ExecutablePath:    command.ExecutablePath,
			Argv:              append([]string(nil), command.Argv...),
			RepositoryRoot:    request.RepositoryRoot,
			WorkingDirectory:  command.WorkingDirectory,
			Environment:       append([]SoakCanaryEnvironment(nil), command.Environment...),
			TimeoutMS:         partition.PerAttemptTimeoutMS,
			OutputLimitBytes:  request.OutputLimitBytes,
		}
		attemptContext, cancel := context.WithTimeout(ctx, time.Duration(partition.PerAttemptTimeoutMS)*time.Millisecond)
		process, startErr := request.Executor.Start(attemptContext, execRequest)
		if startErr != nil {
			cancel()
			return rejectedSoakCanarySummary(request, []string{"child_process_start_failed"}), startErr
		}
		attempt.ChildProcessLaunched = true
		attempt.ChildPID = process.PID()
		result := process.Wait()
		cancel()
		attempt.CompletedAtUTC = request.Clock.Now().UTC().Format(time.RFC3339Nano)
		attempt.ElapsedMS = result.ElapsedMS
		attempt.ExitCode = result.ExitCode
		attempt.Signal = result.Signal
		stdout, stdoutTruncated := boundSoakCanaryOutput(result.Stdout, request.OutputLimitBytes)
		stderr, stderrTruncated := boundSoakCanaryOutput(result.Stderr, request.OutputLimitBytes)
		stdoutTruncated = stdoutTruncated || result.StdoutTruncated
		stderrTruncated = stderrTruncated || result.StderrTruncated
		attempt.GoTestEvents = parseSoakCanaryGoTestEvents(stdout, command.TestName)
		attempt.Stdout, err = persistSoakCanaryOutput(
			request, partition.NodeID, attemptNumber, "stdout.jsonl", stdout, stdoutTruncated,
		)
		if err != nil {
			return rejectedSoakCanarySummary(request, []string{"stdout_evidence_write_failed"}), err
		}
		attempt.Stderr, err = persistSoakCanaryOutput(
			request, partition.NodeID, attemptNumber, "stderr.txt", stderr, stderrTruncated,
		)
		if err != nil {
			return rejectedSoakCanarySummary(request, []string{"stderr_evidence_write_failed"}), err
		}
		attempt.OutcomeClass = "passed"
		if result.ExitCode != 0 {
			attempt.OutcomeClass = "test_failure"
		}
		if attempt.GoTestEvents.MatchingPasses != partition.EffectiveRepeatCount {
			attempt.OutcomeClass = "go_event_count_mismatch"
		}
		if attempt.ElapsedMS > partition.EstimatedDurationMS {
			attempt.OutcomeClass = "actual_duration_above_estimate"
		}
		if attempt.ElapsedMS > partition.PerAttemptTimeoutMS {
			attempt.OutcomeClass = "per_attempt_timeout_exceeded"
		}
		if err := appendSoakCanaryAttempt(request, &checkpoint, attempt); err != nil {
			return rejectedSoakCanarySummary(request, []string{"checkpoint_write_failed"}), err
		}
		if attempt.OutcomeClass != "passed" {
			return rejectedSoakCanarySummary(request, []string{attempt.OutcomeClass}),
				fmt.Errorf("soak canary node %s failed: %s", partition.NodeID, attempt.OutcomeClass)
		}
	}
	checkpoint.CompletedAtUTC = request.Clock.Now().UTC().Format(time.RFC3339Nano)
	checkpoint.PriorCheckpointDigest = checkpoint.CheckpointDigest
	signSoakCanaryCheckpoint(&checkpoint)
	if err := writeSoakCanaryCheckpoint(request.CheckpointPath, checkpoint); err != nil {
		return rejectedSoakCanarySummary(request, []string{"checkpoint_write_failed"}), err
	}
	return ReconcileSoakCanary(request, checkpoint, request.Clock.Now().UTC())
}

func loadOrCreateSoakCanaryCheckpoint(request SoakCanaryRunRequest) (SoakCanaryCheckpoint, error) {
	checkpoint, err := LoadSoakCanaryCheckpoint(request.CheckpointPath)
	if err == nil {
		if checkpoint.CanaryID != request.Activation.CanaryID ||
			checkpoint.MissionID != request.Activation.MissionID ||
			checkpoint.PlanID != request.Activation.PlanID ||
			checkpoint.PhaseStartUTC != request.Activation.PhaseStartUTC ||
			checkpoint.SourceHead != request.Activation.SourceHead ||
			checkpoint.PlanInputDigest != request.Activation.PlanInputDigest ||
			checkpoint.PolicyDigest != request.Activation.PolicyDigest ||
			checkpoint.CommandCatalogDigest != request.Activation.CommandCatalogDigest ||
			checkpoint.AuthorityRecordDigest != request.Activation.AuthorityRecordDigest ||
			checkpoint.ActivationManifestDigest != request.Activation.ActivationManifestDigest {
			return checkpoint, errors.New("soak canary checkpoint binding mismatch")
		}
		return checkpoint, nil
	}
	if !os.IsNotExist(err) {
		return checkpoint, err
	}
	checkpoint = SoakCanaryCheckpoint{
		Schema:   SoakCanaryCheckpointSchema,
		CanaryID: request.Activation.CanaryID, MissionID: request.Activation.MissionID,
		PlanID: request.Activation.PlanID, PhaseStartUTC: request.Activation.PhaseStartUTC,
		SourceHead:               request.Activation.SourceHead,
		PlanInputDigest:          request.Activation.PlanInputDigest,
		PolicyDigest:             request.Activation.PolicyDigest,
		CommandCatalogDigest:     request.Activation.CommandCatalogDigest,
		AuthorityRecordDigest:    request.Activation.AuthorityRecordDigest,
		ActivationManifestDigest: request.Activation.ActivationManifestDigest,
		Attempts:                 []SoakCanaryAttempt{}, CompletedNodeIDs: []string{},
	}
	signSoakCanaryCheckpoint(&checkpoint)
	if err := writeSoakCanaryCheckpoint(request.CheckpointPath, checkpoint); err != nil {
		return checkpoint, err
	}
	return checkpoint, nil
}

func newSoakCanaryAttempt(
	request SoakCanaryRunRequest,
	checkpoint SoakCanaryCheckpoint,
	partition SoakCanaryPartitionBinding,
	command SoakCanaryCommand,
) SoakCanaryAttempt {
	argvBody, _ := json.Marshal(command.Argv)
	return SoakCanaryAttempt{
		Schema:   SoakCanaryAttemptSchema,
		CanaryID: request.Activation.CanaryID, MissionID: request.Activation.MissionID,
		PlanID: request.Activation.PlanID, PartitionID: partition.PartitionID,
		NodeID: partition.NodeID, TestID: partition.TestID,
		PhaseStartUTC:            request.Activation.PhaseStartUTC,
		SourceHead:               request.Activation.SourceHead,
		PlanInputDigest:          request.Activation.PlanInputDigest,
		PolicyDigest:             request.Activation.PolicyDigest,
		ExecutionProfileDigest:   request.Activation.ExecutionProfileDigest,
		CommandCatalogDigest:     request.Activation.CommandCatalogDigest,
		AuthorityRecordDigest:    request.Activation.AuthorityRecordDigest,
		ActivationManifestDigest: request.Activation.ActivationManifestDigest,
		CommandArgvDigest:        digestBytes(argvBody),
		RequestedRepeatCount:     partition.RequestedRepeatCount,
		EffectiveRepeatCount:     partition.EffectiveRepeatCount,
		Classification:           partition.Classification, ScaleDimension: partition.ScaleDimension,
		CheckpointBeforeDigest:  checkpoint.CheckpointDigest,
		CheckpointAfterSequence: checkpoint.Sequence + 1,
		Safety:                  request.Activation.Safety,
	}
}

func appendSoakCanaryAttempt(
	request SoakCanaryRunRequest,
	checkpoint *SoakCanaryCheckpoint,
	attempt SoakCanaryAttempt,
) error {
	if len(checkpoint.Attempts) >= request.Authority.MaximumAttempts {
		return errors.New("soak canary attempt limit exceeded")
	}
	signSoakCanaryAttempt(&attempt)
	prior := checkpoint.CheckpointDigest
	checkpoint.Attempts = append(checkpoint.Attempts, attempt)
	checkpoint.Sequence++
	checkpoint.PriorCheckpointDigest = prior
	if attempt.OutcomeClass == "transient_infrastructure" {
		if checkpoint.ControlledRetryConsumed {
			return errors.New("soak canary controlled retry was already consumed")
		}
		checkpoint.ControlledRetryConsumed = true
	}
	if attempt.ChildProcessLaunched && attempt.Classification == "scale" {
		if checkpoint.ScaleLaunchConsumed {
			return errors.New("soak canary scale launch was already consumed")
		}
		checkpoint.ScaleLaunchConsumed = true
	}
	if attempt.OutcomeClass == "passed" {
		completed := map[string]bool{}
		for _, nodeID := range checkpoint.CompletedNodeIDs {
			completed[nodeID] = true
		}
		if completed[attempt.NodeID] {
			return errors.New("soak canary duplicate node completion")
		}
		completed[attempt.NodeID] = true
		checkpoint.CompletedNodeIDs = sortedSoakCanaryKeys(completed)
	}
	signSoakCanaryCheckpoint(checkpoint)
	return writeSoakCanaryCheckpoint(request.CheckpointPath, *checkpoint)
}

func persistSoakCanaryOutput(
	request SoakCanaryRunRequest,
	nodeID string,
	attemptNumber int,
	suffix string,
	body []byte,
	truncated bool,
) (SoakCanaryOutputArtifact, error) {
	relative := filepath.Join(
		"nodes", safeSoakCanaryFilename(nodeID),
		"attempt-"+strconv.Itoa(attemptNumber)+"."+suffix,
	)
	path := filepath.Join(request.EvidenceRoot, relative)
	if !pathWithin(request.EvidenceRoot, path) {
		return SoakCanaryOutputArtifact{}, errors.New("soak canary output path escaped evidence root")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return SoakCanaryOutputArtifact{}, err
	}
	if err := writeAtomicFile(path, body, 0o600); err != nil {
		return SoakCanaryOutputArtifact{}, err
	}
	return SoakCanaryOutputArtifact{
		Path: relative, SHA256: digestBytes(body), Bytes: len(body), Truncated: truncated,
	}, nil
}

func parseSoakCanaryGoTestEvents(body []byte, expectedTest string) SoakCanaryGoTestCounts {
	counts := SoakCanaryGoTestCounts{}
	scanner := bufio.NewScanner(bytes.NewReader(body))
	scanner.Buffer(make([]byte, 64*1024), soakCanaryDefaultOutputBytes)
	for scanner.Scan() {
		var event struct {
			Action string `json:"Action"`
			Test   string `json:"Test"`
		}
		if json.Unmarshal(scanner.Bytes(), &event) != nil {
			continue
		}
		counts.TotalEvents++
		if event.Action == "pass" && event.Test == expectedTest {
			counts.MatchingPasses++
		}
	}
	return counts
}

func boundSoakCanaryOutput(body []byte, limit int) ([]byte, bool) {
	if len(body) <= limit {
		return append([]byte(nil), body...), false
	}
	return append([]byte(nil), body[:limit]...), true
}

func emptySoakCanaryOutput() SoakCanaryOutputArtifact {
	return SoakCanaryOutputArtifact{SHA256: digestBytes(nil)}
}

func rejectedSoakCanarySummary(request SoakCanaryRunRequest, conflicts []string) SoakCanarySummary {
	summary := SoakCanarySummary{
		Schema: SoakCanarySummarySchema, Status: "rejected",
		CanaryID: request.Activation.CanaryID, MissionID: request.Activation.MissionID,
		PlanID: request.Activation.PlanID, SourceHead: request.Activation.SourceHead,
		PlanInputDigest:          request.Activation.PlanInputDigest,
		PolicyDigest:             request.Activation.PolicyDigest,
		CommandCatalogDigest:     request.Activation.CommandCatalogDigest,
		AuthorityRecordDigest:    request.Activation.AuthorityRecordDigest,
		ActivationManifestDigest: request.Activation.ActivationManifestDigest,
		PlannedPartitions:        len(request.Activation.Partitions),
		PlannedNodes:             len(request.Activation.Partitions),
		PhaseStartUTC:            request.Activation.PhaseStartUTC,
		LeaseMaximumMS:           request.PlanInput.Lease.MaximumMS,
		ConflictCodes:            append([]string(nil), conflicts...),
		Safety:                   request.Activation.Safety,
	}
	sort.Strings(summary.ConflictCodes)
	signSoakCanarySummary(&summary)
	return summary
}

func soakCanarySanitizedEnvironment(bound []SoakCanaryEnvironment) []string {
	environment := make([]string, 0, len(bound)+8)
	for _, pair := range bound {
		environment = append(environment, pair.Name+"="+pair.Value)
	}
	for _, name := range []string{"HOME", "TMPDIR", "GOCACHE", "GOMODCACHE", "GOENV", "CGO_ENABLED"} {
		if value := os.Getenv(name); value != "" {
			environment = append(environment, name+"="+value)
		}
	}
	if runtime.GOOS == "windows" {
		for _, name := range []string{"SYSTEMROOT", "TEMP", "TMP", "USERPROFILE"} {
			if value := os.Getenv(name); value != "" {
				environment = append(environment, name+"="+value)
			}
		}
	}
	return environment
}

func soakCanaryExitCode(err error) int {
	if err == nil {
		return 0
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return exit.ExitCode()
	}
	return -1
}

func soakCanaryProcessSignal(processState string) string {
	if strings.Contains(strings.ToLower(processState), "signal") {
		return processState
	}
	return ""
}

func soakCanaryNodeAttemptCount(attempts []SoakCanaryAttempt, nodeID string) int {
	count := 0
	for _, attempt := range attempts {
		if attempt.NodeID == nodeID {
			count++
		}
	}
	return count
}

func soakCanaryStringPresent(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func safeSoakCanaryFilename(value string) string {
	var builder strings.Builder
	for _, character := range value {
		switch {
		case character >= 'a' && character <= 'z':
			builder.WriteRune(character)
		case character >= 'A' && character <= 'Z':
			builder.WriteRune(character)
		case character >= '0' && character <= '9':
			builder.WriteRune(character)
		case character == '-', character == '_':
			builder.WriteRune(character)
		default:
			builder.WriteByte('_')
		}
	}
	return builder.String()
}
