package commands

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"time"

	"crona/kernel/internal/core"
	storerepositories "crona/kernel/internal/store/repositories"
	sharedtypes "crona/shared/types"
)

type customHabitMomentumSnapshotState struct {
	Version     int                                   `json:"version,omitempty"`
	Definitions []customHabitMomentumSnapshotDefState `json:"definitions,omitempty"`
}

const customHabitMomentumSnapshotStateVersion = 3

type customHabitMomentumSnapshotDefState struct {
	ID          string `json:"id"`
	Current     int    `json:"current"`
	Longest     int    `json:"longest"`
	BucketKey   string `json:"bucketKey,omitempty"`
	BucketCount int    `json:"bucketCount"`
	BucketMet   bool   `json:"bucketMet"`
}

func ComputeCustomHabitStreakSnapshot(
	ctx context.Context,
	c *core.Context,
	throughDate string,
) ([]sharedtypes.CustomHabitStreakSummary, error) {
	snapshot, err := ensureCustomHabitMomentumSnapshot(ctx, c, throughDate, nil, nil)
	if err != nil {
		return nil, err
	}
	if snapshot == nil {
		return nil, nil
	}
	return snapshot.summaries, nil
}

func SeedCustomHabitMomentumSnapshot(ctx context.Context, c *core.Context) error {
	defs, err := c.HabitStreakDefinitions.List(ctx, c.UserID)
	if err != nil {
		return err
	}
	if len(defs) == 0 {
		return nil
	}
	empty, err := c.CustomHabitMomentumSnapshots.IsEmpty(ctx, c.UserID)
	if err != nil {
		return err
	}
	if !empty {
		return nil
	}
	now := c.Now()
	yesterday, err := yesterdayISODate(now)
	if err != nil {
		return err
	}
	_, err = ensureCustomHabitMomentumSnapshot(ctx, c, yesterday, nil, nil)
	return err
}

func InvalidateCustomHabitMomentumSnapshotsFrom(
	ctx context.Context,
	c *core.Context,
	fromDate string,
) error {
	if !isISODate(fromDate) {
		return errors.New("date must be YYYY-MM-DD")
	}
	return c.CustomHabitMomentumSnapshots.DeleteFromDate(ctx, c.UserID, fromDate)
}

type customHabitMomentumSnapshot struct {
	summaries []sharedtypes.CustomHabitStreakSummary
	state     customHabitMomentumSnapshotState
}

func ensureCustomHabitMomentumSnapshot(
	ctx context.Context,
	c *core.Context,
	throughDate string,
	defs []sharedtypes.HabitStreakDefinition,
	countsByDate map[string]map[string]int,
) (*customHabitMomentumSnapshot, error) {
	if !isISODate(throughDate) {
		return nil, errors.New("date must be YYYY-MM-DD")
	}
	var err error
	if defs == nil {
		defs, err = c.HabitStreakDefinitions.List(ctx, c.UserID)
		if err != nil {
			return nil, err
		}
	}
	if len(defs) == 0 {
		return &customHabitMomentumSnapshot{summaries: nil, state: customHabitMomentumSnapshotState{}}, nil
	}
	defs = sharedtypes.NormalizeHabitStreakDefinitions(defs)
	if momentumDefinitionsCanReuseSnapshot(defs) {
		existing, err := c.CustomHabitMomentumSnapshots.GetByDate(ctx, c.UserID, throughDate)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			decoded, err := decodeCustomHabitMomentumSnapshot(existing)
			if err != nil {
				return nil, err
			}
			if decoded != nil {
				return decoded, nil
			}
		}
	}

	var (
		baseState customHabitMomentumSnapshotState
		startDate string
	)
	startDateSet := false
	prev, err := c.CustomHabitMomentumSnapshots.GetLatestBeforeDate(ctx, c.UserID, throughDate)
	if err != nil {
		return nil, err
	}
	if momentumDefinitionsCanReuseSnapshot(defs) && prev != nil {
		decoded, err := decodeCustomHabitMomentumSnapshot(prev)
		if err != nil {
			return nil, err
		}
		if decoded != nil {
			baseState = decoded.state
			startDate = nextISODate(prev.Date)
			startDateSet = true
		}
	}
	if !startDateSet {
		earliest, err := earliestMomentumHistoryDate(ctx, c, throughDate)
		if err != nil {
			return nil, err
		}
		if earliest != nil && *earliest != "" {
			startDate = *earliest
		} else {
			startDate = throughDate
		}
	}
	if startDate == "" {
		startDate = throughDate
	}
	persistEveryDay := momentumDefinitionsCanReuseSnapshot(defs) && prev != nil
	if countsByDate == nil {
		countsByDate, err = loadCustomHabitMomentumCountsByDate(ctx, c, defs, throughDate)
		if err != nil {
			return nil, err
		}
	}
	state := baseState
	for day := startDate; day <= throughDate; day = nextISODate(day) {
		state = advanceCustomHabitMomentumSnapshotState(state, day, countsByDate, defs)
		summaries := customHabitMomentumSummariesFromState(state, defs)
		if persistEveryDay || day == throughDate {
			if momentumDefinitionsCanReuseSnapshot(defs) {
				if err := persistCustomHabitMomentumSnapshot(ctx, c, day, summaries, state); err != nil {
					return nil, err
				}
			}
		}
		if day == throughDate {
			return &customHabitMomentumSnapshot{summaries: summaries, state: state}, nil
		}
	}
	return nil, errors.New("failed to compute custom habit momentum snapshot")
}

