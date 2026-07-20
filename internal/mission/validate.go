package mission

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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
	if schema == ObjectiveWorkflowSchema {
		if blockers := validateObjectiveWorkflowSemantics(doc); len(blockers) > 0 {
			result.Status = "blocked"
			result.Blockers = append(result.Blockers, blockers...)
		}
	}
	if result.Status != "ready" {
		return result, fmt.Errorf(strings.Join(result.Blockers, "; "))
	}
	return result, nil
}

func validateObjectiveWorkflowSemantics(doc map[string]any) []string {
	routingClass := stringFromAny(doc["routing_class"])
	var acceptance, route, next string
	var stages []ObjectiveWorkflowStage
	switch routingClass {
	case "pending_blueprint":
		acceptance = "pending_blueprint"
		route = "ao-blueprint"
		next = "send objective to AO Blueprint for requirements and acceptance"
		stages = objectiveStages("required", "conditional", "conditional")
	case "complex":
		acceptance = "accepted"
		route = "ao-atlas"
		next = "send accepted objective to AO Atlas for workgraph sequencing"
		stages = objectiveStages("omitted", "required", "conditional")
	case "reduced":
		acceptance = "accepted"
		route = "ao-foundry"
		next = "send accepted reduced objective directly to AO Foundry"
		stages = objectiveStages("omitted", "omitted", "required")
	default:
		return []string{"routing_class is unsupported"}
	}
	blockers := []string{}
	if stringFromAny(doc["acceptance_status"]) != acceptance {
		blockers = append(blockers, "acceptance_status does not match routing_class")
	}
	if stringFromAny(doc["initial_route"]) != route {
		blockers = append(blockers, "initial_route does not match routing_class")
	}
	if stringFromAny(doc["exact_next_action"]) != next {
		blockers = append(blockers, "exact_next_action does not match routing_class")
	}
	rawStages, _ := doc["stages"].([]any)
	if len(rawStages) != len(stages) {
		blockers = append(blockers, "stages do not contain the required workflow sequence")
	} else {
		for i, expected := range stages {
			stage, _ := rawStages[i].(map[string]any)
			if stringFromAny(stage["name"]) != expected.Name ||
				stringFromAny(stage["status"]) != expected.Status ||
				stringFromAny(stage["reason"]) != expected.Reason {
				blockers = append(blockers, fmt.Sprintf("stages.%d does not match routing_class", i))
			}
		}
	}
	missionID := stringFromAny(doc["mission_id"])
	rawCommands, _ := doc["lifecycle_commands"].([]any)
	expectedCommands := objectiveLifecycleCommands(missionID)
	if len(rawCommands) != len(expectedCommands) {
		blockers = append(blockers, "lifecycle_commands do not contain the required command sequence")
	} else {
		for i, expected := range expectedCommands {
			if stringFromAny(rawCommands[i]) != expected {
				blockers = append(blockers, fmt.Sprintf("lifecycle_commands.%d does not bind mission_id", i))
			}
		}
	}
	return blockers
}

func validateRecordWorkflowContract(record Record) error {
	contract := record.WorkflowContract
	if contract == nil {
		return nil
	}
	switch {
	case contract.Schema != ObjectiveWorkflowSchema:
		return fmt.Errorf("workflow contract schema must be %s", ObjectiveWorkflowSchema)
	case contract.Status != "ready":
		return fmt.Errorf("workflow contract status must be ready")
	case record.CorrelationID == "" || contract.CorrelationID != record.CorrelationID:
		return fmt.Errorf("workflow contract correlation_id does not match mission record")
	case contract.MissionID != record.MissionID:
		return fmt.Errorf("workflow contract mission_id does not match mission record")
	case record.ObjectiveDigest != DigestObjective(record.Objective) &&
		(!record.ObjectiveRedacted || !strings.Contains(record.Objective, "<local-path-redacted>")):
		return fmt.Errorf("mission record objective_digest does not match objective")
	case contract.ObjectiveDigest != record.ObjectiveDigest:
		return fmt.Errorf("workflow contract objective_digest does not match mission record")
	case contract.SafeToExecute || contract.ExecutesWork || contract.ApprovesWork || contract.MutatesRepositories:
		return fmt.Errorf("workflow contract must not claim execution, approval, or mutation authority")
	}
	body, err := json.Marshal(contract)
	if err != nil {
		return err
	}
	var doc map[string]any
	if err := json.Unmarshal(body, &doc); err != nil {
		return err
	}
	if blockers := validateObjectiveWorkflowSemantics(doc); len(blockers) > 0 {
		return fmt.Errorf("workflow contract is inconsistent: %s", strings.Join(blockers, "; "))
	}
	return nil
}

