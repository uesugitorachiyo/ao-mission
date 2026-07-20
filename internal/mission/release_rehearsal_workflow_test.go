package mission

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"
)

type workflowShape struct {
	events         map[string]bool
	inputs         map[string]map[string]string
	topPermissions map[string]string
	jobs           map[string]*workflowJob
}

type workflowJob struct {
	condition   string
	environment string
	permissions map[string]string
	uses        []string
}

type releaseVerifierFixture struct {
	candidatesPath string
	environment    map[string]string
	manifestPath   string
	planChecksum   string
	planPath       string
}

type releaseTargetFixture struct {
	archive      string
	architecture string
	binaryFormat string
	entryPoint   string
	goarch       string
	goos         string
	machine      string
	os           string
	runnerArch   string
	runnerLabel  string
	runnerOS     string
	targetLabel  string
}

func TestReleaseRehearsalWorkflowStructure(t *testing.T) {
	workflow := readReleaseWorkflow(t)
	shape := parseWorkflowShape(t, workflow)

	if len(shape.events) != 1 || !shape.events["workflow_dispatch"] {
		t.Fatalf("release workflow events=%v, want workflow_dispatch only", shape.events)
	}
	for _, input := range []string{
		"version",
		"tag",
		"source_sha",
		"approved_manifest_digest",
		"approved_manifest_base64",
		"release_notes",
		"dry_run",
		"live_confirmation",
	} {
		if _, ok := shape.inputs[input]; !ok {
			t.Fatalf("release workflow missing input %q", input)
		}
	}
	if shape.inputs["dry_run"]["default"] != "true" || shape.inputs["dry_run"]["type"] != "boolean" {
		t.Fatalf("dry_run input=%v, want boolean default true", shape.inputs["dry_run"])
	}
	if shape.topPermissions["contents"] != "read" {
		t.Fatalf("top-level permissions=%v, want contents read", shape.topPermissions)
	}

	wantJobs := []string{
		"bind-release-inputs",
		"native-candidates",
		"assemble-promotion-plan",
		"publish-release",
		"verify-published-release",
	}
	for _, name := range wantJobs {
		if shape.jobs[name] == nil {
			t.Fatalf("release workflow missing job %q", name)
		}
	}
	for name, job := range shape.jobs {
		want := "read"
		if name == "publish-release" {
			want = "write"
		}
		if job.permissions["contents"] != want {
			t.Fatalf("job %s contents permission=%q, want %q", name, job.permissions["contents"], want)
		}
	}

	publisher := shape.jobs["publish-release"]
	wantCondition := "${{ inputs.dry_run == false && inputs.live_confirmation == format('publish-ao-mission-{0}-{1}-{2}-{3}', inputs.version, inputs.tag, inputs.source_sha, inputs.approved_manifest_digest) }}"
	if publisher.condition != wantCondition {
		t.Fatalf("publisher condition=%q, want %q", publisher.condition, wantCondition)
	}
	if publisher.environment != "ao-mission-release" {
		t.Fatalf("publisher environment=%q, want ao-mission-release", publisher.environment)
	}
	if publisher.permissions["actions"] != "read" {
		t.Fatalf("publisher actions permission=%q, want read for environment API readback", publisher.permissions["actions"])
	}
	if _, ok := publisher.permissions["deployments"]; ok {
		t.Fatalf("publisher has unnecessary deployments permission: %v", publisher.permissions)
	}

	actionPin := regexp.MustCompile(`^actions/[a-z0-9-]+@[0-9a-f]{40}$`)
	actionCount := 0
	for jobName, job := range shape.jobs {
		for _, use := range job.uses {
			actionCount++
			if !actionPin.MatchString(use) {
				t.Fatalf("job %s action is not pinned to a full commit SHA: %q", jobName, use)
			}
		}
	}
	if actionCount < 10 {
		t.Fatalf("parsed only %d action uses, want all workflow actions", actionCount)
	}
}

func TestApprovedManifestDecoderRejectsMalformedAndOversizedInput(t *testing.T) {
	decoder := extractPythonBlock(t, readReleaseWorkflow(t), "approved-manifest-decoder")
	run := func(t *testing.T, encoded string) ([]byte, error) {
		t.Helper()
		path := filepath.Join(t.TempDir(), "manifest.json")
		_, err := runPythonBlock(t, decoder, []string{path}, map[string]string{
			"APPROVED_MANIFEST_BASE64": encoded,
		})
		if err != nil {
			return nil, err
		}
		raw, readErr := os.ReadFile(path)
		return raw, readErr
	}

	valid := []byte(`{"schema_version":"test"}`)
	raw, err := run(t, base64.StdEncoding.EncodeToString(valid))
	if err != nil {
		t.Fatalf("valid bounded manifest rejected: %v", err)
	}
	if !bytes.Equal(raw, valid) {
		t.Fatalf("decoded manifest=%q, want exact bytes %q", raw, valid)
	}
	if _, err := run(t, "%%%"); err == nil {
		t.Fatal("malformed base64 manifest accepted")
	}
	oversized := bytes.Repeat([]byte("a"), 32769)
	if _, err := run(t, base64.StdEncoding.EncodeToString(oversized)); err == nil {
		t.Fatal("oversized decoded manifest accepted")
	}
}