func loadCustomHabitMomentumCountsByDate(
	ctx context.Context,
	c *core.Context,
	defs []sharedtypes.HabitStreakDefinition,
	throughDate string,
) (map[string]map[string]int, error) {
	return buildCustomMomentumCounts(ctx, c, defs, throughDate)
}

func decodeCustomHabitMomentumSnapshot(
	row *storerepositories.CustomHabitMomentumSnapshotRow,
) (*customHabitMomentumSnapshot, error) {
	if row == nil {
		return nil, nil
	}
	var summaries []sharedtypes.CustomHabitStreakSummary
	if err := json.Unmarshal([]byte(row.SummaryJSON), &summaries); err != nil {
		return nil, err
	}
	var state customHabitMomentumSnapshotState
	if err := json.Unmarshal([]byte(row.StateJSON), &state); err != nil {
		return nil, err
	}
	if state.Version != customHabitMomentumSnapshotStateVersion {
		return nil, nil
	}
	return &customHabitMomentumSnapshot{summaries: summaries, state: state}, nil
}

func persistCustomHabitMomentumSnapshot(
	ctx context.Context,
	c *core.Context,
	date string,
	summaries []sharedtypes.CustomHabitStreakSummary,
	state customHabitMomentumSnapshotState,
) error {
	summaryJSON, err := json.Marshal(summaries)
	if err != nil {
		return err
	}
	stateJSON, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return c.CustomHabitMomentumSnapshots.Upsert(
		ctx,
		c.UserID,
		date,
		string(summaryJSON),
		string(stateJSON),
		c.Now(),
	)
}

func advanceCustomHabitMomentumSnapshotState(
	prev customHabitMomentumSnapshotState,
	date string,
	countsByDate map[string]map[string]int,
	defs []sharedtypes.HabitStreakDefinition,
) customHabitMomentumSnapshotState {
	prevByID := map[string]customHabitMomentumSnapshotDefState{}
	for _, item := range prev.Definitions {
		prevByID[item.ID] = item
	}
	out := customHabitMomentumSnapshotState{
		Version:     customHabitMomentumSnapshotStateVersion,
		Definitions: make([]customHabitMomentumSnapshotDefState, 0, len(defs)),
	}
	dayCounts := countsByDate[date]
	for _, rawDef := range defs {
		def := sharedtypes.NormalizeHabitStreakDefinition(rawDef)
		state := customHabitMomentumSnapshotDefState{ID: def.ID}
		if !def.Enabled {
			out.Definitions = append(out.Definitions, state)
			continue
		}

		prevState, hasPrev := prevByID[def.ID]
		count := momentumCountForDefinition(def, dayCounts)
		required := momentumRequiredUnits(def)
		prevBucketMet := prevState.BucketCount >= required
		bucketKey := customHabitBucketKey(date, def.Period)
		state.BucketKey = bucketKey
		if hasPrev && prevState.BucketKey == bucketKey && bucketKey != "" {
			state.BucketCount = prevState.BucketCount + count
			state.BucketMet = state.BucketCount >= required
			state.Current = prevState.Current
			state.Longest = prevState.Longest
			if state.BucketMet && !prevBucketMet {
				state.Current = prevState.Current + 1
			}
		} else {
			state.BucketCount = count
			state.BucketMet = state.BucketCount >= required
			if hasPrev {
				state.Longest = prevState.Longest
				state.Current = prevState.Current
				if prevBucketMet {
					if state.BucketMet {
						state.Current = prevState.Current + 1
					}
				} else if state.BucketMet {
					state.Current = 1
				} else {
					state.Current = 0
				}
			} else if state.BucketMet {
				state.Current = 1
			}
		}
		if state.Current > state.Longest {
			state.Longest = state.Current
		}
		out.Definitions = append(out.Definitions, state)
	}
	return out
}

