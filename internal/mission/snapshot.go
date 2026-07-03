package mission

func Snapshot(r Record) GovernanceSnapshot {
	return GovernanceSnapshot{Schema: SnapshotSchema, MissionID: r.MissionID, ObjectiveDigest: r.ObjectiveDigest, CurrentOwner: "ao-mission", CurrentPhase: r.CurrentPhase, CurrentRoute: r.CurrentRoute, HighestProvenLiveClass: "public_safe_unrestricted_self_modification_authority_request_dry_run_four_attempts", NextDeniedClass: "unrestricted_self_modification", SafeToRequest: true, SafeToExecute: false, SafeToPromote: false, SchedulesWork: false, ExecutesWork: false, ApprovesWork: false, MutatesRepositories: false, ProviderCalls: false, ReleaseOrPublish: false, CredentialUse: false, DirectMainMutation: false, ConcurrentMutation: false, SentinelStatus: "not_requested", PromoterStatus: "not_requested", CovenantStatus: "not_requested", RollbackStatus: "not_required", CIStatus: "not_started", RepoHygieneStatus: "not_started", EvidenceFreshnessStatus: "fresh", KillSwitchStatus: killSwitchStatus(r), Blockers: r.Blockers, ExactNextAction: r.ExactNextAction, ArtifactRefs: r.ArtifactRefs, GeneratedAtUTC: now(nil)}
}
func killSwitchStatus(r Record) string {
	if r.Status == "stopped" {
		return "engaged"
	}
	return "clear"
}
