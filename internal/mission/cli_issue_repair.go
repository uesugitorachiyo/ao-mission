package mission

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
)

const issueRepairRequestLimit = 64 * 1024

func registerIssueRepairCLICommands(registry cliCommandRegistry) {
	registry["issue-repair"] = runIssueRepairCLICommand
}

func runIssueRepairCLICommand(s Store, args []string, stdout io.Writer) error {
	if len(args) < 2 || args[1] != "supervise" {
		return errors.New("issue-repair requires supervise")
	}
	flags := flag.NewFlagSet("issue-repair supervise", flag.ContinueOnError)
	missionID := flags.String("mission", "", "")
	requestPath := flags.String("request", "", "")
	jsonOut := flags.Bool("json", false, "")
	if err := flags.Parse(args[2:]); err != nil {
		return err
	}
	if flags.NArg() != 0 || strings.TrimSpace(*missionID) == "" ||
		strings.TrimSpace(*requestPath) == "" {
		return errors.New("issue-repair supervise requires --mission and --request")
	}
	body, err := readBoundedRegularFile(*requestPath, issueRepairRequestLimit)
	if err != nil {
		return err
	}
	var request IssueRepairSupervisorRequest
	if err := decodeStrictJSONObject(body, &request, "issue repair supervisor request", map[string]string{
		"run_id": "string", "run_envelope_digest": "string", "actor": "string",
		"lease_id": "string", "lease_owner": "string", "lease_expires_at": "string",
		"event_type": "string", "input_digests": "array", "output_digests": "array",
		"reason_code": "string", "expected_checkpoint_digest": "string",
		"event_budget": "integer",
	}, []string{
		"run_id", "run_envelope_digest", "actor", "lease_id", "lease_owner",
		"lease_expires_at", "event_type", "input_digests", "output_digests",
		"reason_code", "event_budget",
	}); err != nil {
		return err
	}
	state, err := SuperviseIssueRepair(s, *missionID, request)
	if err != nil {
		return err
	}
	if *jsonOut {
		return printJSON(stdout, state)
	}
	encoded, err := json.Marshal(state)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "issue_repair_supervisor=%s\n", encoded)
	return nil
}
