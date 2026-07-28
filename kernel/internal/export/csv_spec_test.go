package export

import (
	"encoding/json"
	"testing"
)

func TestCSVExportSpecAcceptsLegacyStringColumns(t *testing.T) {
	var spec csvExportSpec
	err := json.Unmarshal([]byte(`{
		"headers": ["Issue ID", "Title"],
		"columns": ["issueId", "issueTitle"]
	}`), &spec)
	if err != nil {
		t.Fatalf("unmarshal legacy CSV spec: %v", err)
	}
	if len(spec.Columns) != 2 ||
		spec.Columns[0].Header != "Issue ID" ||
		spec.Columns[0].Field != "issueId" ||
		spec.Columns[1].Header != "Title" ||
		spec.Columns[1].Field != "issueTitle" {
		t.Fatalf("unexpected legacy columns: %#v", spec.Columns)
	}
}

func TestCSVExportSpecAcceptsColumnObjects(t *testing.T) {
	var spec csvExportSpec
	err := json.Unmarshal([]byte(`{
		"columns": [{"header": "Issue ID", "field": "issueId"}]
	}`), &spec)
	if err != nil {
		t.Fatalf("unmarshal current CSV spec: %v", err)
	}
	if len(spec.Columns) != 1 ||
		spec.Columns[0].Header != "Issue ID" ||
		spec.Columns[0].Field != "issueId" {
		t.Fatalf("unexpected columns: %#v", spec.Columns)
	}
}
