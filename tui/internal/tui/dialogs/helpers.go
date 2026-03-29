package dialogs

import (
	shareddto "crona/shared/dto"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
)

func ParseEstimateInput(raw string) (*int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return nil, fmt.Errorf("estimate must be a non-negative integer")
	}
	return &value, nil
}

func ParseDueDateInput(raw string) (*string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	if _, err := time.Parse("2006-01-02", raw); err != nil {
		return nil, fmt.Errorf("due date must be YYYY-MM-DD")
	}
	return &raw, nil
}

func ValueToPointer(raw string) *string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	return &raw
}

func ValueOrEmpty(raw *string) string {
	if raw == nil {
		return ""
	}
	return *raw
}

func ParseNumericID(raw string) int64 {
	value, _ := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	return value
}

func EndSessionRequest(inputs []textinput.Model) shareddto.EndSessionRequest {
	if len(inputs) == 0 {
		return shareddto.EndSessionRequest{}
	}
	req := shareddto.EndSessionRequest{
		CommitMessage: ValueToPointer(inputs[0].Value()),
	}
	if len(inputs) > 1 {
		req.WorkedOn = ValueToPointer(inputs[1].Value())
	}
	if len(inputs) > 2 {
		req.Outcome = ValueToPointer(inputs[2].Value())
	}
	if len(inputs) > 3 {
		req.NextStep = ValueToPointer(inputs[3].Value())
	}
	if len(inputs) > 4 {
		req.Blockers = ValueToPointer(inputs[4].Value())
	}
	if len(inputs) > 5 {
		req.Links = ValueToPointer(inputs[5].Value())
	}
	return req
}
