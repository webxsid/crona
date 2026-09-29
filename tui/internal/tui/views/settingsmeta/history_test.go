package settingsmeta

import (
	"testing"

	sharedtypes "crona/shared/types"
)

func TestHistoryEditWindowAppearsOnlyWhenManageHistoryEnabled(t *testing.T) {
	settings := &sharedtypes.CoreSettings{ManageHistoryDays: 3}
	rows := Rows(settings)
	if findRow(rows, "History Edit Window") >= 0 {
		t.Fatal("history edit window should be hidden when manage history is disabled")
	}
	settings.ManageHistory = true
	rows = Rows(settings)
	index := findRow(rows, "History Edit Window")
	if index < 0 || rows[index].Value != "3 days" {
		t.Fatalf("expected configured history window row, got %+v", rows)
	}
}

func findRow(rows []Row, label string) int {
	for index, row := range rows {
		if row.Label == label {
			return index
		}
	}
	return -1
}