func TestReleaseManifestValidatorRejectsMissingMalformedAndSubstitutedInputs(t *testing.T) {
	workflow := readReleaseWorkflow(t)
	validator := extractPythonBlock(t, workflow, "release-manifest-validator")
	dir := t.TempDir()
	versionSource := filepath.Join(dir, "AO-MISSION-V0.1.md")
	versionSourceBytes := []byte("# AO Mission v0.1 SDD\n")
	if err := os.WriteFile(versionSource, versionSourceBytes, 0o644); err != nil {
		t.Fatal(err)
	}
	sourceSHA := strings.Repeat("a", 40)
	versionSourceDigest := sha256Hex(versionSourceBytes)
	valid := releaseManifestFixture(sourceSHA, versionSourceDigest)

	run := func(t *testing.T, raw []byte, digest, path string) error {
		t.Helper()
		outputs := filepath.Join(t.TempDir(), "outputs")
		_, err := runPythonBlock(t, validator, []string{path, versionSource, outputs}, map[string]string{
			"APPROVED_MANIFEST_DIGEST": digest,
			"RELEASE_TAG":              "v0.1.0",
			"RELEASE_VERSION":          "0.1.0",
			"SOURCE_SHA":               sourceSHA,
		})
		return err
	}
	writeManifest := func(t *testing.T, raw []byte) string {
		t.Helper()
		path := filepath.Join(t.TempDir(), "manifest.json")
		if err := os.WriteFile(path, raw, 0o600); err != nil {
			t.Fatal(err)
		}
		return path
	}

	validBytes := marshalJSON(t, valid)
	if err := run(t, validBytes, sha256Hex(validBytes), writeManifest(t, validBytes)); err != nil {
		t.Fatalf("valid manifest rejected: %v", err)
	}

	t.Run("missing", func(t *testing.T) {
		if err := run(t, nil, strings.Repeat("0", 64), filepath.Join(t.TempDir(), "missing.json")); err == nil {
			t.Fatal("missing manifest accepted")
		}
	})
	t.Run("malformed", func(t *testing.T) {
		raw := []byte("{")
		if err := run(t, raw, sha256Hex(raw), writeManifest(t, raw)); err == nil {
			t.Fatal("malformed manifest accepted")
		}
	})
	t.Run("arbitrary-digest", func(t *testing.T) {
		if err := run(t, validBytes, strings.Repeat("0", 64), writeManifest(t, validBytes)); err == nil {
			t.Fatal("manifest with non-matching approved digest accepted")
		}
	})
	t.Run("source-drift", func(t *testing.T) {
		manifest := cloneJSONMap(t, valid)
		manifest["source_sha"] = strings.Repeat("b", 40)
		raw := marshalJSON(t, manifest)
		if err := run(t, raw, sha256Hex(raw), writeManifest(t, raw)); err == nil {
			t.Fatal("manifest source drift accepted")
		}
	})
	t.Run("tag-drift", func(t *testing.T) {
		manifest := cloneJSONMap(t, valid)
		manifest["tag"] = "v0.1.1"
		raw := marshalJSON(t, manifest)
		if err := run(t, raw, sha256Hex(raw), writeManifest(t, raw)); err == nil {
			t.Fatal("manifest tag drift accepted")
		}
	})
	t.Run("candidate-inventory-substitution", func(t *testing.T) {
		manifest := cloneJSONMap(t, valid)
		artifacts := manifest["artifacts"].([]any)
		artifacts[0].(map[string]any)["archive"] = "substituted.tar.gz"
		raw := marshalJSON(t, manifest)
		if err := run(t, raw, sha256Hex(raw), writeManifest(t, raw)); err == nil {
			t.Fatal("substituted candidate inventory accepted")
		}
	})
}

func TestRemoteReleaseStateValidatorFailsClosed(t *testing.T) {
	validator := extractPythonBlock(t, readReleaseWorkflow(t), "release-state-validator")
	sourceSHA := strings.Repeat("a", 40)
	run := func(t *testing.T, state map[string]any) error {
		t.Helper()
		dir := t.TempDir()
		statePath := filepath.Join(dir, "remote-state.json")
		if err := os.WriteFile(statePath, marshalJSON(t, state), 0o600); err != nil {
			t.Fatal(err)
		}
		_, err := runPythonBlock(t, validator, []string{statePath, filepath.Join(dir, "readback.json")}, map[string]string{
			"RELEASE_TAG": "v0.1.0",
			"SOURCE_SHA":  sourceSHA,
		})
		return err
	}

	for name, state := range map[string]map[string]any{
		"no-existing-tag-or-release": {
			"release_http_status": 404,
			"tag_exists":          false,
			"tag_source_sha":      nil,
		},
	} {
		t.Run(name, func(t *testing.T) {
			if err := run(t, state); err != nil {
				t.Fatalf("safe remote state rejected: %v", err)
			}
		})
	}
	for name, state := range map[string]map[string]any{
		"tag-source-drift": {
			"release_http_status": 404,
			"tag_exists":          true,
			"tag_source_sha":      strings.Repeat("b", 40),
		},
		"exact-existing-tag-without-release": {
			"release_http_status": 404,
			"tag_exists":          true,
			"tag_source_sha":      sourceSHA,
		},
		"existing-release": {
			"release_http_status": 200,
			"tag_exists":          true,
			"tag_source_sha":      sourceSHA,
		},
		"unknown-release-state": {
			"release_http_status": 500,
			"tag_exists":          false,
			"tag_source_sha":      nil,
		},
	} {
		t.Run(name, func(t *testing.T) {
			if err := run(t, state); err == nil {
				t.Fatalf("unsafe remote state %v accepted", state)
			}
		})
	}
}

