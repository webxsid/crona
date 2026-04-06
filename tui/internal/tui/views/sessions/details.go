package sessions

import "strings"

func collapseSpace(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}
