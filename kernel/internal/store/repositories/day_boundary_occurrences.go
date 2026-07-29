package repositories

import (
	"context"
	"time"

	storemodels "crona/kernel/internal/store/models"

	"github.com/uptrace/bun"
)

type DayBoundaryOccurrenceRepository struct {
	db *bun.DB
}

func NewDayBoundaryOccurrenceRepository(db *bun.DB) *DayBoundaryOccurrenceRepository {
	return &DayBoundaryOccurrenceRepository{db: db}
}

// Claim atomically records an occurrence. It returns false when it was already claimed.
func (r *DayBoundaryOccurrenceRepository) Claim(
	ctx context.Context,
	id string,
	userID string,
	kind string,
	scheduledAtUTC string,
	timezone string,
) (bool, error) {
	result, err := r.db.NewInsert().Model(&storemodels.DayBoundaryOccurrenceModel{
		ID:             id,
		UserID:         userID,
		Kind:           kind,
		ScheduledAtUTC: scheduledAtUTC,
		Timezone:       timezone,
		ClaimedAtUTC:   time.Now().UTC().Format(time.RFC3339Nano),
	}).On("CONFLICT (id) DO NOTHING").Exec(ctx)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	return rows == 1, err
}