func TestEnvironmentGateValidatorRequiresProtectedEnvironment(t *testing.T) {
	validator := extractPythonBlock(t, readReleaseWorkflow(t), "environment-gate-validator")
	run := func(t *testing.T, state map[string]any) (map[string]any, error) {
		t.Helper()
		dir := t.TempDir()
		statePath := filepath.Join(dir, "environment.json")
		readbackPath := filepath.Join(dir, "readback.json")
		if err := os.WriteFile(statePath, marshalJSON(t, state), 0o600); err != nil {
			t.Fatal(err)
		}
		_, err := runPythonBlock(t, validator, []string{statePath, readbackPath}, map[string]string{
			"RELEASE_ENVIRONMENT": "ao-mission-release",
		})
		var readback map[string]any
		if raw, readErr := os.ReadFile(readbackPath); readErr == nil {
			if jsonErr := json.Unmarshal(raw, &readback); jsonErr != nil {
				t.Fatal(jsonErr)
			}
		}
		return readback, err
	}

	protected := map[string]any{
		"name": "ao-mission-release",
		"protection_rules": []any{
			map[string]any{
				"type": "required_reviewers",
				"reviewers": []any{
					map[string]any{"type": "User"},
				},
			},
		},
	}
	readback, err := run(t, protected)
	if err != nil {
		t.Fatalf("protected environment rejected: %v", err)
	}
	if readback["status"] != "ready" || readback["protected"] != true {
		t.Fatalf("environment readback=%v, want ready and protected", readback)
	}

	unprotected := map[string]any{"name": "ao-mission-release", "protection_rules": []any{}}
	readback, err = run(t, unprotected)
	if err == nil {
		t.Fatal("unprotected environment accepted")
	}
	if readback["status"] != "blocked" || readback["protected"] != false {
		t.Fatalf("blocked environment readback=%v", readback)
	}
}

func TestDryRunBoundaryAllowsPrivateArtifactsButNoPublicMutation(t *testing.T) {
	writer := extractPythonBlock(t, readReleaseWorkflow(t), "dry-run-boundary-writer")
	path := filepath.Join(t.TempDir(), "dry-run-boundary.json")
	if _, err := runPythonBlock(t, writer, []string{path}, map[string]string{
		"DRY_RUN":     "true",
		"RELEASE_TAG": "v0.1.0",
		"SOURCE_SHA":  strings.Repeat("a", 40),
	}); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var boundary map[string]any
	if err := json.Unmarshal(raw, &boundary); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{
		"private_candidate_artifact_uploads_performed",
		"private_plan_artifact_upload_authorized",
	} {
		if boundary[field] != true {
			t.Fatalf("%s=%v, want true", field, boundary[field])
		}
	}
	for _, field := range []string{
		"deployment_attempted",
		"publication_performed",
		"public_release_asset_upload_attempted",
		"release_creation_attempted",
		"tag_creation_attempted",
	} {
		if boundary[field] != false {
			t.Fatalf("%s=%v, want false", field, boundary[field])
		}
	}
}

func TestNativeCandidatesEmitSeparateHelpVersionAndFunctionalSmokeEvidence(t *testing.T) {
	workflow := readReleaseWorkflow(t)
	for _, want := range []string{
		"help-evidence.json",
		"version-evidence.json",
		"version-output.txt",
		"functional-smoke-evidence.json",
		"validate contract --path examples/valid/mission-record.json",
		`if json.load(open(sys.argv[1], encoding="utf-8")).get("status") != "ready":`,
		"provider_calls",
		"-X github.com/uesugitorachiyo/ao-mission/internal/mission.BuildVersion=${RELEASE_VERSION}",
		"-X github.com/uesugitorachiyo/ao-mission/internal/mission.BuildSourceSHA=${SOURCE_SHA}",
		`version_output=$("$package_dir/$binary" --version)`,
		`version_output=${version_output%$'\r'}`,
		`[ "$version_output" = "$expected_version_output" ]`,
		`git cat-file blob "${SOURCE_SHA}:${VERSION_SOURCE_PATH}" > "$version_source_blob"`,
		`actual_version_source_sha256=$(sha256sum "$version_source_blob" | awk '{print $1}')`,
		"docs/sdd/AO-MISSION-V0.1.md",
		"DISPATCH_SHA: ${{ github.sha }}",
		`[ "$SOURCE_SHA" = "$DISPATCH_SHA" ]`,
		"environment-gate-readback.json",
		"release-preflight-readback.json",
		"base64.b64decode(encoded, validate=True)",
		"if len(encoded) > 49152:",
		"if not manifest_bytes or len(manifest_bytes) > 32768:",
	} {
		if !strings.Contains(workflow, want) {
			t.Fatalf("release workflow missing %q", want)
		}
	}
	if strings.Contains(workflow, `"smoke":{"command":"no-args-usage"`) {
		t.Fatal("workflow still classifies no-argument failure as functional smoke")
	}
	if strings.Contains(workflow, `grep -F '"status": "ready"'`) {
		t.Fatal("workflow must parse functional smoke JSON instead of matching its formatting")
	}
	if strings.Contains(workflow, `sha256sum "$VERSION_SOURCE_PATH"`) ||
		strings.Contains(workflow, `shasum -a 256 "$VERSION_SOURCE_PATH"`) {
		t.Fatal("workflow must hash exact committed version-source bytes, not a checkout-normalized working-tree file")
	}
}

