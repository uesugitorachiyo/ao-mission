package mission

type FinalRollup struct {
	Schema                  string        `json:"schema"`
	MissionID               string        `json:"mission_id"`
	Status                  string        `json:"status"`
	CompletedSteps          int           `json:"completed_steps"`
	ArtifactRefs            []ArtifactRef `json:"artifact_refs"`
	HighestProvenLiveClass  string        `json:"highest_proven_live_class"`
	NextDeniedClass         string        `json:"next_denied_class"`
	SafeToExecute           bool          `json:"safe_to_execute"`
	ExecutesWork            bool          `json:"executes_work"`
	ApprovesWork            bool          `json:"approves_work"`
	ProviderCalls           bool          `json:"provider_calls"`
	ExactNextAction         string        `json:"exact_next_action"`
	RemainingDeniedSurfaces []string      `json:"remaining_denied_surfaces"`
	GeneratedAtUTC          string        `json:"generated_at_utc"`
}

func BuildFinalRollup(r Record) FinalRollup {
	return FinalRollup{
		Schema:                 "ao.mission.final-rollup.v0.1",
		MissionID:              r.MissionID,
		Status:                 r.Status,
		CompletedSteps:         len(r.Steps),
		ArtifactRefs:           r.ArtifactRefs,
		HighestProvenLiveClass: "public_safe_unrestricted_self_modification_authority_request_dry_run_four_attempts",
		NextDeniedClass:        "unrestricted_self_modification",
		SafeToExecute:          false,
		ExecutesWork:           false,
		ApprovesWork:           false,
		ProviderCalls:          false,
		ExactNextAction:        r.ExactNextAction,
		RemainingDeniedSurfaces: []string{
			"unrestricted_self_modification",
			"hidden_instruction_mutation",
			"policy_changing_autonomy",
			"provider_calls",
			"credential_use",
			"release_deploy_publish_upload_tag",
			"dependency_updates",
			"direct_main_mutation",
			"concurrent_mutation",
			"broad_RSI",
		},
		GeneratedAtUTC: now(nil),
	}
}
