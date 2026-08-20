package mission

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestNativeArtifactWorkflowContract(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", "native-artifacts.yml"))
	if err != nil {
		t.Fatal(err)
	}
	workflow := string(data)
	for _, want := range []string{
		"ubuntu-latest",
		"macos-latest",
		"windows-latest",
		"linux-x86_64",
		"macos-aarch64",
		"windows-x86_64",
		"actions/upload-artifact",
		"ao-mission-native-artifact-${{ matrix.target_label }}-${{ github.sha }}",
		"native-artifact-summary.json",
		"SHA256SUMS",
		"LICENSE",
		"NOTICE",
		"./cmd/ao-mission",
		"no-args-usage",
		`"$artifact_dir/$binary" --version > "$artifact_dir/version.txt"`,
		`if [ "$smoke_exit" -ne 1 ]; then`,
		`grep -F "error: usage: ao-mission" "$artifact_dir/smoke.txt"`,
		"contents: read",
		"uses: actions/setup-go@924ae3a1cded613372ab5595356fb5720e22ba16",
		"go-version: '1.26.4'",
		`[ "$(go env GOVERSION)" = "go1.26.4" ]`,
		"ref: 4c501b4f1e55cb9b926709e19d496edf41984fb1",
	} {
		if !strings.Contains(workflow, want) {
			t.Fatalf("native artifact workflow missing %q", want)
		}
	}
	module, err := os.ReadFile(filepath.Join("..", "..", "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(module), "toolchain go1.26.4") {
		t.Fatal("native artifact workflow requires the pinned Go 1.26.4 toolchain")
	}
	for _, forbidden := range []string{"contents: write", "gh release", "actions/create-release", "softprops/action-gh-release"} {
		if strings.Contains(workflow, forbidden) {
			t.Fatalf("native artifact workflow must not include %q", forbidden)
		}
	}

	nativeBuild := strings.Index(workflow, "go build -trimpath")
	policyCheckout := strings.Index(workflow, "repository: uesugitorachiyo/ao-architecture")
	metadataReader := strings.Index(workflow, "scripts/read_go_binary_metadata.go")
	builder := strings.Index(workflow, "scripts/build_go_supply_chain_candidate.py")
	verifier := strings.Index(workflow, "scripts/verify_supply_chain_policy.py")
	if nativeBuild < 0 || policyCheckout < 0 || metadataReader < 0 || builder < 0 || verifier < 0 ||
		!(nativeBuild < policyCheckout && policyCheckout < metadataReader && metadataReader < builder && builder < verifier) {
		t.Fatal("native build, policy checkout, metadata reader, builder, and verifier are required in order")
	}
	hasExactLine := func(section, want string) bool {
		for _, line := range strings.Split(section, "\n") {
			if strings.TrimSpace(line) == want {
				return true
			}
		}
		return false
	}
	if !hasExactLine(workflow[policyCheckout:metadataReader], "ref: 4c501b4f1e55cb9b926709e19d496edf41984fb1") ||
		!hasExactLine(workflow[metadataReader:builder], `"$artifact_dir/$binary" > "$artifact_dir/go-modules.json"`) ||
		!hasExactLine(workflow[builder:verifier], `--workspace-root . \`) ||
		!hasExactLine(workflow[verifier:], `--workspace-root "$supply_chain_dir" \`) {
		t.Fatal("supply-chain policy ref, binary input, and workspace roots must match the approved contract")
	}
	for _, path := range []string{"ci.yml", "release-rehearsal.yml"} {
		data, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", path))
		if err != nil {
			t.Fatal(err)
		}
		workflow := string(data)
		for _, want := range []string{
			"uses: actions/setup-go@924ae3a1cded613372ab5595356fb5720e22ba16",
			"go-version: '1.26.4'",
			`[ "$(go env GOVERSION)" = "go1.26.4" ]`,
		} {
			if !strings.Contains(workflow, want) {
				t.Fatalf("%s missing explicit Go toolchain contract %q", path, want)
			}
		}
	}
}

func TestCIWorkflowRunsForCandidateBranches(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", "ci.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if blockers := ciWorkflowContractBlockers(string(data)); len(blockers) != 0 {
		t.Fatalf("CI workflow does not cover candidate branches: %v", blockers)
	}
}

func TestCIWorkflowMissingDiagnosticHelperFails(t *testing.T) {
	pwsh, err := exec.LookPath("pwsh")
	if err != nil {
		t.Skip("pwsh is required to exercise the cross-platform CI wrapper")
	}
	data, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", "ci.yml"))
	if err != nil {
		t.Fatal(err)
	}
	script, ok := workflowDiagnosticScript(string(data))
	if !ok {
		t.Fatal("missing bounded diagnostic workflow script")
	}
	script = strings.Replace(script, "go build -o $helper ./scripts/ci-go-test", "$global:LASTEXITCODE = 0", 1)
	command := exec.Command(pwsh, "-NoProfile", "-NonInteractive", "-Command", script)
	command.Env = append(os.Environ(), "RUNNER_TEMP="+t.TempDir())
	if err := command.Run(); err == nil {
		t.Fatal("missing diagnostic helper returned exit 0")
	}
}

func TestCIWorkflowMissingDiagnosticBuildCommandFails(t *testing.T) {
	pwsh, err := exec.LookPath("pwsh")
	if err != nil {
		t.Skip("pwsh is required to exercise the cross-platform CI wrapper")
	}
	script := readWorkflowDiagnosticScript(t)
	script = strings.Replace(script, "go build -o $helper ./scripts/ci-go-test", "definitely-missing-ao-mission-build-command", 1)
	command := exec.Command(pwsh, "-NoProfile", "-NonInteractive", "-Command", script)
	command.Env = append(os.Environ(), "RUNNER_TEMP="+t.TempDir())
	if err := command.Run(); err == nil {
		t.Fatal("missing diagnostic build command returned exit 0")
	}
}

func TestCIWorkflowPreservesDiagnosticBuildExit(t *testing.T) {
	pwsh, err := exec.LookPath("pwsh")
	if err != nil {
		t.Skip("pwsh is required to exercise the cross-platform CI wrapper")
	}
	script := readWorkflowDiagnosticScript(t)
	script = strings.Replace(script, "go build -o $helper ./scripts/ci-go-test", "& pwsh -NoProfile -NonInteractive -Command 'exit 6'", 1)
	command := exec.Command(pwsh, "-NoProfile", "-NonInteractive", "-Command", script)
	command.Env = append(os.Environ(), "RUNNER_TEMP="+t.TempDir())
	err = command.Run()
	var exitError *exec.ExitError
	if !errors.As(err, &exitError) || exitError.ExitCode() != 6 {
		t.Fatalf("diagnostic build wrapper error = %v, want exit 6", err)
	}
}

func TestCIWorkflowPreservesDiagnosticHelperExit(t *testing.T) {
	pwsh, err := exec.LookPath("pwsh")
	if err != nil {
		t.Skip("pwsh is required to exercise the cross-platform CI wrapper")
	}
	data, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", "ci.yml"))
	if err != nil {
		t.Fatal(err)
	}
	script, ok := workflowDiagnosticScript(string(data))
	if !ok {
		t.Fatal("missing bounded diagnostic workflow script")
	}
	script = strings.Replace(script, "go build -o $helper ./scripts/ci-go-test", "$helper = (Get-Command pwsh).Source; $global:LASTEXITCODE = 0", 1)
	script = strings.Replace(script, "& $helper", "& $helper -NoProfile -NonInteractive -Command 'exit 7'", 1)
	command := exec.Command(pwsh, "-NoProfile", "-NonInteractive", "-Command", script)
	command.Env = append(os.Environ(), "RUNNER_TEMP="+t.TempDir())
	err = command.Run()
	var exitError *exec.ExitError
	if !errors.As(err, &exitError) || exitError.ExitCode() != 7 {
		t.Fatalf("diagnostic wrapper error = %v, want exit 7", err)
	}
}

func readWorkflowDiagnosticScript(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", "ci.yml"))
	if err != nil {
		t.Fatal(err)
	}
	script, ok := workflowDiagnosticScript(string(data))
	if !ok {
		t.Fatal("missing bounded diagnostic workflow script")
	}
	return script
}

func TestCIWorkflowCandidateBranchContractRejectsDecoys(t *testing.T) {
	for name, workflow := range map[string]string{
		"comment decoys":        "on:\n  pull_request:\n# push: branches main 'codex/**'\njobs:\n  test:\n    runs-on: ubuntu-latest\n# windows-latest\n",
		"unrelated branch list": "on:\n  push:\n    branches:\n      - main\n  workflow_dispatch:\n    inputs:\n      codex/**: {}\njobs:\n  test:\n    strategy:\n      matrix:\n        os: [ubuntu-latest]\n    runs-on: windows-latest\n",
		"unrelated windows job": "on:\n  push:\n    branches:\n      - main\n      - 'codex/**'\njobs:\n  docs:\n    runs-on: windows-latest\n  test:\n    strategy:\n      matrix:\n        os: [ubuntu-latest]\n",
		"diagnostic decoys":     "on:\n  push:\n    branches:\n      - main\n      - 'codex/**'\njobs:\n  test:\n    strategy:\n      fail-fast: true\n      matrix:\n        os: [ubuntu-latest, windows-latest]\n    steps:\n      # shell: pwsh; go build ./scripts/ci-go-test; exit $LASTEXITCODE\n      - run: go test ./... -count=1\n  docs:\n    strategy:\n      fail-fast: false\n    steps:\n      - shell: pwsh\n        run: go build ./scripts/ci-go-test\n",
	} {
		t.Run(name, func(t *testing.T) {
			if blockers := ciWorkflowContractBlockers(workflow); len(blockers) == 0 {
				t.Fatal("decoy workflow passed the candidate-branch Windows matrix contract")
			}
		})
	}
}

func ciWorkflowContractBlockers(workflow string) []string {
	blockers := make([]string, 0)
	lines := workflowYAMLLines(workflow)
	onStart, onEnd, ok := workflowYAMLSection(lines, 0, len(lines), -2, "on")
	if !ok {
		return append(blockers, "missing on section")
	}
	pushStart, pushEnd, ok := workflowYAMLSection(lines, onStart, onEnd, 0, "push")
	if !ok {
		blockers = append(blockers, "missing on.push section")
	} else if branchStart, branchEnd, found := workflowYAMLSection(lines, pushStart, pushEnd, 2, "branches"); !found {
		blockers = append(blockers, "missing on.push.branches")
	} else {
		branches := workflowYAMLList(lines, branchStart, branchEnd, 4)
		for _, want := range []string{"main", "codex/**"} {
			if _, found := branches[want]; !found {
				blockers = append(blockers, "missing on.push.branches "+want)
			}
		}
	}

	jobsStart, jobsEnd, ok := workflowYAMLSection(lines, 0, len(lines), -2, "jobs")
	if !ok {
		return append(blockers, "missing jobs section")
	}
	testStart, testEnd, ok := workflowYAMLSection(lines, jobsStart, jobsEnd, 0, "test")
	strategyStart, strategyEnd, strategyOK := workflowYAMLSection(lines, testStart, testEnd, 2, "strategy")
	if value, found := workflowYAMLScalar(lines, strategyStart, strategyEnd, 4, "fail-fast"); !found || value != "false" {
		blockers = append(blockers, "jobs.test.strategy.fail-fast must be false")
	}
	matrixStart, matrixEnd, matrixOK := workflowYAMLSection(lines, strategyStart, strategyEnd, 4, "matrix")
	osValues, osOK := workflowYAMLInlineList(lines, matrixStart, matrixEnd, 6, "os")
	if !ok || !strategyOK || !matrixOK || !osOK {
		blockers = append(blockers, "missing jobs.test.strategy.matrix.os")
	} else if _, found := osValues["windows-latest"]; !found {
		blockers = append(blockers, "missing windows-latest in jobs.test.strategy.matrix.os")
	}
	if !workflowYAMLHasDiagnosticTestStep(lines, testStart, testEnd) {
		blockers = append(blockers, "missing bounded diagnostic go-test pwsh step")
	}
	return blockers
}

type workflowYAMLLine struct {
	indent int
	text   string
}

func workflowYAMLLines(workflow string) []workflowYAMLLine {
	lines := make([]workflowYAMLLine, 0)
	for _, raw := range strings.Split(workflow, "\n") {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		lines = append(lines, workflowYAMLLine{
			indent: len(raw) - len(strings.TrimLeft(raw, " ")),
			text:   trimmed,
		})
	}
	return lines
}

func workflowYAMLSection(
	lines []workflowYAMLLine,
	start, end, parentIndent int,
	key string,
) (int, int, bool) {
	indent := parentIndent + 2
	for i := start; i < end; i++ {
		if lines[i].indent != indent || lines[i].text != key+":" {
			continue
		}
		sectionEnd := end
		for j := i + 1; j < end; j++ {
			if lines[j].indent <= indent {
				sectionEnd = j
				break
			}
		}
		return i + 1, sectionEnd, true
	}
	return 0, 0, false
}

func workflowYAMLList(lines []workflowYAMLLine, start, end, parentIndent int) map[string]struct{} {
	values := make(map[string]struct{})
	for i := start; i < end; i++ {
		if lines[i].indent != parentIndent+2 || !strings.HasPrefix(lines[i].text, "- ") {
			continue
		}
		value := strings.Trim(strings.TrimSpace(strings.TrimPrefix(lines[i].text, "- ")), "'\"")
		values[value] = struct{}{}
	}
	return values
}

func workflowYAMLInlineList(
	lines []workflowYAMLLine,
	start, end, parentIndent int,
	key string,
) (map[string]struct{}, bool) {
	prefix := key + ":"
	for i := start; i < end; i++ {
		if lines[i].indent != parentIndent+2 || !strings.HasPrefix(lines[i].text, prefix) {
			continue
		}
		raw := strings.TrimSpace(strings.TrimPrefix(lines[i].text, prefix))
		if len(raw) < 2 || raw[0] != '[' || raw[len(raw)-1] != ']' {
			return nil, false
		}
		values := make(map[string]struct{})
		for _, value := range strings.Split(raw[1:len(raw)-1], ",") {
			values[strings.Trim(strings.TrimSpace(value), "'\"")] = struct{}{}
		}
		return values, true
	}
	return nil, false
}

func workflowYAMLScalar(
	lines []workflowYAMLLine,
	start, end, parentIndent int,
	key string,
) (string, bool) {
	prefix := key + ":"
	for i := start; i < end; i++ {
		if lines[i].indent != parentIndent+2 || !strings.HasPrefix(lines[i].text, prefix) {
			continue
		}
		return strings.Trim(strings.TrimSpace(strings.TrimPrefix(lines[i].text, prefix)), "'\""), true
	}
	return "", false
}

func workflowYAMLHasDiagnosticTestStep(lines []workflowYAMLLine, start, end int) bool {
	stepsStart, stepsEnd, ok := workflowYAMLSection(lines, start, end, 2, "steps")
	if !ok {
		return false
	}
	for i := stepsStart; i < stepsEnd; i++ {
		if lines[i].indent != 6 || lines[i].text != "- name: Run test suite with bounded diagnostics" {
			continue
		}
		stepEnd := stepsEnd
		for j := i + 1; j < stepsEnd; j++ {
			if lines[j].indent <= 6 {
				stepEnd = j
				break
			}
		}
		stepLines := make([]string, 0, stepEnd-i)
		for _, line := range lines[i+1 : stepEnd] {
			if line.indent == 8 && (line.text == "shell: pwsh" || line.text == "run: |") {
				stepLines = append(stepLines, line.text)
			}
			if line.indent == 10 {
				stepLines = append(stepLines, line.text)
			}
		}
		ordered := []string{
			"shell: pwsh",
			"run: |",
			"$helper = Join-Path $env:RUNNER_TEMP 'ao-mission-ci-go-test'",
			"if ($IsWindows) { $helper += '.exe' }",
			"$buildExit = $null",
			"$global:LASTEXITCODE = $null",
			"try {",
			"go build -o $helper ./scripts/ci-go-test",
			"$buildSucceeded = $?",
			"$buildExit = $LASTEXITCODE",
			"} catch {",
			"Write-Error 'CI test helper build failed to start'",
			"exit 1",
			"}",
			"if ($null -eq $buildExit) { exit 1 }",
			"if ($buildSucceeded -ne ($buildExit -eq 0)) { exit 1 }",
			"if ($buildExit -ne 0) { exit $buildExit }",
			"$childExit = $null",
			"$global:LASTEXITCODE = $null",
			"try {",
			"& $helper",
			"$commandSucceeded = $?",
			"$childExit = $LASTEXITCODE",
			"} catch {",
			"Write-Error 'CI test helper failed to launch'",
			"exit 1",
			"}",
			"if ($null -eq $childExit) { exit 1 }",
			"if ($commandSucceeded -ne ($childExit -eq 0)) { exit 1 }",
			"exit $childExit",
		}
		cursor := 0
		for _, line := range stepLines {
			if cursor < len(ordered) && line == ordered[cursor] {
				cursor++
			}
		}
		return cursor == len(ordered)
	}
	return false
}

func workflowDiagnosticScript(workflow string) (string, bool) {
	lines := strings.Split(workflow, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) != "- name: Run test suite with bounded diagnostics" {
			continue
		}
		for j := i + 1; j < len(lines); j++ {
			if strings.TrimSpace(lines[j]) != "run: |" {
				continue
			}
			var script []string
			for k := j + 1; k < len(lines); k++ {
				if strings.HasPrefix(lines[k], "          ") {
					script = append(script, strings.TrimPrefix(lines[k], "          "))
					continue
				}
				break
			}
			return strings.Join(script, "\n"), len(script) != 0
		}
	}
	return "", false
}