func TestCandidateVersionCommandReportsLinkerBoundVersionAndSource(t *testing.T) {
	sourceSHA := strings.Repeat("a", 40)
	binaryName := "ao-mission"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	binaryPath := filepath.Join(t.TempDir(), binaryName)
	ldflags := strings.Join([]string{
		"-X github.com/uesugitorachiyo/ao-mission/internal/mission.BuildVersion=0.1.0",
		"-X github.com/uesugitorachiyo/ao-mission/internal/mission.BuildSourceSHA=" + sourceSHA,
	}, " ")
	build := exec.Command("go", "build", "-trimpath", "-ldflags", ldflags, "-o", binaryPath, "../../cmd/ao-mission")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build candidate: %v\n%s", err, output)
	}
	command := exec.Command(binaryPath, "--version")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("candidate --version failed: %v\n%s", err, output)
	}
	want := "ao-mission version=0.1.0 source_sha=" + sourceSHA + "\n"
	if string(output) != want {
		t.Fatalf("candidate --version output = %q, want %q", output, want)
	}
}

func TestVersionCommandDefaultsToDevelopmentIdentity(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"--version"}, &stdout, &stderr); code != 0 {
		t.Fatalf("development --version failed: code=%d stderr=%q", code, stderr.String())
	}
	if got, want := stdout.String(), "ao-mission version=dev source_sha=unknown\n"; got != want {
		t.Fatalf("development --version output = %q, want %q", got, want)
	}
}

func TestNativeMatrixBindsExactRunnerAndGoIdentity(t *testing.T) {
	workflow := readReleaseWorkflow(t)
	for _, want := range []string{
		"os: ubuntu-24.04",
		"os: macos-15",
		"os: windows-2025",
		"expected_runner_os: Linux",
		"expected_runner_os: macOS",
		"expected_runner_os: Windows",
		"expected_runner_arch: X64",
		"expected_runner_arch: ARM64",
		"expected_goos: linux",
		"expected_goos: darwin",
		"expected_goos: windows",
		"expected_goarch: amd64",
		"expected_goarch: arm64",
		`[ "$RUNNER_OS" = "$EXPECTED_RUNNER_OS" ]`,
		`[ "$RUNNER_ARCH" = "$EXPECTED_RUNNER_ARCH" ]`,
		`[ "$(go env GOOS)" = "$EXPECTED_GOOS" ]`,
		`[ "$(go env GOARCH)" = "$EXPECTED_GOARCH" ]`,
	} {
		if !strings.Contains(workflow, want) {
			t.Fatalf("release workflow missing exact native identity binding %q", want)
		}
	}
	for _, forbidden := range []string{
		"os: macos-latest",
		"os: ubuntu-latest",
		"os: windows-latest",
	} {
		if strings.Contains(workflow, forbidden) {
			t.Fatalf("release workflow uses mutable runner architecture label %q", forbidden)
		}
	}
}

func TestStrictReleaseVerifierRejectsWrongVersionAndTargetSubstitution(t *testing.T) {
	verifier := extractPythonBlock(t, readReleaseWorkflow(t), "strict-release-verifier")

	t.Run("valid-candidates", func(t *testing.T) {
		fixture := writeReleaseVerifierFixture(t, verifier, nil)
		if output, err := runStrictVerifier(t, verifier, fixture, "candidates"); err != nil {
			t.Fatalf("valid candidates rejected: %v\n%s", err, output)
		}
	})
	t.Run("wrong-version", func(t *testing.T) {
		fixture := writeReleaseVerifierFixture(t, verifier, func(target releaseTargetFixture, files map[string][]byte, summary, provenance map[string]any) {
			if target.targetLabel == "linux-x86_64" {
				summary["version"] = "0.1.1"
				provenance["version"] = "0.1.1"
				files["provenance.json"] = marshalJSON(t, provenance)
			}
		})
		if _, err := runStrictVerifier(t, verifier, fixture, "candidates"); err == nil {
			t.Fatal("coherent wrong-version candidate accepted")
		}
	})
	t.Run("self-asserted-version-output", func(t *testing.T) {
		fixture := writeReleaseVerifierFixture(t, verifier, func(target releaseTargetFixture, files map[string][]byte, summary, _ map[string]any) {
			if target.targetLabel == "linux-x86_64" {
				output := "ao-mission version=0.1.1 source_sha=" + strings.Repeat("b", 40)
				evidence := summary["version_evidence"].(map[string]any)
				evidence["release_version"] = "0.1.1"
				evidence["source_sha"] = strings.Repeat("b", 40)
				evidence["output"] = output
				files["version-output.txt"] = []byte(output + "\n")
				files["version-evidence.json"] = marshalJSON(t, evidence)
			}
		})
		if _, err := runStrictVerifier(t, verifier, fixture, "candidates"); err == nil {
			t.Fatal("self-asserted substituted candidate --version output accepted")
		}
	})
	t.Run("binary-format-substitution", func(t *testing.T) {
		fixture := writeReleaseVerifierFixture(t, verifier, func(target releaseTargetFixture, files map[string][]byte, summary, provenance map[string]any) {
			if target.targetLabel == "linux-x86_64" {
				files[target.entryPoint] = fixtureBinary("pe-x86_64")
				summary["binary_sha256"] = sha256Hex(files[target.entryPoint])
				provenance["binary_sha256"] = summary["binary_sha256"]
				files["provenance.json"] = marshalJSON(t, provenance)
			}
		})
		if _, err := runStrictVerifier(t, verifier, fixture, "candidates"); err == nil {
			t.Fatal("PE binary substituted into Linux candidate accepted")
		}
	})
	t.Run("archive-traversal", func(t *testing.T) {
		fixture := writeReleaseVerifierFixture(t, verifier, func(target releaseTargetFixture, files map[string][]byte, _, _ map[string]any) {
			if target.targetLabel == "linux-x86_64" {
				files["../escape"] = []byte("escape")
			}
		})
		if _, err := runStrictVerifier(t, verifier, fixture, "candidates"); err == nil {
			t.Fatal("archive traversal member accepted")
		}
	})
}

