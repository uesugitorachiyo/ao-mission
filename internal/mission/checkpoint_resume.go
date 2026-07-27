package mission

import "fmt"

const (
	MissionCheckpointSchema = "ao.mission.checkpoint.v0.3"
	CheckpointBundleSchema  = "ao.mission.checkpoint-resume-bundle.v0.3"
)

func appendMissionCheckpoint(r *Record, step ContinuationStep) MissionCheckpoint {
	checkpoint := MissionCheckpoint{
		Schema:          MissionCheckpointSchema,
		MissionID:       r.MissionID,
		CorrelationID:   r.CorrelationID,
		Sequence:        len(r.Checkpoints) + 1,
		Iteration:       step.Iteration,
		Route:           step.Route,
		Phase:           r.CurrentPhase,
		Result:          step.Result,
		ExactNextAction: step.ExactNextAction,
		ResumeCommand:   fmt.Sprintf("ao-mission continue --mission %s --until-done --max-iterations 10", r.MissionID),
		GeneratedAtUTC:  step.GeneratedAtUTC,
	}
	r.Checkpoints = append(r.Checkpoints, checkpoint)
	return checkpoint
}

func BuildCheckpointBundle(r Record) MissionCheckpointBundle {
	var latest *MissionCheckpoint
	if n := len(r.Checkpoints); n > 0 {
		cp := r.Checkpoints[n-1]
		latest = &cp
	}
	gate := EvaluateReturnGate(r)
	return MissionCheckpointBundle{
		Schema:              CheckpointBundleSchema,
		MissionID:           r.MissionID,
		CorrelationID:       r.CorrelationID,
		Status:              "ready",
		CheckpointCount:     len(r.Checkpoints),
		LatestCheckpoint:    latest,
		ReturnGate:          &gate,
		ResumePrompt:        fmt.Sprintf("ao-mission continue --mission %s --until-done --max-iterations 10", r.MissionID),
		SafeToExecute:       false,
		ExecutesWork:        false,
		ApprovesWork:        false,
		MutatesRepositories: false,
		GeneratedAtUTC:      now(nil),
	}
}
