package mission

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type ContractValidation struct {
	Schema    string   `json:"schema"`
	Path      string   `json:"path"`
	Status    string   `json:"status"`
	Contract  string   `json:"contract"`
	Blockers  []string `json:"blockers"`
	ReadOnly  bool     `json:"read_only"`
	Executes  bool     `json:"executes_work"`
	Approves  bool     `json:"approves_work"`
	Mutates   bool     `json:"mutates_repositories"`
	Generated string   `json:"generated_at_utc"`
}

func ValidateContractFile(path string) (ContractValidation, error) {
	result := ContractValidation{
		Schema:    "ao.mission.contract-validation.v0.1",
		Path:      path,
		Status:    "ready",
		Blockers:  []string{},
		ReadOnly:  true,
		Generated: now(nil),
	}
	body, err := os.ReadFile(path)
	if err != nil {
		result.Status = "blocked"
		result.Blockers = append(result.Blockers, err.Error())
		return result, err
	}
	if err := ValidatePublicSafeText(string(body)); err != nil {
		result.Status = "blocked"
		result.Blockers = append(result.Blockers, err.Error())
		return result, err
	}
	var doc map[string]any
	if err := json.Unmarshal(body, &doc); err != nil {
		result.Status = "blocked"
		result.Blockers = append(result.Blockers, "invalid JSON")
		return result, err
	}
	schema, _ := doc["schema"].(string)
	if schema == "" {
		schema, _ = doc["contract_version"].(string)
	}
	if schema == "" {
		result.Status = "blocked"
		result.Blockers = append(result.Blockers, "schema or contract_version is required")
		return result, fmt.Errorf("schema or contract_version is required")
	}
	result.Contract = schema
	for _, field := range requiredFieldsForContract(schema) {
		if _, ok := doc[field]; !ok {
			result.Status = "blocked"
			result.Blockers = append(result.Blockers, field+" is required")
		}
	}
	if result.Status != "ready" {
		return result, fmt.Errorf(strings.Join(result.Blockers, "; "))
	}
	return result, nil
}

func requiredFieldsForContract(schema string) []string {
	switch schema {
	case RecordSchema:
		return []string{"schema", "mission_id", "objective_digest", "status", "created_at_utc", "current_route"}
	case SnapshotSchema:
		return []string{"schema", "mission_id", "highest_proven_live_class", "next_denied_class", "safe_to_execute", "exact_next_action", "generated_at_utc"}
	case RouteSchema:
		return []string{"schema", "mission_id", "route", "reason", "safe_to_execute"}
	case SchedulerReadbackSchema:
		return []string{"schema", "mission_id", "status", "scheduler", "event_loop"}
	case TelegramCommandSchema:
		return []string{"schema", "chat_id", "command", "role"}
	case A2ATaskSchema:
		return []string{"schema", "task_id", "method", "status"}
	case ArtifactRefSchema:
		return []string{"schema", "ref"}
	default:
		return []string{"schema"}
	}
}
