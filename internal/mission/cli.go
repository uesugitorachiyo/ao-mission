package mission

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
)

func Run(args []string, stdout, stderr io.Writer) int {
	if err := run(args, stdout); err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 1
	}
	return 0
}
func printJSON(w io.Writer, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	_, err = w.Write(append(b, '\n'))
	return err
}

func run(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return errors.New("usage: ao-mission <init|start|continue|status|next|stop|pause|resume|schedule|daemon|telegram|a2a|governance|artifacts>")
	}
	s := NewStore("")
	switch args[0] {
	case "init":
		if err := s.Init(); err != nil {
			return err
		}
		fmt.Fprintln(stdout, "status=initialized")
		return nil
	case "start":
		if len(args) < 2 {
			return errors.New("start requires objective")
		}
		r, err := s.Start(strings.Join(args[1:], " "))
		if err != nil {
			return err
		}
		return printJSON(stdout, r)
	case "status":
		fs := flag.NewFlagSet("status", flag.ContinueOnError)
		id := fs.String("mission", "", "")
		jsonOut := fs.Bool("json", false, "")
		_ = fs.Parse(args[1:])
		r, err := s.Load(*id)
		if err != nil {
			return err
		}
		if *jsonOut {
			return printJSON(stdout, r)
		}
		fmt.Fprintf(stdout, "mission=%s\nstatus=%s\nroute=%s\nnext=%s\n", r.MissionID, r.Status, r.CurrentRoute, r.ExactNextAction)
		return nil
	case "next":
		fs := flag.NewFlagSet("next", flag.ContinueOnError)
		id := fs.String("mission", "", "")
		jsonOut := fs.Bool("json", false, "")
		_ = fs.Parse(args[1:])
		r, err := s.Load(*id)
		if err != nil {
			return err
		}
		d := NextAction(r)
		if *jsonOut {
			return printJSON(stdout, d)
		}
		fmt.Fprintf(stdout, "route=%s\nreason=%s\nnext=%s\n", d.Route, d.Reason, d.ExactNextAction)
		return nil
	case "continue":
		fs := flag.NewFlagSet("continue", flag.ContinueOnError)
		id := fs.String("mission", "", "")
		until := fs.Bool("until-done", false, "")
		max := fs.Int("max-iterations", 1, "")
		_ = fs.Parse(args[1:])
		r, err := Continue(s, *id, ContinueOptions{UntilDone: *until, MaxIterations: *max})
		if err != nil {
			return err
		}
		return printJSON(stdout, r)
	case "pause":
		id := missionFlag(args[1:])
		r, err := Pause(s, id)
		if err != nil {
			return err
		}
		return printJSON(stdout, r)
	case "resume":
		id := missionFlag(args[1:])
		r, err := Resume(s, id)
		if err != nil {
			return err
		}
		return printJSON(stdout, r)
	case "stop":
		id := missionFlag(args[1:])
		r, err := Stop(s, id)
		if err != nil {
			return err
		}
		return printJSON(stdout, r)
	case "schedule":
		fs := flag.NewFlagSet("schedule", flag.ContinueOnError)
		id := fs.String("mission", "", "")
		every := fs.String("every", "", "")
		eventLoop := fs.Bool("event-loop", false, "")
		_ = fs.Parse(args[1:])
		_ = every
		return printJSON(stdout, ScheduleReadback(*id, *every, *eventLoop))
	case "daemon":
		if len(args) < 2 {
			return errors.New("daemon requires install/status/uninstall")
		}
		fmt.Fprintf(stdout, "daemon=%s\nstatus=readback_only\n", args[1])
		return nil
	case "telegram":
		if len(args) >= 2 && args[1] == "serve" {
			return printJSON(stdout, TelegramReadback{Schema: TelegramReadbackSchema, Status: "disabled", Message: "telegram gateway disabled by default; configure fake-token-safe environment and allowlist", MutationAuthority: false})
		}
		return errors.New("telegram requires serve")
	case "a2a":
		if len(args) >= 2 && args[1] == "serve" {
			return printJSON(stdout, AgentCard())
		}
		return errors.New("a2a requires serve")
	case "governance":
		if len(args) >= 2 && args[1] == "snapshot" {
			id := missionFlag(args[2:])
			r, err := s.Load(id)
			if err != nil {
				return err
			}
			return printJSON(stdout, Snapshot(r))
		}
		return errors.New("governance requires snapshot")
	case "artifacts":
		id := missionFlag(args[1:])
		r, err := s.Load(id)
		if err != nil {
			return err
		}
		return printJSON(stdout, r.ArtifactRefs)
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}
func missionFlag(args []string) string {
	fs := flag.NewFlagSet("mission", flag.ContinueOnError)
	id := fs.String("mission", "", "")
	_ = fs.Parse(args)
	return *id
}
