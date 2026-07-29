package mission

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func TestSoakCanaryActivationRejectsUncleanOrMismatchedGitRepository(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*testing.T, *soakCanaryTestFixture, *git.Worktree)
	}{
		{
			name: "tracked modification",
			mutate: func(t *testing.T, fixture *soakCanaryTestFixture, _ *git.Worktree) {
				mustWriteSoakCanaryTestFile(
					t,
					filepath.Join(fixture.request.RepositoryRoot, "tracked.txt"),
					[]byte("modified\n"),
				)
			},
		},
		{
			name: "staged modification",
			mutate: func(t *testing.T, fixture *soakCanaryTestFixture, worktree *git.Worktree) {
				mustWriteSoakCanaryTestFile(
					t,
					filepath.Join(fixture.request.RepositoryRoot, "tracked.txt"),
					[]byte("staged\n"),
				)
				if _, err := worktree.Add("tracked.txt"); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "untracked file",
			mutate: func(t *testing.T, fixture *soakCanaryTestFixture, _ *git.Worktree) {
				mustWriteSoakCanaryTestFile(
					t,
					filepath.Join(fixture.request.RepositoryRoot, "untracked.txt"),
					[]byte("untracked\n"),
				)
			},
		},
		{
			name: "mismatched HEAD",
			mutate: func(t *testing.T, fixture *soakCanaryTestFixture, _ *git.Worktree) {
				fixture.request.Activation.SourceHead = strings.Repeat("f", 40)
				signSoakCanaryActivation(&fixture.request.Activation)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := validSoakCanaryFixture(t)
			worktree, head := initializeSoakCanaryGitRepository(t, fixture.request.RepositoryRoot)
			snapshot, err := BuildSoakCanaryRepositorySnapshot(fixture.request.RepositoryRoot)
			if err != nil {
				t.Fatal(err)
			}
			fixture.request.RepositorySnapshot = snapshot
			rebindSoakCanarySource(t, &fixture, head)
			fixture.request.GitVerifier = InProcessSoakCanaryGitVerifier{}

			clean := ValidateSoakCanaryActivation(fixture.request)
			if !clean.ActivationAllowed {
				t.Fatalf("clean Git repository rejected: %+v", clean)
			}

			test.mutate(t, &fixture, worktree)
			validation := ValidateSoakCanaryActivation(fixture.request)
			if validation.ActivationAllowed ||
				!containsSoakConflict(validation.ConflictCodes, "repository_git_state_mismatch") {
				t.Fatalf("unsafe Git state accepted: %+v", validation)
			}
		})
	}
}

func TestSoakCanaryGitVerifierSupportsGitFile(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "worktree")
	if err := os.Mkdir(root, 0o700); err != nil {
		t.Fatal(err)
	}
	_, head := initializeSoakCanaryGitRepository(t, root)
	gitDirectory := filepath.Join(parent, "metadata")
	if err := os.Rename(filepath.Join(root, ".git"), gitDirectory); err != nil {
		t.Fatal(err)
	}
	mustWriteSoakCanaryTestFile(
		t,
		filepath.Join(root, ".git"),
		[]byte("gitdir: "+gitDirectory+"\n"),
	)

	if err := (InProcessSoakCanaryGitVerifier{}).Verify(root, head); err != nil {
		t.Fatalf("valid Git file repository rejected: %v", err)
	}
}

