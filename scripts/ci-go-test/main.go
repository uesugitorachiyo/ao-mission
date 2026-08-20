package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

const maxAnnotations = 20

var (
	packageFailure = regexp.MustCompile(`^FAIL\s+\S+(?:\s+\d+(?:\.\d+)?s)?$`)
	windowsPath    = regexp.MustCompile(`[A-Za-z]:[\\/][^\s]+`)
	posixPath      = regexp.MustCompile(`/(?:home|Users|runner|workspace|tmp)/[^\s]+`)
)

func main() {
	args := goTestCommand()
	command := exec.Command(args[0], args[1:]...)
	var stdout, stderr bytes.Buffer
	command.Stdout = io.MultiWriter(os.Stdout, &stdout)
	command.Stderr = io.MultiWriter(os.Stderr, &stderr)
	err := command.Run()
	if err == nil {
		return
	}

	summary := summarizeGoTestOutput(stdout.String()+"\n"+stderr.String(), maxAnnotations)
	if len(summary) == 0 {
		summary = []string{"go test failed without a recognized bounded summary line"}
	}
	for _, line := range summary {
		fmt.Printf("::error title=go test failure::%s\n", sanitizeAnnotation(line))
	}

	exitCode := 1
	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		exitCode = exitError.ExitCode()
	}
	os.Exit(exitCode)
}

func goTestCommand() []string {
	return []string{"go", "test", "./...", "-count=1"}
}

func summarizeGoTestOutput(output string, limit int) []string {
	if limit <= 0 {
		return nil
	}
	summary := make([]string, 0, limit)
	seen := make(map[string]struct{})
	for _, raw := range strings.Split(output, "\n") {
		line := strings.TrimSpace(strings.TrimSuffix(raw, "\r"))
		selected := strings.HasPrefix(line, "--- FAIL: Test") ||
			strings.HasPrefix(line, "panic: test timed out after ") ||
			packageFailure.MatchString(line)
		if !selected {
			continue
		}
		if _, duplicate := seen[line]; duplicate {
			continue
		}
		seen[line] = struct{}{}
		summary = append(summary, line)
		if len(summary) == limit {
			break
		}
	}
	return summary
}

func sanitizeAnnotation(value string) string {
	value = windowsPath.ReplaceAllString(value, "[path]")
	value = posixPath.ReplaceAllString(value, "[path]")
	return strings.NewReplacer(
		"%", "%25",
		"\r", "%0D",
		"\n", "%0A",
		":", "%3A",
		",", "%2C",
	).Replace(value)
}
