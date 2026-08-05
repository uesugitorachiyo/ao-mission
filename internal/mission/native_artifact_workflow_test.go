package mission

import (
	"os"
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
