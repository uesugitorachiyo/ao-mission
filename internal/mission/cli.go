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
	"path/filepath"
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
		return errors.New("usage: ao-mission [--home <dir>] <init|start|mission|continue|status|next|stop|pause|resume|doctor|schedule|daemon|telegram|a2a|gateway|governance|command|artifacts|validate|import|final>")
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
			timeline := fs.Bool("timeline", false, "")
			if err := fs.Parse(args[2:]); err != nil {
				return err
			}
			if strings.TrimSpace(*id) == "" {
				return errors.New("mission compact requires --mission")
			}
			if *timeline {
				readback, err := CompactMissionTimeline(s, *id, LedgerCompactionOptions{KeepRouteHistory: *keepRouteHistory, KeepSteps: *keepSteps, DryRun: *dryRun})
				if err != nil {
					return err
				}
				return printJSON(stdout, readback)
			}
			readback, err := CompactMissionLedger(s, *id, LedgerCompactionOptions{KeepRouteHistory: *keepRouteHistory, KeepSteps: *keepSteps, DryRun: *dryRun})
			if err != nil {
				return err
			}
			return printJSON(stdout, readback)
		case "archive":
			fs := flag.NewFlagSet("mission archive", flag.ContinueOnError)
			id := fs.String("mission", "", "")
			outPath := fs.String("out", "", "")
			if err := fs.Parse(args[2:]); err != nil {
				return err
			}
			if strings.TrimSpace(*id) == "" || strings.TrimSpace(*outPath) == "" {
				return errors.New("mission archive requires --mission and --out")
			}
			r, err := s.Load(*id)
			if err != nil {
				return err
			}
			archive, err := BuildMissionArchive(r)
			if err != nil {
				return err
			}
			body, err := json.MarshalIndent(archive, "", "  ")
			if err != nil {
				return err
			}
			if err := os.WriteFile(*outPath, append(body, '\n'), 0o644); err != nil {
				return err
			}
			fmt.Fprintf(stdout, "mission_archive=%s\nmission=%s\nsafe_to_execute=false\nexecutes_work=false\napproves_work=false\n", *outPath, archive.MissionID)
			return nil
		case "validate-archive":
			fs := flag.NewFlagSet("mission validate-archive", flag.ContinueOnError)
			path := fs.String("path", "", "")
			outPath := fs.String("out", "", "")
			if err := fs.Parse(args[2:]); err != nil {
				return err
			}
			if strings.TrimSpace(*path) == "" {
				return errors.New("mission validate-archive requires --path")
			}
			validation, err := ValidateMissionArchive(*path)
			if err != nil {
				return err
			}
			if strings.TrimSpace(*outPath) == "" {
				return printJSON(stdout, validation)
			}
			body, err := json.MarshalIndent(validation, "", "  ")
			if err != nil {
				return err
			}
			if err := os.WriteFile(*outPath, append(body, '\n'), 0o644); err != nil {
				return err
			}
			fmt.Fprintf(stdout, "mission_archive_validation=%s\nmission=%s\nsafe_to_execute=false\nexecutes_work=false\napproves_work=false\n", *outPath, validation.MissionID)
			return nil
		case "import-archive":
			fs := flag.NewFlagSet("mission import-archive", flag.ContinueOnError)
			path := fs.String("path", "", "")
			if err := fs.Parse(args[2:]); err != nil {
				return err
			}
			if strings.TrimSpace(*path) == "" {
				return errors.New("mission import-archive requires --path")
			}
			readback, err := ImportMissionArchive(s, *path)
			if err != nil {
				return err
			}
			return printJSON(stdout, readback)
		case "events":
			if len(args) < 3 {
				return errors.New("mission events requires index or search")
			}
			switch args[2] {
			case "index":
				fs := flag.NewFlagSet("mission events index", flag.ContinueOnError)
				outPath := fs.String("out", "", "")
				if err := fs.Parse(args[3:]); err != nil {
					return err
				}
				index, err := BuildMissionEventIndex(s)
				if err != nil {
					return err
				}
				if strings.TrimSpace(*outPath) != "" {
					body, err := json.MarshalIndent(index, "", "  ")
					if err != nil {
						return err
					}
					if err := os.WriteFile(*outPath, append(body, '\n'), 0o644); err != nil {
						return err
					}
				}
				return printJSON(stdout, index)
			case "search":
				fs := flag.NewFlagSet("mission events search", flag.ContinueOnError)
				missionID := fs.String("mission", "", "")
				kind := fs.String("kind", "", "")
				query := fs.String("query", "", "")
				indexPath := fs.String("index", "", "")
				outPath := fs.String("out", "", "")
				jsonOut := fs.Bool("json", false, "")
				if err := fs.Parse(args[3:]); err != nil {
					return err
				}
				var index MissionEventIndex
				if strings.TrimSpace(*indexPath) != "" {
					body, err := os.ReadFile(*indexPath)
					if err != nil {
						return err
					}
					if err := json.Unmarshal(body, &index); err != nil {
						return err
					}
					if err := ValidateMissionEventIndexDigest(index); err != nil {
						return err
					}
				} else {
					var err error
					index, err = BuildMissionEventIndex(s)
					if err != nil {
						return err
					}
				}
				readback := SearchMissionEvents(index, MissionEventSearchFilters{MissionID: *missionID, Kind: *kind, Query: *query})
				if strings.TrimSpace(*outPath) != "" {
					body, err := json.MarshalIndent(readback, "", "  ")
					if err != nil {
						return err
					}
					if err := os.WriteFile(*outPath, append(body, '\n'), 0o644); err != nil {
						return err
					}
				}
				if *jsonOut {
					return printJSON(stdout, readback)
				}
				fmt.Fprintf(stdout, "mission_events=%d status=%s safe_to_execute=false executes_work=false approves_work=false\n", readback.TotalMatches, readback.Status)
				for _, event := range readback.Events {
					fmt.Fprintf(stdout, "mission=%s kind=%s route=%s summary=%s\n", event.MissionID, event.Kind, event.Route, event.Summary)
				}
				return nil
			default:
				return errors.New("mission events requires index or search")
			}
		case "readiness-bundle":
			fs := flag.NewFlagSet("mission readiness-bundle", flag.ContinueOnError)
			var repos readinessRepoFlags
			outPath := fs.String("out", "", "")
			jsonOut := fs.Bool("json", false, "")
			fs.Var(&repos, "repo", "repo=path readiness summary input")
			if err := fs.Parse(args[2:]); err != nil {
				return err
			}
			inputs, err := repos.inputs()
			if err != nil {
				return err
			}
			readback, err := BuildMissionReadinessBundleReadback(inputs)
			if err != nil {
				return err
			}
			if strings.TrimSpace(*outPath) != "" {
				body, err := json.MarshalIndent(readback, "", "  ")
				if err != nil {
					return err
				}
				if err := os.WriteFile(*outPath, append(body, '\n'), 0o644); err != nil {
					return err
				}
			}
			if *jsonOut || strings.TrimSpace(*outPath) == "" {
				return printJSON(stdout, readback)
			}
			fmt.Fprintf(stdout, "mission_readiness_bundle=%s\nstatus=%s\nready_repos=%d\nsafe_to_execute=false\nexecutes_work=false\napproves_work=false\n", *outPath, readback.Status, readback.ReadyRepos)
			return nil
		case "dashboard":
			fs := flag.NewFlagSet("mission dashboard", flag.ContinueOnError)
			id := fs.String("mission", "", "")
			compact := fs.Bool("compact", false, "")
			jsonOut := fs.Bool("json", false, "")
			outPath := fs.String("out", "", "")
			if err := fs.Parse(args[2:]); err != nil {
				return err
			}
			if strings.TrimSpace(*id) == "" {
				return errors.New("mission dashboard requires --mission")
			}
			readback, err := BuildMissionDashboardReadback(s, *id, *compact)
			if err != nil {
				return err
			}
			if strings.TrimSpace(*outPath) != "" {
				body, err := json.MarshalIndent(readback, "", "  ")
				if err != nil {
					return err
				}
				if err := os.WriteFile(*outPath, append(body, '\n'), 0o644); err != nil {
					return err
				}
			}
			if *jsonOut || strings.TrimSpace(*outPath) == "" {
				return printJSON(stdout, readback)
			}
			fmt.Fprintf(stdout, "mission_dashboard=%s\nmission=%s\nstatus=%s\nlatest_route=%s\nsafe_to_execute=false\nexecutes_work=false\napproves_work=false\n", *outPath, readback.MissionID, readback.Status, readback.LatestRoute)
			return nil
		case "verification-bundle":
			fs := flag.NewFlagSet("mission verification-bundle", flag.ContinueOnError)
			id := fs.String("mission", "", "")
			readinessBundlePath := fs.String("readiness-bundle", "", "")
			gatewayReplayBundlePath := fs.String("gateway-replay-bundle", "", "")
			outPath := fs.String("out", "", "")
			jsonOut := fs.Bool("json", false, "")
			if err := fs.Parse(args[2:]); err != nil {
				return err
			}
			if strings.TrimSpace(*id) == "" {
				return errors.New("mission verification-bundle requires --mission")
			}
			readback, err := BuildMissionVerificationBundleReadback(s, *id, MissionVerificationBundleOptions{
				ReadinessBundlePath:     *readinessBundlePath,
				GatewayReplayBundlePath: *gatewayReplayBundlePath,
			})
			if err != nil {
				return err
			}
			if strings.TrimSpace(*outPath) != "" {
				body, err := json.MarshalIndent(readback, "", "  ")
				if err != nil {
					return err
				}
				if err := os.WriteFile(*outPath, append(body, '\n'), 0o644); err != nil {
					return err
				}
			}
			if *jsonOut || strings.TrimSpace(*outPath) == "" {
				return printJSON(stdout, readback)
			}
			fmt.Fprintf(stdout, "mission_verification_bundle=%s\nmission=%s\nstatus=%s\ncomponent_count=%d\nsafe_to_execute=false\nexecutes_work=false\napproves_work=false\n", *outPath, readback.MissionID, readback.Status, readback.ComponentCount)
			return nil
		default:
			return errors.New("mission requires list, inspect, history, compact, archive, validate-archive, import-archive, events, readiness-bundle, dashboard, or verification-bundle")
		}
	case "doctor":
		fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
		jsonOut := fs.Bool("json", false, "")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		readback := BuildMissionDoctorReadback(s)
		if *jsonOut {
			return printJSON(stdout, readback)
		}
		fmt.Fprintf(stdout, "status=%s\nmissions=%d\nevents=%d\nsafe_to_execute=false\nexecutes_work=false\napproves_work=false\n", readback.Status, readback.MissionCount, readback.EventCount)
		return nil
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
		minNodes := fs.Int("min-nodes", 0, "")
		minMinutes := fs.Int("min-minutes", 0, "")
		maxMinutes := fs.Int("max-minutes", 0, "")
		returnOnlyWhen := fs.String("return-only-when", "", "")
		checkpointPolicy := fs.String("checkpoint-policy", "", "")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		r, err := Continue(s, *id, ContinueOptions{
			UntilDone:        *until,
			MaxIterations:    *max,
			MinNodes:         *minNodes,
			MinMinutes:       *minMinutes,
			MaxMinutes:       *maxMinutes,
			ReturnOnlyWhen:   *returnOnlyWhen,
			CheckpointPolicy: *checkpointPolicy,
		})
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
		if len(args) >= 2 && args[1] == "role-matrix" {
			fs := flag.NewFlagSet("telegram role-matrix", flag.ContinueOnError)
			configPath := fs.String("config", "", "")
			outPath := fs.String("out", "", "")
			if err := fs.Parse(args[2:]); err != nil {
				return err
			}
			if strings.TrimSpace(*configPath) == "" {
				return errors.New("telegram role-matrix requires --config")
			}
			cfg, err := LoadTelegramConfig(*configPath)
			if err != nil {
				return err
			}
			matrix := BuildTelegramRoleMatrix(cfg)
			if strings.TrimSpace(*outPath) == "" {
				return printJSON(stdout, matrix)
			}
			body, err := json.MarshalIndent(matrix, "", "  ")
			if err != nil {
				return err
			}
			if err := os.WriteFile(*outPath, append(body, '\n'), 0o644); err != nil {
				return err
			}
			fmt.Fprintf(stdout, "telegram_role_matrix=%s\nstatus=%s\nsafe_to_execute=false\nexecutes_work=false\napproves_work=false\n", *outPath, matrix.Status)
			return nil
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
		return errors.New("telegram requires serve, replay, replay-updates, webhook-replay, or role-matrix")
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
		if len(args) >= 2 && args[1] == "compatibility" {
			fs := flag.NewFlagSet("a2a compatibility", flag.ContinueOnError)
			agentCardPath := fs.String("agent-card", "", "")
			httpPath := fs.String("http", "", "")
			lifecyclePath := fs.String("lifecycle", "", "")
			outPath := fs.String("out", "", "")
			if err := fs.Parse(args[2:]); err != nil {
				return err
			}
			if strings.TrimSpace(*agentCardPath) == "" || strings.TrimSpace(*httpPath) == "" || strings.TrimSpace(*lifecyclePath) == "" {
				return errors.New("a2a compatibility requires --agent-card, --http, and --lifecycle")
			}
			readback, err := BuildA2ACompatibilityReadback(*agentCardPath, *httpPath, *lifecyclePath)
			if err != nil {
				return err
			}
			if strings.TrimSpace(*outPath) == "" {
				return printJSON(stdout, readback)
			}
			body, err := json.MarshalIndent(readback, "", "  ")
			if err != nil {
				return err
			}
			if err := os.WriteFile(*outPath, append(body, '\n'), 0o644); err != nil {
				return err
			}
			fmt.Fprintf(stdout, "a2a_compatibility=%s\nstatus=%s\nsafe_to_execute=false\nexecutes_work=false\napproves_work=false\n", *outPath, readback.Status)
			return nil
		}
		if len(args) >= 2 && args[1] == "streaming-denial" {
			fs := flag.NewFlagSet("a2a streaming-denial", flag.ContinueOnError)
			agentCardPath := fs.String("agent-card", "", "")
			outPath := fs.String("out", "", "")
			if err := fs.Parse(args[2:]); err != nil {
				return err
			}
			if strings.TrimSpace(*agentCardPath) == "" {
				return errors.New("a2a streaming-denial requires --agent-card")
			}
			readback, err := BuildA2AStreamingDenialReadback(*agentCardPath)
			if err != nil {
				return err
			}
			if strings.TrimSpace(*outPath) == "" {
				return printJSON(stdout, readback)
			}
			body, err := json.MarshalIndent(readback, "", "  ")
			if err != nil {
				return err
			}
			if err := os.WriteFile(*outPath, append(body, '\n'), 0o644); err != nil {
				return err
			}
			fmt.Fprintf(stdout, "a2a_streaming_denial=%s\nstatus=%s\nsafe_to_execute=false\nexecutes_work=false\napproves_work=false\n", *outPath, readback.Status)
			return nil
		}
		if len(args) >= 2 && args[1] == "cancellation-replay" {
			fs := flag.NewFlagSet("a2a cancellation-replay", flag.ContinueOnError)
			lifecyclePath := fs.String("lifecycle", "", "")
			outPath := fs.String("out", "", "")
			if err := fs.Parse(args[2:]); err != nil {
				return err
			}
			if strings.TrimSpace(*lifecyclePath) == "" {
				return errors.New("a2a cancellation-replay requires --lifecycle")
			}
			readback, err := BuildA2ACancellationReplayReadback(*lifecyclePath)
			if err != nil {
				return err
			}
			if strings.TrimSpace(*outPath) == "" {
				return printJSON(stdout, readback)
			}
			body, err := json.MarshalIndent(readback, "", "  ")
			if err != nil {
				return err
			}
			if err := os.WriteFile(*outPath, append(body, '\n'), 0o644); err != nil {
				return err
			}
			fmt.Fprintf(stdout, "a2a_cancellation_replay=%s\nstatus=%s\nsafe_to_execute=false\nexecutes_work=false\napproves_work=false\n", *outPath, readback.Status)
			return nil
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
				return printJSON(stdout, map[string]any{
					"schema":             "ao.mission.a2a-fixture-server-readback.v0.1",
					"gateway":            "a2a",
					"status":             "ready",
					"listen":             addr,
					"agent_card_path":    "/.well-known/agent-card.json",
					"jsonrpc_path":       "/",
					"methods":            AgentCard().Methods,
					"message":            "A2A local HTTP fixture server can bind and records intents only",
					"mutation_authority": false,
					"executes_work":      false,
					"approves_work":      false,
					"generated_at_utc":   now(nil),
				})
			}
			fmt.Fprintf(stdout, "a2a_listen=%s\nmutation_authority=false\n", ln.Addr().String())
			return server.Serve(ln)
		}
		return errors.New("a2a requires serve, replay, lifecycle, compatibility, streaming-denial, or cancellation-replay")
	case "gateway":
		if len(args) >= 2 && args[1] == "replay-bundle" {
			fs := flag.NewFlagSet("gateway replay-bundle", flag.ContinueOnError)
			telegramConfigPath := fs.String("telegram-config", "", "")
			telegramMatrixPath := fs.String("telegram-matrix", "", "")
			telegramUpdatesPath := fs.String("telegram-updates", "", "")
			telegramWebhookPath := fs.String("telegram-webhook", "", "")
			a2aHTTPPath := fs.String("a2a-http", "", "")
			a2aLifecyclePath := fs.String("a2a-lifecycle", "", "")
			schedulerPath := fs.String("scheduler", "", "")
			outPath := fs.String("out", "", "")
			jsonOut := fs.Bool("json", false, "")
			if err := fs.Parse(args[2:]); err != nil {
				return err
			}
			readback, err := BuildGatewayReplayBundleReadback(GatewayReplayBundleInputs{
				TelegramConfigPath:  *telegramConfigPath,
				TelegramMatrixPath:  *telegramMatrixPath,
				TelegramUpdatesPath: *telegramUpdatesPath,
				TelegramWebhookPath: *telegramWebhookPath,
				A2AHTTPPath:         *a2aHTTPPath,
				A2ALifecyclePath:    *a2aLifecyclePath,
				SchedulerPath:       *schedulerPath,
			})
			if err != nil {
				return err
			}
			if strings.TrimSpace(*outPath) != "" {
				body, err := json.MarshalIndent(readback, "", "  ")
				if err != nil {
					return err
				}
				if err := os.WriteFile(*outPath, append(body, '\n'), 0o644); err != nil {
					return err
				}
			}
			if *jsonOut || strings.TrimSpace(*outPath) == "" {
				return printJSON(stdout, readback)
			}
			fmt.Fprintf(stdout, "gateway_replay_bundle=%s\nstatus=%s\nsafe_to_execute=false\nexecutes_work=false\napproves_work=false\n", *outPath, readback.Status)
			return nil
		}
		if len(args) >= 2 && args[1] == "replay-suite" {
			fs := flag.NewFlagSet("gateway replay-suite", flag.ContinueOnError)
			telegramConfigPath := fs.String("telegram-config", "", "")
			telegramWebhookPath := fs.String("telegram-webhook", "", "")
			telegramUpdatesPath := fs.String("telegram-updates", "", "")
			a2aHTTPPath := fs.String("a2a-http", "", "")
			a2aLifecyclePath := fs.String("a2a-lifecycle", "", "")
			outPath := fs.String("out", "", "")
			if err := fs.Parse(args[2:]); err != nil {
				return err
			}
			if strings.TrimSpace(*outPath) == "" {
				return errors.New("gateway replay-suite requires --out")
			}
			readbacks := []GatewayReplayReadback{}
			refs := []string{}
			var lifecycle *A2ATaskLifecycleReadback
			var allowedChats map[string]string
			if strings.TrimSpace(*telegramWebhookPath) != "" || strings.TrimSpace(*telegramUpdatesPath) != "" {
				if strings.TrimSpace(*telegramConfigPath) == "" {
					return errors.New("gateway replay-suite requires --telegram-config with Telegram fixtures")
				}
				cfg, err := LoadTelegramConfig(*telegramConfigPath)
				if err != nil {
					return err
				}
				allowedChats = cfg.AllowedChats
			}
			if strings.TrimSpace(*telegramWebhookPath) != "" {
				readback, err := ReplayTelegramWebhookFixture(*telegramWebhookPath, allowedChats)
				if err != nil {
					return err
				}
				readbacks = append(readbacks, readback)
				refs = append(refs, filepath.ToSlash(*telegramWebhookPath))
			}
			if strings.TrimSpace(*telegramUpdatesPath) != "" {
				readback, err := ReplayTelegramUpdates(*telegramUpdatesPath, allowedChats)
				if err != nil {
					return err
				}
				readbacks = append(readbacks, readback)
				refs = append(refs, filepath.ToSlash(*telegramUpdatesPath))
			}
			if strings.TrimSpace(*a2aHTTPPath) != "" {
				readback, err := ReplayA2AHTTPFixture(*a2aHTTPPath)
				if err != nil {
					return err
				}
				readbacks = append(readbacks, readback)
				refs = append(refs, filepath.ToSlash(*a2aHTTPPath))
			}
			if strings.TrimSpace(*a2aLifecyclePath) != "" {
				readback, err := ReplayA2ATaskLifecycle(*a2aLifecyclePath)
				if err != nil {
					return err
				}
				lifecycle = &readback
				refs = append(refs, filepath.ToSlash(*a2aLifecyclePath))
			}
			if len(readbacks) == 0 && lifecycle == nil {
				return errors.New("gateway replay-suite requires at least one replay input")
			}
			suite := BuildGatewayReplaySuite(readbacks, lifecycle, refs)
			body, err := json.MarshalIndent(suite, "", "  ")
			if err != nil {
				return err
			}
			if err := os.WriteFile(*outPath, append(body, '\n'), 0o644); err != nil {
				return err
			}
			fmt.Fprintf(stdout, "gateway_replay_suite=%s\nstatus=%s\nsafe_to_execute=false\nexecutes_work=false\napproves_work=false\n", *outPath, suite.Status)
			return nil
		}
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
		if len(args) >= 2 && args[1] == "readiness-rollup" {
			fs := flag.NewFlagSet("gateway readiness-rollup", flag.ContinueOnError)
			missionID := fs.String("mission", "", "")
			suitePath := fs.String("suite", "", "")
			a2aCompatibilityPath := fs.String("a2a-compatibility", "", "")
			archiveValidationPath := fs.String("archive-validation", "", "")
			snapshotDiffPath := fs.String("snapshot-diff", "", "")
			correlationID := fs.String("correlation-id", "", "")
			outPath := fs.String("out", "", "")
			if err := fs.Parse(args[2:]); err != nil {
				return err
			}
			if strings.TrimSpace(*outPath) == "" {
				return errors.New("gateway readiness-rollup requires --out")
			}
			if strings.TrimSpace(*missionID) == "" {
				return errors.New("gateway readiness-rollup requires --mission")
			}
			rollup, err := BuildGatewayReadinessRollupWithMissionAndCorrelation(*missionID, *correlationID, *suitePath, *a2aCompatibilityPath, *archiveValidationPath, *snapshotDiffPath)
			if err != nil {
				return err
			}
			body, err := json.MarshalIndent(rollup, "", "  ")
			if err != nil {
				return err
			}
			if err := os.WriteFile(*outPath, append(body, '\n'), 0o644); err != nil {
				return err
			}
			fmt.Fprintf(stdout, "gateway_readiness_rollup=%s\nstatus=%s\nsafe_to_execute=false\nexecutes_work=false\napproves_work=false\n", *outPath, rollup.Status)
			return nil
		}
		return errors.New("gateway requires ledger, replay-suite, replay-bundle, or readiness-rollup")
	case "governance":
		if len(args) >= 2 && args[1] == "snapshot" {
			id := missionFlag(args[2:])
			r, err := s.Load(id)
			if err != nil {
				return err
			}
			return printJSON(stdout, Snapshot(r))
		}
		if len(args) >= 2 && args[1] == "diff" {
			fs := flag.NewFlagSet("governance diff", flag.ContinueOnError)
			beforePath := fs.String("before", "", "")
			afterPath := fs.String("after", "", "")
			outPath := fs.String("out", "", "")
			if err := fs.Parse(args[2:]); err != nil {
				return err
			}
			if strings.TrimSpace(*beforePath) == "" || strings.TrimSpace(*afterPath) == "" {
				return errors.New("governance diff requires --before and --after")
			}
			before, err := LoadGovernanceSnapshot(*beforePath)
			if err != nil {
				return err
			}
			after, err := LoadGovernanceSnapshot(*afterPath)
			if err != nil {
				return err
			}
			diff := DiffGovernanceSnapshots(before, after)
			if strings.TrimSpace(*outPath) == "" {
				return printJSON(stdout, diff)
			}
			body, err := json.MarshalIndent(diff, "", "  ")
			if err != nil {
				return err
			}
			if err := os.WriteFile(*outPath, append(body, '\n'), 0o644); err != nil {
				return err
			}
			fmt.Fprintf(stdout, "governance_snapshot_diff=%s\nmission=%s\nsafe_to_execute=false\nexecutes_work=false\napproves_work=false\n", *outPath, diff.MissionID)
			return nil
		}
		return errors.New("governance requires snapshot or diff")
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
			if status.AtlasRecommendation != nil {
				fmt.Fprintf(stdout, "atlas_recommendation=%s completed_nodes=%d total_nodes=%d ready_nodes=%d final_response_allowed=%t\n", status.AtlasRecommendation.Status, status.AtlasRecommendation.CompletedNodes, status.AtlasRecommendation.TotalNodes, status.AtlasRecommendation.ReadyNodes, status.AtlasRecommendation.FinalResponseAllowed)
			}
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
			return errors.New("import requires blueprint-authorization, atlas-workgraph, atlas-recommendation-readback, atlas-final-synthesis-readback, foundry-run-link, foundry-final-rollup, scheduler-readback, scheduler-recovery-readback, or ledger-compaction-readback")
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
		if len(args) >= 2 && args[1] == "reconcile" {
			id := missionFlag(args[2:])
			r, err := s.Load(id)
			if err != nil {
				return err
			}
			return printJSON(stdout, BuildFinalReconciliationPacket(r))
		}
		if len(args) >= 2 && args[1] == "synthesize" {
			fs := flag.NewFlagSet("final synthesize", flag.ContinueOnError)
			id := fs.String("mission", "", "")
			evidenceRoot := fs.String("evidence-root", "", "")
			if err := fs.Parse(args[2:]); err != nil {
				return err
			}
			r, err := s.Load(*id)
			if err != nil {
				return err
			}
			synthesis, err := BuildAtlasWaveFinalSynthesis(r, *evidenceRoot)
			if err != nil {
				return err
			}
			return printJSON(stdout, synthesis)
		}
		return errors.New("final requires rollup --mission <id>, reconcile --mission <id>, or synthesize --mission <id> --evidence-root <path>")
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

type readinessRepoFlags []string

func (f *readinessRepoFlags) String() string {
	return strings.Join(*f, ",")
}

func (f *readinessRepoFlags) Set(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return errors.New("--repo requires repo=path")
	}
	*f = append(*f, value)
	return nil
}

func (f readinessRepoFlags) inputs() ([]MissionReadinessBundleInput, error) {
	inputs := []MissionReadinessBundleInput{}
	for _, value := range f {
		repo, path, ok := strings.Cut(value, "=")
		if !ok || strings.TrimSpace(repo) == "" || strings.TrimSpace(path) == "" {
			return nil, fmt.Errorf("--repo must be repo=path")
		}
		inputs = append(inputs, MissionReadinessBundleInput{Repo: strings.TrimSpace(repo), Path: strings.TrimSpace(path)})
	}
	return inputs, nil
}