func TestStrictReleaseVerifierRejectsMalformedAndSemanticPlanMutation(t *testing.T) {
	verifier := extractPythonBlock(t, readReleaseWorkflow(t), "strict-release-verifier")

	t.Run("valid-plan", func(t *testing.T) {
		fixture := writeReleaseVerifierFixture(t, verifier, nil)
		if output, err := runStrictVerifier(t, verifier, fixture, "plan"); err != nil {
			t.Fatalf("valid plan rejected: %v\n%s", err, output)
		}
	})
	t.Run("malformed-plan-with-recomputed-sidecar", func(t *testing.T) {
		fixture := writeReleaseVerifierFixture(t, verifier, nil)
		raw := []byte("{")
		if err := os.WriteFile(fixture.planPath, raw, 0o600); err != nil {
			t.Fatal(err)
		}
		writePlanChecksum(t, fixture.planChecksum, raw)
		if _, err := runStrictVerifier(t, verifier, fixture, "plan"); err == nil {
			t.Fatal("malformed plan accepted after recomputing sidecar")
		}
	})
	t.Run("semantic-plan-mutation-with-recomputed-sidecar", func(t *testing.T) {
		fixture := writeReleaseVerifierFixture(t, verifier, nil)
		var plan map[string]any
		raw, err := os.ReadFile(fixture.planPath)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(raw, &plan); err != nil {
			t.Fatal(err)
		}
		plan["repository"] = "substituted-repository"
		raw = marshalJSON(t, plan)
		if err := os.WriteFile(fixture.planPath, raw, 0o600); err != nil {
			t.Fatal(err)
		}
		writePlanChecksum(t, fixture.planChecksum, raw)
		if _, err := runStrictVerifier(t, verifier, fixture, "plan"); err == nil {
			t.Fatal("semantically mutated plan accepted after recomputing sidecar")
		}
	})
}

func readReleaseWorkflow(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", "release-rehearsal.yml"))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func parseWorkflowShape(t *testing.T, workflow string) workflowShape {
	t.Helper()
	shape := workflowShape{
		events:         map[string]bool{},
		inputs:         map[string]map[string]string{},
		topPermissions: map[string]string{},
		jobs:           map[string]*workflowJob{},
	}
	section := ""
	event := ""
	input := ""
	jobName := ""
	jobSection := ""
	for lineNumber, line := range strings.Split(workflow, "\n") {
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " "))
		trimmed := strings.TrimSpace(line)
		if indent == 0 {
			if strings.HasSuffix(trimmed, ":") {
				section = strings.TrimSuffix(trimmed, ":")
				event, input, jobName, jobSection = "", "", "", ""
			} else {
				section = ""
			}
			continue
		}
		switch section {
		case "on":
			if indent == 2 && strings.HasSuffix(trimmed, ":") {
				event = strings.TrimSuffix(trimmed, ":")
				shape.events[event] = true
				continue
			}
			if event == "workflow_dispatch" && indent == 6 && strings.HasSuffix(trimmed, ":") {
				input = strings.TrimSuffix(trimmed, ":")
				shape.inputs[input] = map[string]string{}
				continue
			}
			if input != "" && indent == 8 {
				key, value, ok := yamlField(trimmed)
				if ok {
					shape.inputs[input][key] = value
				}
			}
		case "permissions":
			if indent == 2 {
				key, value, ok := yamlField(trimmed)
				if ok {
					shape.topPermissions[key] = value
				}
			}
		case "jobs":
			if indent == 2 && strings.HasSuffix(trimmed, ":") {
				jobName = strings.TrimSuffix(trimmed, ":")
				shape.jobs[jobName] = &workflowJob{permissions: map[string]string{}}
				jobSection = ""
				continue
			}
			if jobName == "" {
				continue
			}
			job := shape.jobs[jobName]
			if indent == 4 && strings.HasSuffix(trimmed, ":") {
				jobSection = strings.TrimSuffix(trimmed, ":")
				continue
			}
			if indent == 4 {
				key, value, ok := yamlField(trimmed)
				if !ok {
					continue
				}
				switch key {
				case "if":
					job.condition = value
				case "environment":
					job.environment = value
				}
				jobSection = ""
				continue
			}
			if jobSection == "permissions" && indent == 6 {
				key, value, ok := yamlField(trimmed)
				if ok {
					job.permissions[key] = value
				}
			}
			if jobSection == "steps" && indent == 8 {
				key, value, ok := yamlField(trimmed)
				if ok && key == "uses" {
					job.uses = append(job.uses, value)
				}
			}
		default:
			t.Fatalf("line %d is under unknown top-level section %q", lineNumber+1, section)
		}
	}
	return shape
}

func yamlField(line string) (string, string, bool) {
	key, value, ok := strings.Cut(line, ":")
	if !ok || strings.TrimSpace(value) == "" {
		return "", "", false
	}
	value = strings.TrimSpace(value)
	if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'')) {
		value = value[1 : len(value)-1]
	}
	return strings.TrimSpace(key), value, true
}

