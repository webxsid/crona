package commands

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"crona/kernel/internal/core"
	storerepositories "crona/kernel/internal/store/repositories"
	sharedtypes "crona/shared/types"
)

type customHabitMomentumSnapshotState struct {
	Definitions []customHabitMomentumSnapshotDefState `json:"definitions,omitempty"`
}

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
	snapshot, err := ensureCustomHabitMomentumSnapshot(ctx, c, throughDate)
	if err != nil {
		return nil, err
	}
	if snapshot == nil {
		return nil, nil
	}
	return snapshot.summaries, nil
}

func SeedCustomHabitMomentumSnapshot(ctx context.Context, c *core.Context) error {
	settings, err := c.CoreSettings.Get(ctx, c.UserID)
	if err != nil {
		return err
	}
	if settings == nil || len(settings.HabitStreakDefs) == 0 {
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
	_, err = ensureCustomHabitMomentumSnapshot(ctx, c, yesterday)
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
) (*customHabitMomentumSnapshot, error) {
	if !isISODate(throughDate) {
		return nil, errors.New("date must be YYYY-MM-DD")
	}
	settings, err := c.CoreSettings.Get(ctx, c.UserID)
	if err != nil {
		return nil, err
	}
	if settings == nil || len(settings.HabitStreakDefs) == 0 {
		return &customHabitMomentumSnapshot{summaries: nil, state: customHabitMomentumSnapshotState{}}, nil
	}
	settings.HabitStreakDefs = sharedtypes.NormalizeHabitStreakDefinitions(settings.HabitStreakDefs)
	existing, err := c.CustomHabitMomentumSnapshots.GetByDate(ctx, c.UserID, throughDate)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return decodeCustomHabitMomentumSnapshot(existing)
	}

	prev, err := c.CustomHabitMomentumSnapshots.GetLatestBeforeDate(ctx, c.UserID, throughDate)
	if err != nil {
		return nil, err
	}

	var (
		baseState customHabitMomentumSnapshotState
		startDate string
	)
	if prev != nil {
		decoded, err := decodeCustomHabitMomentumSnapshot(prev)
		if err != nil {
			return nil, err
		}
		baseState = decoded.state
		startDate = nextISODate(prev.Date)
	} else {
		earliest, err := c.HabitCompletions.EarliestDate(ctx, c.UserID, throughDate)
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
	persistEveryDay := prev != nil

	history, err := c.HabitCompletions.ListHistory(ctx, c.UserID, nil, nil)
	if err != nil {
		return nil, err
	}
	countsByDate := buildCustomHabitCounts(history, settings.HabitStreakDefs, throughDate)
	state := baseState
	for day := startDate; day <= throughDate; day = nextISODate(day) {
		state = advanceCustomHabitMomentumSnapshotState(state, day, countsByDate, settings.HabitStreakDefs)
		summaries := customHabitMomentumSummariesFromState(state, settings.HabitStreakDefs)
		if persistEveryDay || day == throughDate {
			if err := persistCustomHabitMomentumSnapshot(ctx, c, day, summaries, state); err != nil {
				return nil, err
			}
		}
		if day == throughDate {
			return &customHabitMomentumSnapshot{summaries: summaries, state: state}, nil
		}
	}
	return nil, errors.New("failed to compute custom habit momentum snapshot")
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

func buildCustomHabitCounts(
	history []sharedtypes.HabitCompletion,
	defs []sharedtypes.HabitStreakDefinition,
	throughDate string,
) map[string]map[string]int {
	habitToDefs := make(map[int64][]string, len(defs))
	for _, def := range defs {
		if !def.Enabled || len(def.HabitIDs) == 0 {
			continue
		}
		for _, habitID := range def.HabitIDs {
			habitToDefs[habitID] = append(habitToDefs[habitID], def.ID)
		}
	}
	out := map[string]map[string]int{}
	for _, entry := range history {
		if entry.Status != sharedtypes.HabitCompletionStatusCompleted {
			continue
		}
		if entry.Date > throughDate {
			continue
		}
		defIDs := habitToDefs[entry.HabitID]
		if len(defIDs) == 0 {
			continue
		}
		dayCounts := out[entry.Date]
		if dayCounts == nil {
			dayCounts = map[string]int{}
			out[entry.Date] = dayCounts
		}
		for _, defID := range defIDs {
			dayCounts[defID]++
		}
	}
	return out
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
		Definitions: make([]customHabitMomentumSnapshotDefState, 0, len(defs)),
	}
	dayCounts := countsByDate[date]
	for _, rawDef := range defs {
		def := sharedtypes.NormalizeHabitStreakDefinition(rawDef)
		state := customHabitMomentumSnapshotDefState{ID: def.ID}
		if !def.Enabled || len(def.HabitIDs) == 0 {
			out.Definitions = append(out.Definitions, state)
			continue
		}

		prevState, hasPrev := prevByID[def.ID]
		count := 0
		if dayCounts != nil {
			count = dayCounts[def.ID]
		}
		bucketKey := customHabitBucketKey(date, def.Period)
		state.BucketKey = bucketKey
		if hasPrev && prevState.BucketKey == bucketKey && bucketKey != "" {
			state.BucketCount = prevState.BucketCount + count
			state.BucketMet = state.BucketCount >= def.RequiredCount
			state.Current = prevState.Current
			state.Longest = prevState.Longest
			if state.BucketMet && prevState.BucketCount < def.RequiredCount {
				state.Current = prevState.Current + 1
			}
		} else {
			state.BucketCount = count
			state.BucketMet = state.BucketCount >= def.RequiredCount
			if hasPrev {
				state.Longest = prevState.Longest
				state.Current = prevState.Current
				if prevState.BucketMet {
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
			Period:        def.Period,
			RequiredCount: def.RequiredCount,
		}
		if def.Enabled && len(def.HabitIDs) > 0 {
			if current, ok := stateByID[def.ID]; ok {
				summary.Current = current.Current
				summary.Longest = current.Longest
			}
		}
		out = append(out, summary)
	}
	return out
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