func customHabitMomentumSummariesFromState(
	state customHabitMomentumSnapshotState,
	defs []sharedtypes.HabitStreakDefinition,
) []sharedtypes.CustomHabitStreakSummary {
	if len(defs) == 0 {
		return nil
	}
	stateByID := map[string]customHabitMomentumSnapshotDefState{}
	for _, item := range state.Definitions {
		stateByID[item.ID] = item
	}
	out := make([]sharedtypes.CustomHabitStreakSummary, 0, len(defs))
	for _, rawDef := range defs {
		def := sharedtypes.NormalizeHabitStreakDefinition(rawDef)
		summary := sharedtypes.CustomHabitStreakSummary{
			ID:            def.ID,
			Name:          def.Name,
			Enabled:       def.Enabled,
			TargetKind:    def.TargetKind,
			MatchMode:     def.MatchMode,
			Contexts:      def.Contexts,
			Period:        def.Period,
			RequiredCount: def.RequiredCount,
		}
		if def.Enabled {
			if current, ok := stateByID[def.ID]; ok {
				summary.Current = current.Current
				summary.Longest = current.Longest
			}
		}
		out = append(out, summary)
	}
	return out
}

func buildCustomMomentumCounts(
	ctx context.Context,
	c *core.Context,
	defs []sharedtypes.HabitStreakDefinition,
	throughDate string,
) (map[string]map[string]int, error) {
	history, err := c.HabitCompletions.ListHistory(ctx, c.UserID, nil, nil)
	if err != nil {
		return nil, err
	}
	sessions, err := c.Sessions.ListEnded(ctx, struct {
		UserID   string
		RepoID   *int64
		StreamID *int64
		IssueID  *int64
		Since    *string
		Until    *string
		Limit    *int
		Offset   *int
	}{
		UserID: c.UserID,
		Until:  stringPtr(throughDate + "T23:59:59.999Z"),
	})
	if err != nil {
		return nil, err
	}
	issues, err := c.Issues.ListAll(ctx, c.UserID)
	if err != nil {
		return nil, err
	}
	issueMetaByID := make(map[int64]sharedtypes.IssueWithMeta, len(issues))
	for _, issue := range issues {
		issueMetaByID[issue.ID] = issue
	}
	habitTargets := make(map[string][]int64)
	contextTargets := make(map[string][]sharedtypes.MomentumContext)
	for _, def := range defs {
		def = sharedtypes.NormalizeHabitStreakDefinition(def)
		if !def.Enabled {
			continue
		}
		switch sharedtypes.NormalizeMomentumTargetKind(def.TargetKind) {
		case sharedtypes.MomentumTargetKindContext:
			contextTargets[def.ID] = append([]sharedtypes.MomentumContext(nil), def.Contexts...)
		default:
			habitTargets[def.ID] = append([]int64(nil), def.HabitIDs...)
		}
	}
	out := map[string]map[string]int{}
	setCount := func(day, defID string, amount int) {
		if amount < 0 {
			return
		}
		dayCounts := out[day]
		if dayCounts == nil {
			dayCounts = map[string]int{}
			out[day] = dayCounts
		}
		dayCounts[defID] = amount
	}
	type habitDateEvents map[string]map[int64]int
	type contextDateEvents map[string]map[sharedtypes.MomentumContext]int
	habitEvents := make(map[string]habitDateEvents, len(habitTargets))
	for _, entry := range history {
		if entry.Status != sharedtypes.HabitCompletionStatusCompleted {
			continue
		}
		if entry.Date > throughDate {
			continue
		}
		for defID, habitIDs := range habitTargets {
			if !slices.Contains(habitIDs, entry.HabitID) {
				continue
			}
			dates := habitEvents[defID]
			if dates == nil {
				dates = habitDateEvents{}
				habitEvents[defID] = dates
			}
			dayCounts := dates[entry.Date]
			if dayCounts == nil {
				dayCounts = map[int64]int{}
				dates[entry.Date] = dayCounts
			}
			dayCounts[entry.HabitID]++
		}
	}
	contextEvents := make(map[string]contextDateEvents, len(contextTargets))
	for _, session := range sessions {
		day := extractISODate(session.StartTime)
		if day == "" || day > throughDate {
			continue
		}
		meta, ok := issueMetaByID[session.IssueID]
		if !ok {
			continue
		}
		workedSeconds := derefIssueEstimate(session.DurationSeconds)
		if workedSeconds <= 0 {
			continue
		}
		for defID, contexts := range contextTargets {
			for _, contextItem := range contexts {
				if contextItem.RepoID != meta.RepoID {
					continue
				}
				if contextItem.StreamID != nil && *contextItem.StreamID != meta.StreamID {
					continue
				}
				dates := contextEvents[defID]
				if dates == nil {
					dates = contextDateEvents{}
					contextEvents[defID] = dates
				}
				dayCounts := dates[day]
				if dayCounts == nil {
					dayCounts = map[sharedtypes.MomentumContext]int{}
					dates[day] = dayCounts
				}
				dayCounts[contextItem] += workedSeconds
			}
		}
	}
	for defID, dates := range habitEvents {
		def := defByID(defs, defID)
		if def.ID == "" {
			continue
		}
		dayKeys := make([]string, 0, len(dates))
		for day := range dates {
			dayKeys = append(dayKeys, day)
		}
		slices.Sort(dayKeys)
		for _, day := range dayKeys {
			setCount(day, def.ID, momentumDailyCount(dates[day], def.HabitIDs, def.MatchMode))
		}
	}
	for defID, dates := range contextEvents {
		def := defByID(defs, defID)
		if def.ID == "" {
			continue
		}
		dayKeys := make([]string, 0, len(dates))
		for day := range dates {
			dayKeys = append(dayKeys, day)
		}
		slices.Sort(dayKeys)
		for _, day := range dayKeys {
			setCount(day, def.ID, momentumDailyCount(dates[day], def.Contexts, def.MatchMode))
		}
	}
	return out, nil
}

