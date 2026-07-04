package mission

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
	required, propertyTypes := contractRules(schema)
	for _, field := range required {
		if _, ok := doc[field]; !ok {
			result.Status = "blocked"
			result.Blockers = append(result.Blockers, field+" is required")
		}
	}
	for field, want := range propertyTypes {
		value, ok := doc[field]
		if !ok {
			continue
		}
		if !jsonTypeMatches(value, want) {
			result.Status = "blocked"
			result.Blockers = append(result.Blockers, fmt.Sprintf("%s must be %s", field, want))
		}
	}
	if blockers := validateAgainstSchemaFile(path, doc, schema); len(blockers) > 0 {
		result.Status = "blocked"
		result.Blockers = append(result.Blockers, blockers...)
	}
	if result.Status != "ready" {
		return result, fmt.Errorf(strings.Join(result.Blockers, "; "))
	}
	return result, nil
}

func contractRules(schema string) ([]string, map[string]string) {
	return requiredFieldsForContract(schema), propertyTypesForContract(schema)
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
	case A2AAgentCardSchema:
		return []string{"schema", "name", "protocol_version", "description", "endpoint", "methods", "capabilities", "mutation_authority"}
	case A2ATaskSchema:
		return []string{"schema", "task_id", "method", "status"}
	case ArtifactRefSchema:
		return []string{"schema", "ref"}
	default:
		return []string{"schema"}
	}
}

func propertyTypesForContract(schema string) map[string]string {
	commonString := map[string]string{"schema": "string"}
	switch schema {
	case RecordSchema:
		return map[string]string{"schema": "string", "mission_id": "string", "objective": "string", "objective_digest": "string", "status": "string", "created_at_utc": "string", "updated_at_utc": "string", "current_route": "string", "current_phase": "string", "blockers": "array", "exact_next_action": "string", "artifact_refs": "array", "steps": "array"}
	case SnapshotSchema:
		return map[string]string{"schema": "string", "mission_id": "string", "highest_proven_live_class": "string", "next_denied_class": "string", "safe_to_execute": "boolean", "exact_next_action": "string", "generated_at_utc": "string"}
	case RouteSchema:
		return map[string]string{"schema": "string", "mission_id": "string", "route": "string", "reason": "string", "safe_to_execute": "boolean"}
	case SchedulerReadbackSchema:
		return map[string]string{"schema": "string", "mission_id": "string", "status": "string", "scheduler": "string", "event_loop": "boolean"}
	case TelegramCommandSchema:
		return map[string]string{"schema": "string", "chat_id": "string", "command": "string", "role": "string"}
	case A2AAgentCardSchema:
		return map[string]string{"schema": "string", "name": "string", "protocol_version": "string", "description": "string", "endpoint": "string", "methods": "array", "capabilities": "array", "mutation_authority": "boolean"}
	case A2ATaskSchema:
		return map[string]string{"schema": "string", "task_id": "string", "method": "string", "status": "string", "mutation_authority": "boolean"}
	case ArtifactRefSchema:
		return map[string]string{"schema": "string", "ref": "string", "digest": "string", "kind": "string"}
	default:
		return commonString
	}
}

func validateAgainstSchemaFile(path string, doc map[string]any, schema string) []string {
	schemaPath := filepath.Join("docs", "contracts", contractFileName(schema))
	if _, err := os.Stat(schemaPath); err != nil {
		alt := filepath.Join(filepath.Dir(path), "..", "..", "docs", "contracts", contractFileName(schema))
		if _, altErr := os.Stat(alt); altErr != nil {
			return nil
		}
		schemaPath = alt
	}
	body, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil
	}
	var schemaDoc map[string]any
	if err := json.Unmarshal(body, &schemaDoc); err != nil {
		return nil
	}
	blockers := []string{}
	if required, ok := schemaDoc["required"].([]any); ok {
		for _, item := range required {
			field, _ := item.(string)
			if field == "" {
				continue
			}
			if _, exists := doc[field]; !exists {
				blockers = append(blockers, field+" is required")
			}
		}
	}
	if props, ok := schemaDoc["properties"].(map[string]any); ok {
		for field, propAny := range props {
			value, exists := doc[field]
			if !exists {
				continue
			}
			prop, _ := propAny.(map[string]any)
			want, _ := prop["type"].(string)
			if want != "" && !jsonTypeMatches(value, want) {
				blockers = append(blockers, fmt.Sprintf("%s must be %s", field, want))
			}
			if constValue, hasConst := prop["const"]; hasConst && value != constValue {
				blockers = append(blockers, fmt.Sprintf("%s must equal %v", field, constValue))
			}
		}
	}
	return blockers
}

func contractFileName(schema string) string {
	name := strings.TrimPrefix(schema, "ao.mission.")
	name = strings.ReplaceAll(name, ".v", "-v")
	return name + ".schema.json"
}

func jsonTypeMatches(value any, want string) bool {
	switch want {
	case "string":
		_, ok := value.(string)
		return ok
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "object":
		_, ok := value.(map[string]any)
		return ok
	case "array":
		_, ok := value.([]any)
		return ok
	case "integer":
		f, ok := value.(float64)
		return ok && f == float64(int(f))
	case "number":
		_, ok := value.(float64)
		return ok
	default:
		return true
	}
}
