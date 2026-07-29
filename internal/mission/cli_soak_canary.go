package mission

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func runSoakCanaryCLI(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("qualification soak-canary", flag.ContinueOnError)
	planPath := fs.String("plan", "", "")
	authorityPath := fs.String("authority", "", "")
	catalogPath := fs.String("catalog", "", "")
	activationPath := fs.String("activation", "", "")
	checkpointPath := fs.String("checkpoint", "", "")
	evidenceRoot := fs.String("evidence-root", "", "")
	repositoryRoot := fs.String("repository-root", "", "")
	validateOnly := fs.Bool("validate-only", false, "")
	jsonOut := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 || soakCanaryCLIValueMissing(
		*planPath, *authorityPath, *catalogPath, *activationPath,
		*checkpointPath, *evidenceRoot, *repositoryRoot,
	) {
		return errors.New("qualification soak-canary requires --plan, --authority, --catalog, --activation, --checkpoint, --evidence-root, and --repository-root")
	}
	rootAbs, err := filepath.Abs(*repositoryRoot)
	if err != nil {
		return err
	}
	repositoryVerifier, err := newPinnedGitSoakCanaryRepositoryVerifier()
	if err != nil {
		return err
	}
	verifiedHead, err := repositoryVerifier.Head(rootAbs)
	if err != nil {
		return err
	}
	if !*validateOnly {
		clean, err := repositoryVerifier.Clean(rootAbs)
		if err != nil {
			return err
		}
		if !clean {
			return errors.New("qualification soak-canary requires a clean repository before execution")
		}
	}
	planBody, err := readBoundedRegularFile(*planPath, soakCanaryMaxInputBytes)
	if err != nil {
		return err
	}
	input, err := decodeSoakCanaryJSON[SoakPlanInput](planBody)
	if err != nil {
		return err
	}
	plan, err := BuildSoakPlan(input)
	if err != nil {
		return err
	}
	authority, err := LoadSoakCanaryAuthority(*authorityPath)
	if err != nil {
		return err
	}
	catalog, err := LoadSoakCanaryCommandCatalog(*catalogPath)
	if err != nil {
		return err
	}
	activation, err := LoadSoakCanaryActivation(*activationPath)
	if err != nil {
		return err
	}
	request := SoakCanaryRunRequest{
		PlanInput: input, PlanReadback: plan, PlanFixtureSHA256: digestBytes(planBody),
		Authority: authority, Catalog: catalog, Activation: activation,
		VerifiedSourceHead: verifiedHead, RepositoryRoot: rootAbs,
		EvidenceRoot: *evidenceRoot, CheckpointPath: *checkpointPath,
		OutputLimitBytes:   soakCanaryDefaultOutputBytes,
		RepositoryVerifier: repositoryVerifier,
	}
	validation := ValidateSoakCanaryActivation(request)
	if *validateOnly {
		if *jsonOut {
			if err := printJSON(stdout, validation); err != nil {
				return err
			}
		} else {
			fmt.Fprintf(stdout, "activation_allowed=%t\nplanner_activation_eligible=%t\ncanary_execution_authorized=%t\nchild_process_launches=0\nconflicts=%s\nnext=%s\n",
				validation.ActivationAllowed,
				validation.PlannerActivationEligible,
				validation.CanaryExecutionAuthorized,
				strings.Join(validation.ConflictCodes, ","),
				validation.ExactNextAction,
			)
		}
		if !validation.ActivationAllowed {
			return fmt.Errorf("soak canary activation rejected: %s", strings.Join(validation.ConflictCodes, ","))
		}
		return nil
	}
	request.Executor = OSExecSoakCanaryExecutor{}
	summary, err := RunSoakCanary(context.Background(), request)
	if *jsonOut {
		if printErr := printJSON(stdout, summary); printErr != nil {
			return printErr
		}
	}
	if err != nil {
		return err
	}
	if err := PersistSoakCanaryCompletion(request.EvidenceRoot, summary); err != nil {
		return err
	}
	if !*jsonOut {
		fmt.Fprintf(stdout, "status=%s\nnodes=%d/%d\nattempts=%d\nchild_process_launches=%d\nscale_launches=%d\ncontrolled_retries=%d\nlocal_test_execution_performed=%t\nsummary_digest=%s\n",
			summary.Status, summary.CompletedNodes, summary.PlannedNodes,
			summary.TotalAttempts, summary.ChildProcessLaunches,
			summary.ScaleLaunches, summary.ControlledRetryCount,
			summary.LocalTestExecutionPerformed, summary.SummaryDigest,
		)
	}
	return nil
}