func extractPythonBlock(t *testing.T, workflow, name string) string {
	t.Helper()
	begin := "# " + name + "-begin"
	end := "# " + name + "-end"
	lines := strings.Split(workflow, "\n")
	start, finish := -1, -1
	for i, line := range lines {
		switch strings.TrimSpace(line) {
		case begin:
			start = i + 1
		case end:
			finish = i
		}
	}
	if start < 0 || finish < start {
		t.Fatalf("workflow missing embedded Python block %q", name)
	}
	block := append([]string(nil), lines[start:finish]...)
	minIndent := -1
	for _, line := range block {
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " "))
		if minIndent < 0 || indent < minIndent {
			minIndent = indent
		}
	}
	for i := range block {
		if len(block[i]) >= minIndent {
			block[i] = block[i][minIndent:]
		}
	}
	return strings.Join(block, "\n") + "\n"
}

func runPythonBlock(t *testing.T, script string, args []string, environment map[string]string) (string, error) {
	t.Helper()
	scriptPath := filepath.Join(t.TempDir(), "validator.py")
	if err := os.WriteFile(scriptPath, []byte(script), 0o600); err != nil {
		t.Fatal(err)
	}
	command := exec.Command("python3", append([]string{scriptPath}, args...)...)
	command.Env = os.Environ()
	keys := make([]string, 0, len(environment))
	for key := range environment {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		command.Env = append(command.Env, key+"="+environment[key])
	}
	var output bytes.Buffer
	command.Stdout = &output
	command.Stderr = &output
	err := command.Run()
	return output.String(), err
}

func runStrictVerifier(t *testing.T, verifier string, fixture releaseVerifierFixture, mode string) (string, error) {
	t.Helper()
	args := []string{
		mode,
		"--manifest", fixture.manifestPath,
		"--candidates", fixture.candidatesPath,
	}
	if mode == "candidates" {
		args = append(args, "--output", filepath.Join(t.TempDir(), "verified-candidates.json"))
	} else {
		args = append(args,
			"--plan", fixture.planPath,
			"--plan-checksum", fixture.planChecksum,
		)
	}
	return runPythonBlock(t, verifier, args, fixture.environment)
}

