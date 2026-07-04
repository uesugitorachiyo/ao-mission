package mission

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

type Store struct {
	Root  string
	Clock func() time.Time
}

func DefaultRoot() string {
	if v := os.Getenv("AO_MISSION_HOME"); strings.TrimSpace(v) != "" {
		return v
	}
	return ".ao-mission"
}
func NewStore(root string) Store {
	if root == "" {
		root = DefaultRoot()
	}
	return Store{Root: root}
}
func (s Store) Init() error           { return os.MkdirAll(filepath.Join(s.Root, "missions"), 0o755) }
func (s Store) path(id string) string { return filepath.Join(s.Root, "missions", id+".json") }
func (s Store) eventLoopPath(id string) string {
	return filepath.Join(s.Root, "missions", id+".event-loop-decision.json")
}
func (s Store) checkpointPath(id string) string {
	return filepath.Join(s.Root, "missions", id+".checkpoint-resume-bundle.json")
}

func DigestObjective(objective string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(objective)))
	return "sha256:" + hex.EncodeToString(sum[:])
}
func MissionID(objective string, t time.Time) string {
	sum := sha256.Sum256([]byte(t.UTC().Format(time.RFC3339Nano) + "\x00" + objective))
	return "mission-" + hex.EncodeToString(sum[:])[:16]
}

func (s Store) Start(objective string) (Record, error) {
	if strings.TrimSpace(objective) == "" {
		return Record{}, errors.New("objective is required")
	}
	if err := s.Init(); err != nil {
		return Record{}, err
	}
	t := time.Now
	if s.Clock != nil {
		t = s.Clock
	}
	stamp := now(t)
	id := MissionID(objective, t())
	route := DecideRoute(id, objective, nil)
	rec := Record{Schema: RecordSchema, MissionID: id, Objective: objective, ObjectiveDigest: DigestObjective(objective), Status: "active", CreatedAtUTC: stamp, UpdatedAtUTC: stamp, CurrentRoute: route.Route, CurrentPhase: "routing", ExactNextAction: route.ExactNextAction, ArtifactRefs: []ArtifactRef{}, Blockers: []string{}, Steps: []ContinuationStep{}}
	AppendRouteHistory(&rec, route)
	return rec, s.Save(rec)
}
func (s Store) Load(id string) (Record, error) {
	var r Record
	b, err := os.ReadFile(s.path(id))
	if err != nil {
		return r, err
	}
	return r, json.Unmarshal(b, &r)
}

type ListFilters struct {
	Status string
	Route  string
}

func (s Store) List() ([]Record, error) {
	return s.ListFiltered(ListFilters{})
}

func (s Store) ListFiltered(filters ListFilters) ([]Record, error) {
	if err := s.Init(); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(filepath.Join(s.Root, "missions"))
	if err != nil {
		return nil, err
	}
	records := make([]Record, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		var rec Record
		body, err := os.ReadFile(filepath.Join(s.Root, "missions", entry.Name()))
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(body, &rec); err != nil {
			return nil, err
		}
		if rec.Schema != RecordSchema || rec.MissionID == "" {
			continue
		}
		if filters.Status != "" && rec.Status != filters.Status {
			continue
		}
		if filters.Route != "" && rec.CurrentRoute != filters.Route {
			continue
		}
		records = append(records, rec)
	}
	sort.Slice(records, func(i, j int) bool {
		if records[i].CreatedAtUTC == records[j].CreatedAtUTC {
			return records[i].MissionID < records[j].MissionID
		}
		return records[i].CreatedAtUTC < records[j].CreatedAtUTC
	})
	return records, nil
}
func (s Store) Save(r Record) error {
	if err := s.Init(); err != nil {
		return err
	}
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(s.path(r.MissionID), b, 0o644)
}

func (s Store) SaveEventLoopDecision(decision EventLoopDecision) error {
	if err := s.Init(); err != nil {
		return err
	}
	b, err := json.MarshalIndent(decision, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(s.eventLoopPath(decision.MissionID), b, 0o644)
}

func (s Store) LoadEventLoopDecision(id string) (EventLoopDecision, error) {
	var decision EventLoopDecision
	b, err := os.ReadFile(s.eventLoopPath(id))
	if err != nil {
		return decision, err
	}
	return decision, json.Unmarshal(b, &decision)
}

func (s Store) SaveCheckpointBundle(bundle MissionCheckpointBundle) error {
	if err := s.Init(); err != nil {
		return err
	}
	b, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(s.checkpointPath(bundle.MissionID), b, 0o644)
}

func (s Store) LoadCheckpointBundle(id string) (MissionCheckpointBundle, error) {
	var bundle MissionCheckpointBundle
	b, err := os.ReadFile(s.checkpointPath(id))
	if err != nil {
		return bundle, err
	}
	return bundle, json.Unmarshal(b, &bundle)
}
func (s Store) Update(id string, fn func(*Record) error) (Record, error) {
	r, err := s.Load(id)
	if err != nil {
		return r, err
	}
	if err := fn(&r); err != nil {
		return r, err
	}
	r.UpdatedAtUTC = now(s.Clock)
	return r, s.Save(r)
}

func ValidatePublicSafeText(text string) error {
	localPath := `/` + `Users/`
	privateKey := `BEGIN (RSA |OPENSSH |PRIVATE )?PRIVATE ` + `KEY`
	openAIKey := `sk-` + `[A-Za-z0-9]{20,}`
	gitHubToken := `gh[pousr]_` + `[A-Za-z0-9]{20,}`
	secretAssignment := `(?i)(password|` + `token|cookie)\s*[:=]\s*[^\s]+`
	patterns := []*regexp.Regexp{regexp.MustCompile(localPath), regexp.MustCompile(privateKey), regexp.MustCompile(openAIKey), regexp.MustCompile(gitHubToken), regexp.MustCompile(secretAssignment)}
	for _, p := range patterns {
		if p.MatchString(text) {
			return fmt.Errorf("unsafe public content detected")
		}
	}
	return nil
}
