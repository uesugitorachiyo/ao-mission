package mission

import "errors"

type ContinueOptions struct {
	UntilDone        bool
	MaxIterations    int
	MinNodes         int
	MinMinutes       int
	MaxMinutes       int
	ReturnOnlyWhen   string
	CheckpointPolicy string
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
		lease := ensureGoalLease(r, opts)
		for i := 0; i < opts.MaxIterations; i++ {
			if r.Status == "done" || hardBlockerExists(*r) {
				break
			}
			decision := NextActionForRecord(*r)
			step := ContinuationStep{Schema: StepSchema, MissionID: r.MissionID, CorrelationID: r.CorrelationID, Iteration: len(r.Steps) + 1, Route: decision.Route, Result: "handoff_required", ExactNextAction: decision.ExactNextAction, GeneratedAtUTC: now(s.Clock)}
			r.Steps = append(r.Steps, step)
			r.CurrentRoute = decision.Route
			r.CurrentPhase = "handoff_required"
			r.ExactNextAction = decision.ExactNextAction
			appendMissionCheckpoint(r, step)
			gate := EvaluateReturnGate(*r)
			r.ReturnGate = &gate
			reconciliation := BuildRouteReconciliation(*r)
			r.Reconciliation = &reconciliation
			if err := s.SaveEventLoopDecision(EventLoopDecision{
				Schema:              EventLoopDecisionSchema,
				MissionID:           r.MissionID,
				CorrelationID:       r.CorrelationID,
				Iteration:           step.Iteration,
				Status:              step.Result,
				Route:               step.Route,
				ExactNextAction:     step.ExactNextAction,
				ExecutesWork:        false,
				ApprovesWork:        false,
				MutatesRepositories: false,
				GeneratedAtUTC:      step.GeneratedAtUTC,
			}); err != nil {
				return err
			}
			if err := s.SaveCheckpointBundle(BuildCheckpointBundle(*r)); err != nil {
				return err
			}
			if !opts.UntilDone {
				break
			}
			if r.ReturnGate != nil && r.ReturnGate.FinalResponseAllowed && len(r.Steps) >= lease.MinNodes {
				break
			}
		}
		gate := EvaluateReturnGate(*r)
		r.ReturnGate = &gate
		reconciliation := BuildRouteReconciliation(*r)
		r.Reconciliation = &reconciliation
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
		if r.WorkflowContract != nil && r.CurrentRoute == r.WorkflowContract.InitialRoute {
			r.ExactNextAction = r.WorkflowContract.ExactNextAction
		} else {
			r.ExactNextAction = NextActionForRecord(*r).ExactNextAction
		}
		gate := EvaluateReturnGate(*r)
		r.ReturnGate = &gate
		reconciliation := BuildRouteReconciliation(*r)
		r.Reconciliation = &reconciliation
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
