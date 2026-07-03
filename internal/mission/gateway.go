package mission

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
)

type TelegramConfig struct {
	Schema       string            `json:"schema"`
	TokenEnv     string            `json:"token_env"`
	AllowedChats map[string]string `json:"allowed_chats"`
}

type GatewayReadback struct {
	Schema            string   `json:"schema"`
	Gateway           string   `json:"gateway"`
	Status            string   `json:"status"`
	AllowedChatCount  int      `json:"allowed_chat_count,omitempty"`
	Methods           []string `json:"methods,omitempty"`
	Message           string   `json:"message"`
	MutationAuthority bool     `json:"mutation_authority"`
	GeneratedAtUTC    string   `json:"generated_at_utc"`
}

var allowedTelegramCommands = map[string]bool{
	"/status":   true,
	"/next":     true,
	"/continue": true,
	"/pause":    true,
	"/resume":   true,
	"/stop":     true,
	"/approve":  true,
	"/deny":     true,
	"/where":    true,
	"/help":     true,
}

func LoadTelegramConfig(path string) (TelegramConfig, error) {
	var cfg TelegramConfig
	body, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	if err := ValidatePublicSafeText(string(body)); err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(body, &cfg); err != nil {
		return cfg, err
	}
	if cfg.Schema == "" {
		cfg.Schema = "ao.mission.telegram-config.v0.1"
	}
	if cfg.AllowedChats == nil {
		cfg.AllowedChats = map[string]string{}
	}
	return cfg, nil
}

func TelegramConfigReadback(cfg TelegramConfig) GatewayReadback {
	return GatewayReadback{
		Schema:            "ao.mission.gateway-readback.v0.1",
		Gateway:           "telegram",
		Status:            "configured_intent_only",
		AllowedChatCount:  len(cfg.AllowedChats),
		Message:           "telegram gateway records intents and readbacks only; token values are read from environment and never printed",
		MutationAuthority: false,
		GeneratedAtUTC:    now(nil),
	}
}

func LoadTelegramCommandMatrix(path string) (TelegramCommandMatrix, error) {
	var matrix TelegramCommandMatrix
	body, err := os.ReadFile(path)
	if err != nil {
		return matrix, err
	}
	if err := ValidatePublicSafeText(string(body)); err != nil {
		return matrix, err
	}
	if err := json.Unmarshal(body, &matrix); err != nil {
		return matrix, err
	}
	return matrix, nil
}

func HandleTelegramCommand(cmd TelegramCommand, allowlist map[string]string) TelegramReadback {
	if cmd.Schema == "" {
		cmd.Schema = TelegramCommandSchema
	}
	role, ok := allowlist[cmd.ChatID]
	if !ok {
		return TelegramReadback{Schema: TelegramReadbackSchema, Status: "denied", Message: "chat id is not allowlisted", MutationAuthority: false}
	}
	if role != "admin" && (cmd.Command == "/approve" || cmd.Command == "/continue") {
		return TelegramReadback{Schema: TelegramReadbackSchema, Status: "denied", Message: "role cannot request this intent", MutationAuthority: false}
	}
	if !strings.HasPrefix(cmd.Command, "/") {
		return TelegramReadback{Schema: TelegramReadbackSchema, Status: "invalid", Message: "telegram command must start with slash", MutationAuthority: false}
	}
	if !allowedTelegramCommands[cmd.Command] {
		return TelegramReadback{Schema: TelegramReadbackSchema, Status: "invalid", Message: "telegram command is not supported", MutationAuthority: false}
	}
	return TelegramReadback{Schema: TelegramReadbackSchema, Status: "intent_recorded", Message: "telegram gateway records intents and readbacks only", MutationAuthority: false}
}

func AgentCard() A2AAgentCard {
	return A2AAgentCard{Schema: A2AAgentCardSchema, Name: "ao-mission", Methods: []string{"mission.start", "mission.status", "mission.next", "mission.continue", "mission.pause", "mission.resume", "mission.cancel", "mission.artifacts", "mission.governance_snapshot"}, MutationAuthority: false}
}
func A2ATaskFor(method string) A2ATask {
	return A2ATask{Schema: A2ATaskSchema, TaskID: "task-" + strings.ReplaceAll(method, ".", "-"), Method: method, Status: "intent_recorded", MutationAuthority: false}
}
func A2ATaskForParams(method string, params map[string]any) A2ATask {
	task := A2ATaskFor(method)
	if !stringSliceContains(AgentCard().Methods, method) {
		task.Status = "invalid"
		return task
	}
	switch method {
	case "mission.start":
		if strings.TrimSpace(stringParam(params, "objective")) == "" {
			task.Status = "invalid"
		}
	case "mission.status", "mission.next", "mission.continue", "mission.pause", "mission.resume", "mission.cancel", "mission.artifacts", "mission.governance_snapshot":
		if strings.TrimSpace(stringParam(params, "mission_id")) == "" {
			task.Status = "invalid"
		}
	}
	return task
}

func A2AHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/agent-card.json", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(AgentCard())
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(GatewayReadback{Schema: "ao.mission.gateway-readback.v0.1", Gateway: "a2a", Status: "ready", Methods: AgentCard().Methods, Message: "A2A local gateway is intent/readback only", MutationAuthority: false, GeneratedAtUTC: now(nil)})
			return
		}
		var req struct {
			JSONRPC string         `json:"jsonrpc"`
			ID      any            `json:"id"`
			Method  string         `json:"method"`
			Params  map[string]any `json:"params"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.Method == "" {
			req.Method = "mission.status"
		}
		w.Header().Set("Content-Type", "application/json")
		task := A2ATaskForParams(req.Method, req.Params)
		if req.JSONRPC == "2.0" || req.ID != nil {
			_ = json.NewEncoder(w).Encode(A2AJSONRPCResponse{JSONRPC: "2.0", ID: req.ID, Result: task})
			return
		}
		_ = json.NewEncoder(w).Encode(task)
	})
	return mux
}

func stringSliceContains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func stringParam(params map[string]any, key string) string {
	if params == nil {
		return ""
	}
	value, _ := params[key].(string)
	return value
}
