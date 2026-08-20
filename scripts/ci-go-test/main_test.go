package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestGoTestCommandIsExact(t *testing.T) {
	want := []string{"go", "test", "./...", "-count=1"}
	if got := goTestCommand(); !reflect.DeepEqual(got, want) {
		t.Fatalf("go test command = %q, want %q", got, want)
	}
}

func TestSummarizeGoTestOutputSelectsOnlyBoundedSafeFailures(t *testing.T) {
	input := strings.Join([]string{
		"ok example.invalid/ok 0.01s",
		"--- FAIL: TestAlpha (0.01s)",
		"    alpha_test.go:12: token=super-secret C:\\Users\\runner\\work\\repo\\alpha_test.go",
		"FAIL example.invalid/internal/mission 1.23s",
		"panic: test timed out after 10m0s",
		"goroutine 23 [running]: /home/runner/work/repo/private.go:99",
		"--- FAIL: TestBeta/subcase (0.02s)",
	}, "\n")

	want := []string{
		"--- FAIL: TestAlpha (0.01s)",
		"FAIL example.invalid/internal/mission 1.23s",
		"panic: test timed out after 10m0s",
	}
	if got := summarizeGoTestOutput(input, 3); !reflect.DeepEqual(got, want) {
		t.Fatalf("summary = %#v, want %#v", got, want)
	}
}

func TestSummarizeGoTestOutputDeduplicatesAndIgnoresBareFail(t *testing.T) {
	input := "--- FAIL: TestAlpha (0.01s)\n--- FAIL: TestAlpha (0.01s)\nFAIL\n"
	want := []string{"--- FAIL: TestAlpha (0.01s)"}
	if got := summarizeGoTestOutput(input, 20); !reflect.DeepEqual(got, want) {
		t.Fatalf("summary = %#v, want %#v", got, want)
	}
}

func TestSanitizeAnnotationEscapesMetacharactersAndRedactsPaths(t *testing.T) {
	input := "failure: 100%, C:\\Users\\runner\\work\\repo\\file.go and /home/runner/work/repo/file.go\r\nnext"
	got := sanitizeAnnotation(input)
	for _, forbidden := range []string{"\r", "\n", "C:\\Users", "/home/runner", "100%,", "failure:"} {
		if strings.Contains(got, forbidden) {
			t.Fatalf("sanitized annotation %q contains %q", got, forbidden)
		}
	}
	for _, want := range []string{"failure%3A", "100%25%2C", "[path]", "next"} {
		if !strings.Contains(got, want) {
			t.Fatalf("sanitized annotation %q missing %q", got, want)
		}
	}
}