func defByID(defs []sharedtypes.HabitStreakDefinition, id string) sharedtypes.HabitStreakDefinition {
	for _, def := range defs {
		if def.ID == id {
			return sharedtypes.NormalizeHabitStreakDefinition(def)
		}
	}
	return sharedtypes.HabitStreakDefinition{}
}

func momentumDefinitionsCanReuseSnapshot(defs []sharedtypes.HabitStreakDefinition) bool {
	for _, def := range defs {
		if sharedtypes.NormalizeMomentumTargetKind(def.TargetKind) != sharedtypes.MomentumTargetKindHabit {
			return false
		}
	}
	return true
}

func earliestMomentumHistoryDate(
	ctx context.Context,
	c *core.Context,
	throughDate string,
) (*string, error) {
	habitDate, err := c.HabitCompletions.EarliestDate(ctx, c.UserID, throughDate)
	if err != nil {
		return nil, err
	}
	sessionDate, err := c.Sessions.EarliestEndedDate(ctx, c.UserID, throughDate)
	if err != nil {
		return nil, err
	}
	if habitDate == nil {
		return sessionDate, nil
	}
	if sessionDate == nil {
		return habitDate, nil
	}
	if *sessionDate < *habitDate {
		return sessionDate, nil
	}
	return habitDate, nil
}

func momentumCountForDefinition(
	def sharedtypes.HabitStreakDefinition,
	dayCounts map[string]int,
) int {
	if dayCounts == nil {
		return 0
	}
	return dayCounts[def.ID]
}

func momentumRequiredUnits(def sharedtypes.HabitStreakDefinition) int {
	return max(1, def.RequiredCount)
}

func momentumDailyCount[T comparable](
	dayCounts map[T]int,
	targets []T,
	matchMode sharedtypes.MomentumMatchMode,
) int {
	if len(targets) == 0 || len(dayCounts) == 0 {
		return 0
	}
	if sharedtypes.NormalizeMomentumMatchMode(matchMode) == sharedtypes.MomentumMatchModeAll {
		return momentumAllGroupCount(dayCounts, targets)
	}
	total := 0
	for _, count := range dayCounts {
		total += count
	}
	return total
}

func momentumAllGroupCount[T comparable](cumulativeCounts map[T]int, targets []T) int {
	if len(targets) == 0 {
		return 0
	}
	minCount := -1
	for _, target := range targets {
		count := cumulativeCounts[target]
		if minCount < 0 || count < minCount {
			minCount = count
		}
	}
	if minCount < 0 {
		return 0
	}
	return minCount
}

func momentumAllContextGroupCount(
	cumulativeCounts map[sharedtypes.MomentumContext]int,
	targets []sharedtypes.MomentumContext,
) int {
	return momentumAllGroupCount(cumulativeCounts, targets)
}

func nextISODate(value string) string {
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return value
	}
	return parsed.AddDate(0, 0, 1).Format("2006-01-02")
}

func yesterdayISODate(now string) (string, error) {
	current := now
	if idx := strings.Index(current, "T"); idx >= 0 {
		current = current[:idx]
	}
	parsed, err := time.Parse("2006-01-02", current)
	if err != nil {
		return "", err
	}
	return parsed.AddDate(0, 0, -1).Format("2006-01-02"), nil
}
