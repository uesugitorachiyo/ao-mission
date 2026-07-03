package mission

import "strings"

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
	return TelegramReadback{Schema: TelegramReadbackSchema, Status: "intent_recorded", Message: "telegram gateway records intents and readbacks only", MutationAuthority: false}
}

func AgentCard() A2AAgentCard {
	return A2AAgentCard{Schema: A2AAgentCardSchema, Name: "ao-mission", Methods: []string{"mission.start", "mission.status", "mission.next", "mission.continue", "mission.pause", "mission.resume", "mission.cancel", "mission.artifacts", "mission.governance_snapshot"}, MutationAuthority: false}
}
func A2ATaskFor(method string) A2ATask {
	return A2ATask{Schema: A2ATaskSchema, TaskID: "task-" + strings.ReplaceAll(method, ".", "-"), Method: method, Status: "intent_recorded", MutationAuthority: false}
}
