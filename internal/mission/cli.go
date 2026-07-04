package mission

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
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
		return errors.New("usage: ao-mission [--home <dir>] <init|start|mission|continue|status|next|stop|pause|resume|schedule|daemon|telegram|a2a|gateway|governance|command|artifacts|validate|import|final>")
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
			statusFilter := fs.String("status", "", "")
			routeFilter := fs.String("route", "", "")
			if err := fs.Parse(args[2:]); err != nil {
				return err
			}
			records, err := s.ListFiltered(ListFilters{Status: *statusFilter, Route: *routeFilter})
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
			if err := fs.Parse(args[2:]); err != nil {
				return err
			}
			r, err := s.Load(*id)
			if err != nil {
				return err
			}
			if *jsonOut {
				return printJSON(stdout, r)
			}
			fmt.Fprintf(stdout, "mission=%s\nstatus=%s\nphase=%s\nroute=%s\nnext=%s\n", r.MissionID, r.Status, r.CurrentPhase, r.CurrentRoute, r.ExactNextAction)
			return nil
		case "history":
			fs := flag.NewFlagSet("mission history", flag.ContinueOnError)
			id := fs.String("mission", "", "")
			jsonOut := fs.Bool("json", false, "")
			if err := fs.Parse(args[2:]); err != nil {
				return err
			}
			r, err := s.Load(*id)
			if err != nil {
				return err
			}
			if *jsonOut {
				return printJSON(stdout, r.RouteHistory)
			}
			for _, item := range r.RouteHistory {
				fmt.Fprintf(stdout, "route=%s reason=%s safe_to_execute=%t next=%s\n", item.Route, item.Reason, item.SafeToExecute, item.ExactNextAction)
			}
			return nil
		case "compact":
			fs := flag.NewFlagSet("mission compact", flag.ContinueOnError)
			id := fs.String("mission", "", "")
			keepRouteHistory := fs.Int("keep-route-history", 25, "")
			keepSteps := fs.Int("keep-steps", 25, "")
			dryRun := fs.Bool("dry-run", false, "")
			if err := fs.Parse(args[2:]); err != nil {
				return err
			}
			if strings.TrimSpace(*id) == "" {
				return errors.New("mission compact requires --mission")
			}
			readback, err := CompactMissionLedger(s, *id, LedgerCompactionOptions{KeepRouteHistory: *keepRouteHistory, KeepSteps: *keepSteps, DryRun: *dryRun})
			if err != nil {
				return err
			}
			return printJSON(stdout, readback)
		default:
			return errors.New("mission requires list, inspect, history, or compact")
		}
	case "status":
		fs := flag.NewFlagSet("status", flag.ContinueOnError)
		id := fs.String("mission", "", "")
		jsonOut := fs.Bool("json", false, "")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
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
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		var d RouteDecision
		_, err := s.Update(*id, func(r *Record) error {
			d = NextAction(*r)
			AppendRouteHistory(r, d)
			return nil
		})
		if err != nil {
			return err
		}
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
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
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
		if len(args) >= 2 && args[1] == "replay" {
			fs := flag.NewFlagSet("schedule replay", flag.ContinueOnError)
			fixturePath := fs.String("fixture", "", "")
			if err := fs.Parse(args[2:]); err != nil {
				return err
			}
			if *fixturePath == "" {
				return errors.New("schedule replay requires --fixture")
			}
			readback, err := ReplaySchedulerReadbacks(*fixturePath)
			if err != nil {
				return err
			}
			return printJSON(stdout, readback)
		}
		if len(args) >= 2 && args[1] == "alerts" {
			fs := flag.NewFlagSet("schedule alerts", flag.ContinueOnError)
			fixturePath := fs.String("fixture", "", "")
			if err := fs.Parse(args[2:]); err != nil {
				return err
			}
			if *fixturePath == "" {
				return errors.New("schedule alerts requires --fixture")
			}
			readback, err := ReplaySchedulerReadbacks(*fixturePath)
			if err != nil {
				return err
			}
			return printJSON(stdout, BuildSchedulerAlertSummary(readback))
		}
		if len(args) >= 2 && args[1] == "recover" {
			fs := flag.NewFlagSet("schedule recover", flag.ContinueOnError)
			id := fs.String("mission", "", "")
			fixturePath := fs.String("fixture", "", "")
			if err := fs.Parse(args[2:]); err != nil {
				return err
			}
			if strings.TrimSpace(*id) == "" || strings.TrimSpace(*fixturePath) == "" {
				return errors.New("schedule recover requires --mission and --fixture")
			}
			readback, err := ReplaySchedulerReadbacks(*fixturePath)
			if err != nil {
				return err
			}
			return printJSON(stdout, BuildSchedulerRecoveryReadback(*id, readback))
		}
		fs := flag.NewFlagSet("schedule", flag.ContinueOnError)
		id := fs.String("mission", "", "")
		every := fs.String("every", "", "")
		eventLoop := fs.Bool("event-loop", false, "")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		_ = every
		return printJSON(stdout, ScheduleReadback(*id, *every, *eventLoop))
	case "daemon":
		if len(args) < 2 {
			return errors.New("daemon requires install/status/uninstall")
		}
		fmt.Fprintf(stdout, "daemon=%s\nstatus=readback_only\n", args[1])
		return nil
	case "telegram":
		if len(args) >= 2 && args[1] == "webhook-replay" {
			fs := flag.NewFlagSet("telegram webhook-replay", flag.ContinueOnError)
			fixturePath := fs.String("fixture", "", "")
			configPath := fs.String("config", "", "")
			if err := fs.Parse(args[2:]); err != nil {
				return err
			}
			if *fixturePath == "" || *configPath == "" {
				return errors.New("telegram webhook-replay requires --fixture and --config")
			}
			cfg, err := LoadTelegramConfig(*configPath)
			if err != nil {
				return err
			}
			readback, err := ReplayTelegramWebhookFixture(*fixturePath, cfg.AllowedChats)
			if err != nil {
				return err
			}
			return printJSON(stdout, readback)
		}
		if len(args) >= 2 && args[1] == "replay-updates" {
			fs := flag.NewFlagSet("telegram replay-updates", flag.ContinueOnError)
			fixturePath := fs.String("fixture", "", "")
			configPath := fs.String("config", "", "")
			if err := fs.Parse(args[2:]); err != nil {
				return err
			}
			if *fixturePath == "" || *configPath == "" {
				return errors.New("telegram replay-updates requires --fixture and --config")
			}
			cfg, err := LoadTelegramConfig(*configPath)
			if err != nil {
				return err
			}
			readback, err := ReplayTelegramUpdates(*fixturePath, cfg.AllowedChats)
			if err != nil {
				return err
			}
			return printJSON(stdout, readback)
		}
		if len(args) >= 2 && args[1] == "replay" {
			fs := flag.NewFlagSet("telegram replay", flag.ContinueOnError)
			matrixPath := fs.String("matrix", "", "")
			configPath := fs.String("config", "", "")
			if err := fs.Parse(args[2:]); err != nil {
				return err
			}
			if *matrixPath == "" || *configPath == "" {
				return errors.New("telegram replay requires --matrix and --config")
			}
			cfg, err := LoadTelegramConfig(*configPath)
			if err != nil {
				return err
			}
			readback, err := ReplayTelegramCommandMatrix(*matrixPath, cfg.AllowedChats)
			if err != nil {
				return err
			}
			return printJSON(stdout, readback)
		}
		if len(args) >= 2 && args[1] == "serve" {
			fs := flag.NewFlagSet("telegram serve", flag.ContinueOnError)
			configPath := fs.String("config", "", "")
			if err := fs.Parse(args[2:]); err != nil {
				return err
			}
			if *configPath == "" {
				return printJSON(stdout, TelegramReadback{Schema: TelegramReadbackSchema, Status: "disabled", Message: "telegram gateway disabled by default; configure environment token name and allowlist", MutationAuthority: false})
			}
			cfg, err := LoadTelegramConfig(*configPath)
			if err != nil {
				return err
			}
			return printJSON(stdout, TelegramConfigReadback(cfg))
		}
		return errors.New("telegram requires serve, replay, replay-updates, or webhook-replay")
	case "a2a":
		if len(args) >= 2 && args[1] == "replay" {
			fs := flag.NewFlagSet("a2a replay", flag.ContinueOnError)
			fixturePath := fs.String("fixture", "", "")
			if err := fs.Parse(args[2:]); err != nil {
				return err
			}
			if *fixturePath == "" {
				return errors.New("a2a replay requires --fixture")
			}
			readback, err := ReplayA2AHTTPFixture(*fixturePath)
			if err != nil {
				return err
			}
			return printJSON(stdout, readback)
		}
		if len(args) >= 2 && args[1] == "lifecycle" {
			fs := flag.NewFlagSet("a2a lifecycle", flag.ContinueOnError)
			fixturePath := fs.String("fixture", "", "")
			if err := fs.Parse(args[2:]); err != nil {
				return err
			}
			if *fixturePath == "" {
				return errors.New("a2a lifecycle requires --fixture")
			}
			readback, err := ReplayA2ATaskLifecycle(*fixturePath)
			if err != nil {
				return err
			}
			return printJSON(stdout, readback)
		}
		if len(args) >= 2 && args[1] == "serve" {
			fs := flag.NewFlagSet("a2a serve", flag.ContinueOnError)
			httpMode := fs.Bool("http", false, "")
			listen := fs.String("listen", "127.0.0.1:0", "")
			once := fs.Bool("once", false, "")
			if err := fs.Parse(args[2:]); err != nil {
				return err
			}
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
		return errors.New("a2a requires serve, replay, or lifecycle")
	case "gateway":
		if len(args) >= 2 && args[1] == "ledger" {
			fs := flag.NewFlagSet("gateway ledger", flag.ContinueOnError)
			missionID := fs.String("mission", "", "")
			telegramUpdatesPath := fs.String("telegram-updates", "", "")
			telegramConfigPath := fs.String("telegram-config", "", "")
			a2aHTTPPath := fs.String("a2a-http", "", "")
			outPath := fs.String("out", "", "")
			if err := fs.Parse(args[2:]); err != nil {
				return err
			}
			if strings.TrimSpace(*missionID) == "" || strings.TrimSpace(*outPath) == "" {
				return errors.New("gateway ledger requires --mission and --out")
			}
			readbacks := []GatewayReplayReadback{}
			if strings.TrimSpace(*telegramUpdatesPath) != "" {
				if strings.TrimSpace(*telegramConfigPath) == "" {
					return errors.New("gateway ledger requires --telegram-config with --telegram-updates")
				}
				cfg, err := LoadTelegramConfig(*telegramConfigPath)
				if err != nil {
					return err
				}
				readback, err := ReplayTelegramUpdates(*telegramUpdatesPath, cfg.AllowedChats)
				if err != nil {
					return err
				}
				readbacks = append(readbacks, readback)
			}
			if strings.TrimSpace(*a2aHTTPPath) != "" {
				readback, err := ReplayA2AHTTPFixture(*a2aHTTPPath)
				if err != nil {
					return err
				}
				readbacks = append(readbacks, readback)
			}
			if len(readbacks) == 0 {
				return errors.New("gateway ledger requires at least one replay input")
			}
			ledger := BuildGatewayIntentLedger(*missionID, readbacks...)
			body, err := json.MarshalIndent(ledger, "", "  ")
			if err != nil {
				return err
			}
			if err := os.WriteFile(*outPath, append(body, '\n'), 0o644); err != nil {
				return err
			}
			fmt.Fprintf(stdout, "gateway_intent_ledger=%s\nmission=%s\nsafe_to_execute=false\nexecutes_work=false\napproves_work=false\n", *outPath, ledger.MissionID)
			return nil
		}
		return errors.New("gateway requires ledger")
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
			if err := fs.Parse(args[2:]); err != nil {
				return err
			}
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
			fs := flag.NewFlagSet("artifacts manifest", flag.ContinueOnError)
			id := fs.String("mission", "", "")
			outPath := fs.String("out", "", "")
			if err := fs.Parse(args[2:]); err != nil {
				return err
			}
			r, err := s.Load(*id)
			if err != nil {
				return err
			}
			manifest := BuildArtifactManifest(r)
			if *outPath == "" {
				return printJSON(stdout, manifest)
			}
			body, err := json.MarshalIndent(manifest, "", "  ")
			if err != nil {
				return err
			}
			body = append(body, '\n')
			if err := os.WriteFile(*outPath, body, 0o644); err != nil {
				return err
			}
			fmt.Fprintf(stdout, "artifact_manifest=%s\nmission=%s\nsafe_to_execute=false\nexecutes_work=false\napproves_work=false\n", *outPath, manifest.MissionID)
			return nil
		}
		if len(args) >= 2 && args[1] == "validate-manifest" {
			fs := flag.NewFlagSet("artifacts validate-manifest", flag.ContinueOnError)
			path := fs.String("path", "", "")
			if err := fs.Parse(args[2:]); err != nil {
				return err
			}
			if strings.TrimSpace(*path) == "" {
				return errors.New("artifacts validate-manifest requires --path")
			}
			result, err := ValidateArtifactManifestFile(*path)
			if printErr := printJSON(stdout, result); printErr != nil {
				return printErr
			}
			return err
		}
		if len(args) >= 2 && args[1] == "repair-manifest" {
			fs := flag.NewFlagSet("artifacts repair-manifest", flag.ContinueOnError)
			path := fs.String("path", "", "")
			outPath := fs.String("out", "", "")
			if err := fs.Parse(args[2:]); err != nil {
				return err
			}
			if strings.TrimSpace(*path) == "" || strings.TrimSpace(*outPath) == "" {
				return errors.New("artifacts repair-manifest requires --path and --out")
			}
			manifest, err := RepairArtifactManifestFile(*path)
			if err != nil {
				return err
			}
			body, err := json.MarshalIndent(manifest, "", "  ")
			if err != nil {
				return err
			}
			if err := os.WriteFile(*outPath, append(body, '\n'), 0o644); err != nil {
				return err
			}
			fmt.Fprintf(stdout, "artifact_manifest_repaired=%s\nmission=%s\nsafe_to_execute=false\nexecutes_work=false\napproves_work=false\n", *outPath, manifest.MissionID)
			return nil
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
			if err := fs.Parse(args[2:]); err != nil {
				return err
			}
			result, err := ValidateContractFile(*path)
			if printErr := printJSON(stdout, result); printErr != nil {
				return printErr
			}
			return err
		}
		return errors.New("validate requires contract --path <file>")
	case "import":
		if len(args) < 2 {
			return errors.New("import requires blueprint-authorization, atlas-workgraph, foundry-run-link, foundry-final-rollup, scheduler-readback, scheduler-recovery-readback, or ledger-compaction-readback")
		}
		fs := flag.NewFlagSet("import "+args[1], flag.ContinueOnError)
		id := fs.String("mission", "", "")
		path := fs.String("path", "", "")
		if err := fs.Parse(args[2:]); err != nil {
			return err
		}
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