func verifySoakCanaryRepositoryHead(root string) (string, error) {
	verifier, err := newPinnedGitSoakCanaryRepositoryVerifier()
	if err != nil {
		return "", err
	}
	return verifier.Head(root)
}

func verifySoakCanaryRepositoryClean(root string) (bool, error) {
	verifier, err := newPinnedGitSoakCanaryRepositoryVerifier()
	if err != nil {
		return false, err
	}
	return verifier.Clean(root)
}

type pinnedGitSoakCanaryRepositoryVerifier struct {
	executablePath   string
	executableSHA256 string
}

func newPinnedGitSoakCanaryRepositoryVerifier() (*pinnedGitSoakCanaryRepositoryVerifier, error) {
	candidates := []string{"/usr/bin/git", "/opt/homebrew/bin/git", "/usr/local/bin/git"}
	if runtime.GOOS == "windows" {
		candidates = []string{
			`C:\Program Files\Git\cmd\git.exe`,
			`C:\Program Files\Git\bin\git.exe`,
		}
	}
	for _, path := range candidates {
		info, err := os.Lstat(path)
		if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
			continue
		}
		body, err := readBoundedRegularFile(path, 256<<20)
		if err != nil {
			continue
		}
		return &pinnedGitSoakCanaryRepositoryVerifier{
			executablePath: path, executableSHA256: digestBytes(body),
		}, nil
	}
	return nil, errors.New("no fixed absolute Git executable is available for soak canary verification")
}

func (verifier *pinnedGitSoakCanaryRepositoryVerifier) Verify(root, expectedHead string) error {
	head, err := verifier.Head(root)
	if err != nil {
		return err
	}
	if head != expectedHead {
		return fmt.Errorf("soak canary repository head=%s want=%s", head, expectedHead)
	}
	clean, err := verifier.Clean(root)
	if err != nil {
		return err
	}
	if !clean {
		return errors.New("soak canary repository is not clean")
	}
	return nil
}

func (verifier *pinnedGitSoakCanaryRepositoryVerifier) Head(root string) (string, error) {
	body, err := verifier.run(root, "rev-parse", "HEAD")
	if err != nil {
		return "", fmt.Errorf("verify soak canary repository head: %w", err)
	}
	head := strings.TrimSpace(string(body))
	if !validSoakHexDigest(head, 40, "") {
		return "", errors.New("soak canary repository head is invalid")
	}
	return head, nil
}

func (verifier *pinnedGitSoakCanaryRepositoryVerifier) Clean(root string) (bool, error) {
	body, err := verifier.run(root, "status", "--porcelain")
	if err != nil {
		return false, fmt.Errorf("verify soak canary repository cleanliness: %w", err)
	}
	return len(strings.TrimSpace(string(body))) == 0, nil
}

func (verifier *pinnedGitSoakCanaryRepositoryVerifier) run(root string, args ...string) ([]byte, error) {
	body, err := readBoundedRegularFile(verifier.executablePath, 256<<20)
	if err != nil || digestBytes(body) != verifier.executableSHA256 {
		return nil, errors.New("pinned Git executable provenance changed")
	}
	commandArgs := append([]string{"-C", root}, args...)
	command := exec.Command(verifier.executablePath, commandArgs...)
	fixedPath := "/usr/bin:/bin"
	command.Env = []string{
		"PATH=" + fixedPath,
		"LC_ALL=C",
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_CONFIG_GLOBAL=" + os.DevNull,
	}
	if runtime.GOOS == "windows" {
		systemRoot := filepath.Join(
			filepath.VolumeName(verifier.executablePath)+string(os.PathSeparator), "Windows",
		)
		fixedPath = strings.Join(
			[]string{filepath.Dir(verifier.executablePath), filepath.Join(systemRoot, "System32")},
			string(os.PathListSeparator),
		)
		command.Env[0] = "PATH=" + fixedPath
		command.Env = append(command.Env, "SYSTEMROOT="+systemRoot)
	}
	return command.Output()
}

func soakCanaryCLIValueMissing(values ...string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return true
		}
	}
	return false
}