func writeReleaseVerifierFixture(
	t *testing.T,
	verifier string,
	mutate func(releaseTargetFixture, map[string][]byte, map[string]any, map[string]any),
) releaseVerifierFixture {
	t.Helper()
	root := t.TempDir()
	candidatesPath := filepath.Join(root, "candidates")
	if err := os.MkdirAll(candidatesPath, 0o755); err != nil {
		t.Fatal(err)
	}
	sourceSHA := strings.Repeat("a", 40)
	versionSourceDigest := strings.Repeat("b", 64)
	manifest := releaseManifestFixture(sourceSHA, versionSourceDigest)
	manifestBytes := marshalJSON(t, manifest)
	manifestPath := filepath.Join(root, "approved-release-manifest.json")
	if err := os.WriteFile(manifestPath, manifestBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	environment := map[string]string{
		"APPROVED_MANIFEST_DIGEST": sha256Hex(manifestBytes),
		"RELEASE_NOTES_SHA256":     sha256Hex([]byte("approved notes")),
		"RELEASE_TAG":              "v0.1.0",
		"RELEASE_VERSION":          "0.1.0",
		"SOURCE_SHA":               sourceSHA,
		"VERIFIER_SHA256":          sha256Hex([]byte(verifier)),
		"VERSION_SOURCE_PATH":      "docs/sdd/AO-MISSION-V0.1.md",
		"VERSION_SOURCE_SHA256":    versionSourceDigest,
		"WORKFLOW_REF":             "uesugitorachiyo/ao-mission/.github/workflows/release-rehearsal.yml@refs/heads/main",
	}
	for _, target := range releaseTargetFixtures("0.1.0") {
		writeCandidateFixture(t, candidatesPath, target, manifestBytes, environment, mutate)
	}

	plan := map[string]any{
		"approved_manifest_digest":    environment["APPROVED_MANIFEST_DIGEST"],
		"approved_manifest_inventory": manifest["artifacts"],
		"candidates":                  readCandidateSummaries(t, candidatesPath),
		"immutable":                   true,
		"release_notes_sha256":        environment["RELEASE_NOTES_SHA256"],
		"repository":                  "ao-mission",
		"schema_version":              "ao.release-rehearsal-promotion-plan.v0.3",
		"source_sha":                  sourceSHA,
		"tag":                         "v0.1.0",
		"verifier_sha256":             environment["VERIFIER_SHA256"],
		"version":                     "0.1.0",
	}
	planBytes := marshalJSON(t, plan)
	planPath := filepath.Join(root, "immutable-promotion-plan.json")
	if err := os.WriteFile(planPath, planBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	planChecksum := filepath.Join(root, "immutable-promotion-plan.sha256")
	writePlanChecksum(t, planChecksum, planBytes)
	return releaseVerifierFixture{
		candidatesPath: candidatesPath,
		environment:    environment,
		manifestPath:   manifestPath,
		planChecksum:   planChecksum,
		planPath:       planPath,
	}
}

func writeCandidateFixture(
	t *testing.T,
	root string,
	target releaseTargetFixture,
	manifestBytes []byte,
	environment map[string]string,
	mutate func(releaseTargetFixture, map[string][]byte, map[string]any, map[string]any),
) {
	t.Helper()
	binaryBytes := fixtureBinary(target.binaryFormat + "-" + target.machine)
	versionOutput := "ao-mission version=" + environment["RELEASE_VERSION"] + " source_sha=" + environment["SOURCE_SHA"]
	versionEvidence := map[string]any{
		"command":               "--version",
		"output":                versionOutput,
		"release_version":       environment["RELEASE_VERSION"],
		"source_sha":            environment["SOURCE_SHA"],
		"status":                "passed",
		"version_source":        environment["VERSION_SOURCE_PATH"],
		"version_source_sha256": environment["VERSION_SOURCE_SHA256"],
	}
	files := map[string][]byte{
		target.entryPoint:                binaryBytes,
		"LICENSE":                        []byte("license\n"),
		"NOTICE":                         []byte("notice\n"),
		"approved-release-manifest.json": manifestBytes,
		"help.txt":                       []byte("usage: ao-mission <command>\n"),
		"help-evidence.json":             marshalJSON(t, map[string]any{"command": "no-args-usage", "expected_exit": "nonzero", "status": "passed"}),
		"version-evidence.json":          marshalJSON(t, versionEvidence),
		"version-output.txt":             []byte(versionOutput + "\n"),
		"functional-smoke-evidence.json": marshalJSON(t, map[string]any{"command": "validate contract --path examples/valid/mission-record.json", "provider_calls": false, "status": "passed"}),
		"functional-smoke-output.json":   marshalJSON(t, validFunctionalSmokeOutput()),
		"sbom.json":                      marshalJSON(t, map[string]any{"GoVersion": "1.22", "Path": "github.com/uesugitorachiyo/ao-mission"}),
	}
	provenance := map[string]any{
		"approved_manifest_digest": environment["APPROVED_MANIFEST_DIGEST"],
		"archive":                  target.archive,
		"binary_format":            target.binaryFormat,
		"binary_sha256":            sha256Hex(binaryBytes),
		"go_version":               "go version go1.22.12 " + target.goos + "/" + target.goarch,
		"goarch":                   target.goarch,
		"goos":                     target.goos,
		"machine":                  target.machine,
		"release_notes_sha256":     environment["RELEASE_NOTES_SHA256"],
		"repository":               "ao-mission",
		"runner_arch":              target.runnerArch,
		"runner_label":             target.runnerLabel,
		"runner_os":                target.runnerOS,
		"schema_version":           "ao.release-rehearsal-provenance.v0.3",
		"source_sha":               environment["SOURCE_SHA"],
		"target_label":             target.targetLabel,
		"version":                  environment["RELEASE_VERSION"],
		"workflow_ref":             environment["WORKFLOW_REF"],
		"workflow_sha":             environment["SOURCE_SHA"],
	}
	summary := map[string]any{
		"approved_manifest_digest": environment["APPROVED_MANIFEST_DIGEST"],
		"archive":                  target.archive,
		"binary_format":            target.binaryFormat,
		"binary_sha256":            sha256Hex(binaryBytes),
		"checksum_file":            "SHA256SUMS",
		"functional_smoke":         map[string]any{"command": "validate contract --path examples/valid/mission-record.json", "provider_calls": false, "status": "passed"},
		"goarch":                   target.goarch,
		"goos":                     target.goos,
		"help":                     map[string]any{"command": "no-args-usage", "expected_exit": "nonzero", "status": "passed"},
		"machine":                  target.machine,
		"release_notes_sha256":     environment["RELEASE_NOTES_SHA256"],
		"repository":               "ao-mission",
		"runner_arch":              target.runnerArch,
		"runner_label":             target.runnerLabel,
		"runner_os":                target.runnerOS,
		"schema_version":           "ao.release-rehearsal-candidate.v0.3",
		"source_sha":               environment["SOURCE_SHA"],
		"target_label":             target.targetLabel,
		"version":                  environment["RELEASE_VERSION"],
		"version_evidence":         versionEvidence,
	}
	if mutate != nil {
		mutate(target, files, summary, provenance)
	}
	binaryBytes = files[target.entryPoint]
	binaryDigest := sha256Hex(binaryBytes)
	summary["binary_sha256"] = binaryDigest
	provenance["binary_sha256"] = binaryDigest
	provenanceBytes := marshalJSON(t, provenance)
	files["provenance.json"] = provenanceBytes

	candidateDir := filepath.Join(root, target.targetLabel)
	if err := os.MkdirAll(candidateDir, 0o755); err != nil {
		t.Fatal(err)
	}
	archivePath := filepath.Join(candidateDir, target.archive)
	writeCandidateArchive(t, archivePath, target.archive, files)
	archiveBytes, err := os.ReadFile(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	archiveDigest := sha256Hex(archiveBytes)
	summary["archive_sha256"] = archiveDigest
	summary["provenance_sha256"] = sha256Hex(provenanceBytes)
	if err := os.WriteFile(filepath.Join(candidateDir, "SHA256SUMS"), []byte(fmt.Sprintf("%s  %s\n", archiveDigest, target.archive)), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(candidateDir, "candidate-summary.json"), marshalJSON(t, summary), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(candidateDir, "provenance.json"), provenanceBytes, 0o600); err != nil {
		t.Fatal(err)
	}
}

func writeCandidateArchive(t *testing.T, path, archiveName string, files map[string][]byte) {
	t.Helper()
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if strings.HasSuffix(archiveName, ".zip") {
		writer := zip.NewWriter(file)
		for _, name := range names {
			header := &zip.FileHeader{Name: name, Method: zip.Deflate}
			header.SetMode(0o644)
			entry, createErr := writer.CreateHeader(header)
			if createErr != nil {
				t.Fatal(createErr)
			}
			if _, writeErr := entry.Write(files[name]); writeErr != nil {
				t.Fatal(writeErr)
			}
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
		return
	}
	gzipWriter := gzip.NewWriter(file)
	tarWriter := tar.NewWriter(gzipWriter)
	for _, name := range names {
		header := &tar.Header{Name: name, Mode: 0o644, Size: int64(len(files[name])), Typeflag: tar.TypeReg}
		if err := tarWriter.WriteHeader(header); err != nil {
			t.Fatal(err)
		}
		if _, err := tarWriter.Write(files[name]); err != nil {
			t.Fatal(err)
		}
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatal(err)
	}
}

func fixtureBinary(identity string) []byte {
	switch identity {
	case "elf-x86_64":
		raw := make([]byte, 64)
		copy(raw, []byte{0x7f, 'E', 'L', 'F'})
		raw[4], raw[5] = 2, 1
		binary.LittleEndian.PutUint16(raw[18:20], 0x3e)
		return raw
	case "macho-arm64":
		raw := make([]byte, 64)
		binary.LittleEndian.PutUint32(raw[0:4], 0xfeedfacf)
		binary.LittleEndian.PutUint32(raw[4:8], 0x0100000c)
		return raw
	case "pe-x86_64":
		raw := make([]byte, 256)
		copy(raw, []byte{'M', 'Z'})
		binary.LittleEndian.PutUint32(raw[0x3c:0x40], 0x80)
		copy(raw[0x80:0x84], []byte{'P', 'E', 0, 0})
		binary.LittleEndian.PutUint16(raw[0x84:0x86], 0x8664)
		return raw
	default:
		panic("unknown fixture binary identity: " + identity)
	}
}

func validFunctionalSmokeOutput() map[string]any {
	return map[string]any{
		"approves_work":        false,
		"blockers":             []any{},
		"contract":             "ao.mission.record.v0.1",
		"executes_work":        false,
		"generated_at_utc":     "2026-07-20T00:00:00Z",
		"mutates_repositories": false,
		"path":                 "examples/valid/mission-record.json",
		"read_only":            true,
		"schema":               "ao.mission.contract-validation.v0.1",
		"status":               "ready",
	}
}

func readCandidateSummaries(t *testing.T, root string) []any {
	t.Helper()
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	var summaries []any
	for _, entry := range entries {
		raw, readErr := os.ReadFile(filepath.Join(root, entry.Name(), "candidate-summary.json"))
		if readErr != nil {
			t.Fatal(readErr)
		}
		var summary map[string]any
		if err := json.Unmarshal(raw, &summary); err != nil {
			t.Fatal(err)
		}
		summaries = append(summaries, summary)
	}
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].(map[string]any)["target_label"].(string) < summaries[j].(map[string]any)["target_label"].(string)
	})
	return summaries
}

func writePlanChecksum(t *testing.T, path string, raw []byte) {
	t.Helper()
	content := fmt.Sprintf("%s  immutable-promotion-plan.json\n", sha256Hex(raw))
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func releaseTargetFixtures(version string) []releaseTargetFixture {
	return []releaseTargetFixture{
		{
			archive: "ao-mission-" + version + "-linux-x86_64.tar.gz", architecture: "x86_64",
			binaryFormat: "elf", entryPoint: "ao-mission", goarch: "amd64", goos: "linux",
			machine: "x86_64", os: "linux", runnerArch: "X64", runnerLabel: "ubuntu-24.04",
			runnerOS: "Linux", targetLabel: "linux-x86_64",
		},
		{
			archive: "ao-mission-" + version + "-macos-aarch64.tar.gz", architecture: "aarch64",
			binaryFormat: "macho", entryPoint: "ao-mission", goarch: "arm64", goos: "darwin",
			machine: "arm64", os: "macos", runnerArch: "ARM64", runnerLabel: "macos-15",
			runnerOS: "macOS", targetLabel: "macos-aarch64",
		},
		{
			archive: "ao-mission-" + version + "-windows-x86_64.zip", architecture: "x86_64",
			binaryFormat: "pe", entryPoint: "ao-mission.exe", goarch: "amd64", goos: "windows",
			machine: "x86_64", os: "windows", runnerArch: "X64", runnerLabel: "windows-2025",
			runnerOS: "Windows", targetLabel: "windows-x86_64",
		},
	}
}

func releaseManifestFixture(sourceSHA, versionSourceDigest string) map[string]any {
	artifacts := make([]any, 0, 3)
	for _, target := range releaseTargetFixtures("0.1.0") {
		artifacts = append(artifacts, map[string]any{
			"archive":       target.archive,
			"architecture":  target.architecture,
			"binary_format": target.binaryFormat,
			"entry_point":   target.entryPoint,
			"goarch":        target.goarch,
			"goos":          target.goos,
			"machine":       target.machine,
			"os":            target.os,
			"runner_arch":   target.runnerArch,
			"runner_label":  target.runnerLabel,
			"runner_os":     target.runnerOS,
			"target_label":  target.targetLabel,
		})
	}
	return map[string]any{
		"schema_version":        "ao.release-rehearsal-manifest.v0.1",
		"repository":            "ao-mission",
		"version":               "0.1.0",
		"tag":                   "v0.1.0",
		"source_sha":            sourceSHA,
		"version_source":        "docs/sdd/AO-MISSION-V0.1.md",
		"version_source_sha256": versionSourceDigest,
		"artifacts":             artifacts,
	}
}

func marshalJSON(t *testing.T, value any) []byte {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func cloneJSONMap(t *testing.T, value map[string]any) map[string]any {
	t.Helper()
	raw := marshalJSON(t, value)
	var clone map[string]any
	if err := json.Unmarshal(raw, &clone); err != nil {
		t.Fatal(err)
	}
	return clone
}

func sha256Hex(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