func TestSoakCanaryRunningCheckpointFailureReapsChildAndRestartFailsClosed(t *testing.T) {
	fixture := validSoakCanaryFixture(t)
	clock := &soakCanaryFakeClock{now: time.Date(2026, 7, 29, 21, 0, 0, 0, time.UTC)}
	process := &soakCanaryReapTrackingProcess{
		clock: clock,
		result: SoakCanaryProcessResult{
			ExitCode: 0, ElapsedMS: 5,
			Stdout: soakCanaryPassOutput(
				fixture.request.Catalog.Commands[0].TestName,
				fixture.request.Catalog.Commands[0].EffectiveRepeatCount,
			),
		},
	}
	var reservedCheckpoint []byte
	fixture.request.Clock = clock
	fixture.request.Executor = &soakCanaryCheckpointFailureExecutor{
		process: process,
		beforeReturn: func() {
			var err error
			reservedCheckpoint, err = os.ReadFile(fixture.request.CheckpointPath)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.Remove(fixture.request.CheckpointPath); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(t.TempDir(), fixture.request.CheckpointPath); err != nil {
				t.Fatal(err)
			}
		},
	}

	summary, err := RunSoakCanary(context.Background(), fixture.request)
	if err == nil || !containsSoakConflict(summary.ConflictCodes, "checkpoint_write_failed") {
		t.Fatalf("running checkpoint failure error=%v summary=%+v", err, summary)
	}
	if !process.waited || !process.canceledBeforeWait {
		t.Fatalf(
			"RunSoakCanary did not cancel then synchronously reap: waited=%t canceled_before_wait=%t",
			process.waited,
			process.canceledBeforeWait,
		)
	}
	if summary.ChildProcessLaunches != 1 || len(summary.Attempts) != 1 ||
		!summary.Attempts[0].ChildProcessLaunched ||
		summary.Attempts[0].ExecutionState != "completed" ||
		summary.Attempts[0].OutcomeClass != "running_checkpoint_write_failed_reaped" {
		t.Fatalf("observed child truth was lost: %+v", summary)
	}

	if err := os.Remove(fixture.request.CheckpointPath); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fixture.request.CheckpointPath, reservedCheckpoint, 0o600); err != nil {
		t.Fatal(err)
	}
	restartExecutor := &soakCanaryFakeExecutor{clock: clock}
	fixture.request.Executor = restartExecutor
	restart, restartErr := RunSoakCanary(context.Background(), fixture.request)
	if restartErr == nil ||
		!containsSoakConflict(restart.ConflictCodes, "indeterminate_scale_launch") {
		t.Fatalf("restart escaped durable reservation: error=%v summary=%+v", restartErr, restart)
	}
	if restartExecutor.starts != 0 {
		t.Fatalf("restart launched duplicate/retry child: %d", restartExecutor.starts)
	}
}

func initializeSoakCanaryGitRepository(t *testing.T, root string) (*git.Worktree, string) {
	t.Helper()
	repository, err := git.PlainInit(root, false)
	if err != nil {
		t.Fatal(err)
	}
	mustWriteSoakCanaryTestFile(t, filepath.Join(root, "tracked.txt"), []byte("tracked\n"))
	worktree, err := repository.Worktree()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := worktree.Add("."); err != nil {
		t.Fatal(err)
	}
	hash, err := worktree.Commit("fixture", &git.CommitOptions{
		Author: &object.Signature{
			Name: "AO Mission Test", Email: "ao-mission@example.invalid",
			When: time.Unix(1_700_000_000, 0).UTC(),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return worktree, hash.String()
}

type soakCanaryCheckpointFailureExecutor struct {
	process      SoakCanaryProcess
	beforeReturn func()
}

func (executor *soakCanaryCheckpointFailureExecutor) Start(
	ctx context.Context,
	_ SoakCanaryExecRequest,
) (SoakCanaryProcess, error) {
	if executor.process == nil {
		return nil, errors.New("missing reap-tracking process")
	}
	if process, ok := executor.process.(*soakCanaryReapTrackingProcess); ok {
		process.ctx = ctx
	}
	if executor.beforeReturn != nil {
		executor.beforeReturn()
	}
	return executor.process, nil
}

type soakCanaryReapTrackingProcess struct {
	clock              *soakCanaryFakeClock
	result             SoakCanaryProcessResult
	waited             bool
	ctx                context.Context
	canceledBeforeWait bool
}

func (*soakCanaryReapTrackingProcess) PID() int { return 43_001 }

func (process *soakCanaryReapTrackingProcess) Wait() SoakCanaryProcessResult {
	process.waited = true
	if process.ctx != nil {
		select {
		case <-process.ctx.Done():
			process.canceledBeforeWait = true
		default:
		}
	}
	if process.clock != nil {
		process.clock.advance(time.Duration(process.result.ElapsedMS) * time.Millisecond)
	}
	return process.result
}