func contractRules(schema string) ([]string, map[string]string) {
	return requiredFieldsForContract(schema), propertyTypesForContract(schema)
}

func requiredFieldsForContract(schema string) []string {
	switch schema {
	case ObjectiveWorkflowSchema:
		return []string{
			"schema", "status", "mission_id", "correlation_id", "objective_digest",
			"routing_class", "acceptance_status", "initial_route", "stages",
			"lifecycle_commands", "exact_next_action", "safe_to_execute",
			"executes_work", "approves_work", "mutates_repositories", "generated_at_utc",
		}
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
	case ObjectiveWorkflowSchema:
		return map[string]string{
			"schema": "string", "status": "string", "mission_id": "string",
			"correlation_id": "string", "objective_digest": "string",
			"routing_class": "string", "acceptance_status": "string",
			"initial_route": "string", "stages": "array",
			"lifecycle_commands": "array", "exact_next_action": "string",
			"safe_to_execute": "boolean", "executes_work": "boolean",
			"approves_work": "boolean", "mutates_repositories": "boolean",
			"generated_at_utc": "string",
		}
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
	name := contractFileName(schema)
	candidates := []string{
		filepath.Join("docs", "contracts", name),
		filepath.Join("..", "..", "docs", "contracts", name),
		filepath.Join(filepath.Dir(path), "..", "..", "docs", "contracts", name),
	}
	schemaPath := ""
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			schemaPath = candidate
			break
		}
	}
	if schemaPath == "" {
		return nil
	}
	body, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil
	}
	var schemaDoc map[string]any
	if err := json.Unmarshal(body, &schemaDoc); err != nil {
		return nil
	}
	return validateJSONSchemaNode(doc, schemaDoc, "")
}

func validateJSONSchemaNode(value any, schema map[string]any, path string) []string {
	blockers := []string{}
	label := strings.TrimSuffix(path, ".")
	if label == "" {
		label = "document"
	}
	if want, _ := schema["type"].(string); want != "" && !jsonTypeMatches(value, want) {
		return []string{fmt.Sprintf("%s must be %s", label, want)}
	}
	if constValue, ok := schema["const"]; ok && value != constValue {
		blockers = append(blockers, fmt.Sprintf("%s must equal %v", label, constValue))
	}
	if allowed, ok := schema["enum"].([]any); ok {
		found := false
		for _, candidate := range allowed {
			if value == candidate {
				found = true
				break
			}
		}
		if !found {
			blockers = append(blockers, fmt.Sprintf("%s has unsupported value", label))
		}
	}
	switch typed := value.(type) {
	case map[string]any:
		props, _ := schema["properties"].(map[string]any)
		if required, ok := schema["required"].([]any); ok {
			for _, item := range required {
				field, _ := item.(string)
				if field == "" {
					continue
				}
				if _, exists := typed[field]; !exists {
					blockers = append(blockers, path+field+" is required")
				}
			}
		}
		if additional, ok := schema["additionalProperties"].(bool); ok && !additional {
			for field := range typed {
				if _, exists := props[field]; !exists {
					blockers = append(blockers, path+field+" is not allowed")
				}
			}
		}
		for field, child := range typed {
			childSchema, ok := props[field].(map[string]any)
			if !ok {
				continue
			}
			blockers = append(blockers, validateJSONSchemaNode(child, childSchema, path+field+".")...)
		}
	case []any:
		if min, ok := schema["minItems"].(float64); ok && len(typed) < int(min) {
			blockers = append(blockers, fmt.Sprintf("%s must contain at least %d items", label, int(min)))
		}
		if max, ok := schema["maxItems"].(float64); ok && len(typed) > int(max) {
			blockers = append(blockers, fmt.Sprintf("%s must contain at most %d items", label, int(max)))
		}
		if itemSchema, ok := schema["items"].(map[string]any); ok {
			for i, item := range typed {
				blockers = append(blockers, validateJSONSchemaNode(item, itemSchema, fmt.Sprintf("%s%d.", path, i))...)
			}
		}
	case string:
		if min, ok := schema["minLength"].(float64); ok && len(typed) < int(min) {
			blockers = append(blockers, fmt.Sprintf("%s is too short", label))
		}
		if max, ok := schema["maxLength"].(float64); ok && len(typed) > int(max) {
			blockers = append(blockers, fmt.Sprintf("%s is too long", label))
		}
		if pattern, ok := schema["pattern"].(string); ok {
			re, err := regexp.Compile(pattern)
			if err == nil && !re.MatchString(typed) {
				blockers = append(blockers, fmt.Sprintf("%s does not match required pattern", label))
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
