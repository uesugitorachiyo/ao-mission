package mission

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
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
		return errors.New("usage: ao-mission [--home <dir>] <init|start|mission|continue|status|next|stop|pause|resume|schedule|daemon|telegram|a2a|governance|command|artifacts|validate|import|final>")
	}
	home, args, err := parseGlobalHome(args)
	if err != nil {
		return err
	}
	if len(args) == 0 {
		return errors.New("command is required")
	}
	s := NewStore(home)
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
	case "mission":
		if len(args) < 2 {
			return errors.New("mission requires list or inspect")
		}
		switch args[1] {
		case "list":
			fs := flag.NewFlagSet("mission list", flag.ContinueOnError)
			jsonOut := fs.Bool("json", false, "")
			_ = fs.Parse(args[2:])
			records, err := s.List()
			if err != nil {
				return err
			}
			if *jsonOut {
				return printJSON(stdout, records)
			}
			for _, rec := range records {
				fmt.Fprintf(stdout, "mission=%s status=%s route=%s\n", rec.MissionID, rec.Status, rec.CurrentRoute)
			}
			return nil
		case "inspect":
			fs := flag.NewFlagSet("mission inspect", flag.ContinueOnError)
			id := fs.String("mission", "", "")
			jsonOut := fs.Bool("json", false, "")
			_ = fs.Parse(args[2:])
			r, err := s.Load(*id)
			if err != nil {
				return err
			}
			if *jsonOut {
				return printJSON(stdout, r)
			}
			fmt.Fprintf(stdout, "mission=%s\nstatus=%s\nphase=%s\nroute=%s\nnext=%s\n", r.MissionID, r.Status, r.CurrentPhase, r.CurrentRoute, r.ExactNextAction)
			return nil
		default:
			return errors.New("mission requires list or inspect")
		}
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
			fs := flag.NewFlagSet("telegram serve", flag.ContinueOnError)
			configPath := fs.String("config", "", "")
			_ = fs.Parse(args[2:])
			if *configPath == "" {
				return printJSON(stdout, TelegramReadback{Schema: TelegramReadbackSchema, Status: "disabled", Message: "telegram gateway disabled by default; configure environment token name and allowlist", MutationAuthority: false})
			}
			cfg, err := LoadTelegramConfig(*configPath)
			if err != nil {
				return err
			}
			return printJSON(stdout, TelegramConfigReadback(cfg))
		}
		return errors.New("telegram requires serve")
	case "a2a":
		if len(args) >= 2 && args[1] == "serve" {
			fs := flag.NewFlagSet("a2a serve", flag.ContinueOnError)
			httpMode := fs.Bool("http", false, "")
			listen := fs.String("listen", "127.0.0.1:0", "")
			once := fs.Bool("once", false, "")
			_ = fs.Parse(args[2:])
			if !*httpMode {
				return printJSON(stdout, AgentCard())
			}
			ln, err := net.Listen("tcp", *listen)
			if err != nil {
				return err
			}
			server := &http.Server{Handler: A2AHandler()}
			if *once {
				addr := ln.Addr().String()
				_ = ln.Close()
				return printJSON(stdout, GatewayReadback{Schema: "ao.mission.gateway-readback.v0.1", Gateway: "a2a", Status: "ready", Methods: AgentCard().Methods, Message: "A2A local HTTP fixture server can bind at " + addr + " and records intents only", MutationAuthority: false, GeneratedAtUTC: now(nil)})
			}
			fmt.Fprintf(stdout, "a2a_listen=%s\nmutation_authority=false\n", ln.Addr().String())
			return server.Serve(ln)
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
	case "command":
		if len(args) >= 2 && args[1] == "status" {
			fs := flag.NewFlagSet("command status", flag.ContinueOnError)
			id := fs.String("mission", "", "")
			jsonOut := fs.Bool("json", false, "")
			_ = fs.Parse(args[2:])
			r, err := s.Load(*id)
			if err != nil {
				return err
			}
			status := BuildCommandStatus(r)
			if *jsonOut {
				return printJSON(stdout, status)
			}
			fmt.Fprintf(stdout, "mission=%s\nstatus=%s\nread_only=%t\nexecutes_work=%t\nnext=%s\n", status.MissionID, status.Status, status.ReadOnly, status.ExecutesWork, status.ExactNextAction)
			return nil
		}
		return errors.New("command requires status")
	case "artifacts":
		if len(args) >= 2 && args[1] == "manifest" {
			id := missionFlag(args[2:])
			r, err := s.Load(id)
			if err != nil {
				return err
			}
			return printJSON(stdout, BuildArtifactManifest(r))
		}
		id := missionFlag(args[1:])
		r, err := s.Load(id)
		if err != nil {
			return err
		}
		return printJSON(stdout, r.ArtifactRefs)
	case "validate":
		if len(args) >= 2 && args[1] == "contract" {
			fs := flag.NewFlagSet("validate contract", flag.ContinueOnError)
			path := fs.String("path", "", "")
			_ = fs.Parse(args[2:])
			result, err := ValidateContractFile(*path)
			if printErr := printJSON(stdout, result); printErr != nil {
				return printErr
			}
			return err
		}
		return errors.New("validate requires contract --path <file>")
	case "import":
		if len(args) < 2 {
			return errors.New("import requires blueprint-authorization, atlas-workgraph, foundry-run-link, or foundry-final-rollup")
		}
		fs := flag.NewFlagSet("import "+args[1], flag.ContinueOnError)
		id := fs.String("mission", "", "")
		path := fs.String("path", "", "")
		_ = fs.Parse(args[2:])
		rb, err := ImportArtifact(s, *id, args[1], *path)
		if printErr := printJSON(stdout, rb); printErr != nil {
			return printErr
		}
		return err
	case "final":
		if len(args) >= 2 && args[1] == "rollup" {
			id := missionFlag(args[2:])
			r, err := s.Load(id)
			if err != nil {
				return err
			}
			return printJSON(stdout, BuildFinalRollup(r))
		}
		return errors.New("final requires rollup --mission <id>")
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func parseGlobalHome(args []string) (string, []string, error) {
	if len(args) == 0 || args[0] != "--home" {
		return "", args, nil
	}
	if len(args) < 2 || strings.TrimSpace(args[1]) == "" {
		return "", args, errors.New("--home requires a directory")
	}
	return args[1], args[2:], nil
}
func missionFlag(args []string) string {
	fs := flag.NewFlagSet("mission", flag.ContinueOnError)
	id := fs.String("mission", "", "")
	_ = fs.Parse(args)
	return *id
}
