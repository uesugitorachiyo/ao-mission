package mission

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProductizationAdoptionMonth1To6Closure(t *testing.T) {
	path := filepath.Join("..", "..", "docs", "roadmap", "productization-adoption-month1-6-closure.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read closure record: %v", err)
	}
	document := strings.Join(strings.Fields(string(content)), " ")
	required := []string{
		"PRODUCTIZATION_ADOPTION_COMPLETE_RELEASED_AND_VERIFIED",
		"AO2 v0.5.3",
		"https://github.com/uesugitorachiyo/ao2/releases/tag/v0.5.3",
		"`947e566bd3f54ed902f3c14fc0c90e21a24359bc`",
		"AO2 Control Plane v0.1.18",
		"https://github.com/uesugitorachiyo/ao2-control-plane/releases/tag/v0.1.18",
		"`6257ec23fde726d4a0133c5b62231881fb6aaa9a`",
		"AO Mission v0.1.0",
		"https://github.com/uesugitorachiyo/ao-mission/releases/tag/v0.1.0",
		"`2901a9cb887b72296a56b70a5a3be7350b28fe65`",
		"AO Command v0.1.1",
		"https://github.com/uesugitorachiyo/ao-command/releases/tag/v0.1.1",
		"`0bcadf5701fdac88f9fd792cba3a9a6686de16e5`",
		"AO Architecture PR #147",
		"`31b0e9f90a0385cf6f44efbae90a8be71a22b352`",
		"AO Blueprint, AO Atlas, AO Forge, and AO Covenant: `no_release_needed`",
		"Tier 3 remains artifact-only",
		"AO Architecture remains binary-free",
		"Unresolved high-severity findings: `0`",
		"Compatibility gate remains ready, not active",
		"No inbound Windows HTTP",
		"No self-hosted public-repository runner",
		"No credential changes",
		"No provider pilot",
		"No external beta launch",
		"No promotion or RSI authority",
	}
	for _, phrase := range required {
		if !strings.Contains(document, phrase) {
			t.Errorf("closure record missing %q", phrase)
		}
	}
}
