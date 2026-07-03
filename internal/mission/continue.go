package mission

import "errors"

type ContinueOptions struct {
	UntilDone     bool
	MaxIterations int
}

func Continue(s Store, missionID string, opts ContinueOptions) (Record, error) {
	if opts.MaxIterations <= 0 {
		opts.MaxIterations = 1
	}
	return s.Update(missionID, func(r *Record) error {
		if r.Status == "stopped" {
			return errors.New("mission is stopped")
		}
		if r.Status == "paused" {
			return errors.New("mission is paused")
		}
		for i := 0; i < opts.MaxIterations; i++ {
			decision := NextAction(*r)
			step := ContinuationStep{Schema: StepSchema, MissionID: r.MissionID, Iteration: len(r.Steps) + 1, Route: decision.Route, Result: "handoff_required", ExactNextAction: decision.ExactNextAction, GeneratedAtUTC: now(s.Clock)}
			r.Steps = append(r.Steps, step)
			r.CurrentRoute = decision.Route
			r.CurrentPhase = "handoff_required"
			r.ExactNextAction = decision.ExactNextAction
			if !opts.UntilDone {
				break
			}
			// v0.1 does not execute downstream systems itself; stop the zero-wait loop at the first required governed handoff.
			break
		}
		return nil
	})
}

func Pause(s Store, id string) (Record, error) {
	return s.Update(id, func(r *Record) error {
		r.Status = "paused"
		r.CurrentPhase = "paused"
		r.ExactNextAction = "resume mission before continuation"
		return nil
	})
}
func Resume(s Store, id string) (Record, error) {
	return s.Update(id, func(r *Record) error {
		r.Status = "active"
		r.CurrentPhase = "routing"
		r.ExactNextAction = NextAction(*r).ExactNextAction
		return nil
	})
}
func Stop(s Store, id string) (Record, error) {
	return s.Update(id, func(r *Record) error {
		r.Status = "stopped"
		r.CurrentPhase = "stopped"
		r.ExactNextAction = "mission stopped by operator kill switch"
		return nil
	})
}
