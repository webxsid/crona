package export

import (
	"strconv"
	"strings"

	versionpkg "crona/shared/version"
)

func attachFrontmatter(data map[string]any, spec reportWriteSpec) map[string]any {
	if data == nil {
		data = map[string]any{}
	}
	title := reportFrontmatterTitle(spec, data)
	repoName := nestedString(data, "repo", "name")
	streamName := nestedString(data, "stream", "name")
	tags := []string{"crona", "report", "crona/" + slugify(string(spec.Kind))}
	if repoName != "" {
		tags = append(tags, "repo/"+slugify(repoName))
	}
	if streamName != "" {
		tags = append(tags, "stream/"+slugify(streamName))
	}
	fields := map[string]any{
		"title":        yamlString(title),
		"aliases":      yamlStringList([]string{title}),
		"tags":         yamlStringList(tags),
		"reportKind":   yamlString(string(spec.Kind)),
		"date":         yamlString(spec.Date),
		"startDate":    yamlString(spec.StartDate),
		"endDate":      yamlString(spec.EndDate),
		"generatedAt":  yamlString(firstNonEmptyString(stringFromMap(data, "generatedAt"), spec.Date)),
		"created":      yamlString(firstNonEmptyString(stringFromMap(data, "generatedAt"), spec.Date)),
		"updated":      yamlString(firstNonEmptyString(stringFromMap(data, "generatedAt"), spec.Date)),
		"scope":        yamlString(spec.ScopeLabel),
		"repo":         yamlString(repoName),
		"stream":       yamlString(streamName),
		"cronaVersion": yamlString(versionpkg.Current()),
	}
	data["frontmatter"] = fields
	data["frontmatterBlock"] = renderFrontmatterBlock(spec, fields)
	return data
}

func ensureMarkdownFrontmatter(content string, data map[string]any, spec reportWriteSpec) string {
	trimmed := strings.TrimSpace(content)
	if strings.HasPrefix(trimmed, "---") {
		return trimmed
	}
	withFrontmatter := attachFrontmatter(data, spec)
	block, _ := withFrontmatter["frontmatterBlock"].(string)
	if strings.TrimSpace(block) == "" {
		return trimmed
	}
	if trimmed == "" {
		return block
	}
	return block + "\n\n" + trimmed
}

func renderFrontmatterBlock(spec reportWriteSpec, fields map[string]any) string {
	lines := []string{
		"---",
		"title: " + frontmatterString(fields, "title"),
		"aliases:",
	}
	for _, item := range frontmatterList(fields, "aliases") {
		lines = append(lines, "  - "+item)
	}
	lines = append(lines, "tags:")
	for _, item := range frontmatterList(fields, "tags") {
		lines = append(lines, "  - "+item)
	}
	lines = append(lines,
		"report_kind: "+frontmatterString(fields, "reportKind"),
	)
	switch spec.Kind {
	case "daily":
		lines = append(lines, "date: "+frontmatterString(fields, "date"))
	case "weekly", "repo", "stream", "issue_rollup":
		lines = append(lines,
			"start_date: "+frontmatterString(fields, "startDate"),
			"end_date: "+frontmatterString(fields, "endDate"),
		)
	}
	lines = append(lines,
		"generated_at: "+frontmatterString(fields, "generatedAt"),
		"created: "+frontmatterString(fields, "created"),
		"updated: "+frontmatterString(fields, "updated"),
	)
	switch spec.Kind {
	case "repo":
		lines = append(lines,
			"scope: "+frontmatterString(fields, "scope"),
			"repo: "+frontmatterString(fields, "repo"),
		)
	case "stream":
		lines = append(lines,
			"scope: "+frontmatterString(fields, "scope"),
			"repo: "+frontmatterString(fields, "repo"),
			"stream: "+frontmatterString(fields, "stream"),
		)
	case "issue_rollup":
		if frontmatterString(fields, "scope") != `""` {
			lines = append(lines, "scope: "+frontmatterString(fields, "scope"))
		}
	}
	lines = append(lines,
		"crona_version: "+frontmatterString(fields, "cronaVersion"),
		"---",
	)
	return strings.Join(lines, "\n")
}

func frontmatterString(fields map[string]any, key string) string {
	value, _ := fields[key].(string)
	if strings.TrimSpace(value) == "" {
		return `""`
	}
	return value
}

func frontmatterList(fields map[string]any, key string) []string {
	values, _ := fields[key].([]string)
	return values
}

func reportFrontmatterTitle(spec reportWriteSpec, data map[string]any) string {
	switch spec.Kind {
	case "daily":
		return strings.TrimSpace("Daily Report - " + spec.Date)
	case "weekly":
		return strings.TrimSpace("Weekly Summary - " + joinNonEmpty(" to ", spec.StartDate, spec.EndDate))
	case "repo":
		if repoName := nestedString(data, "repo", "name"); repoName != "" {
			return "Repo Report - " + repoName
		}
	case "stream":
		if streamName := nestedString(data, "stream", "name"); streamName != "" {
			return "Stream Report - " + streamName
		}
	case "issue_rollup":
		return strings.TrimSpace("Session to Issue Rollup - " + joinNonEmpty(" to ", spec.StartDate, spec.EndDate))
	}
	if spec.Label != "" {
		return spec.Label
	}
	return string(spec.Kind)
}

func yamlStringList(values []string) []string {
	out := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, yamlString(value))
	}
	return out
}

func yamlString(value string) string {
	return strconv.Quote(strings.TrimSpace(value))
}

func nestedString(data map[string]any, key, nested string) string {
	raw, ok := data[key]
	if !ok {
		return ""
	}
	items, ok := raw.(map[string]any)
	if !ok {
		return ""
	}
	value, _ := items[nested].(string)
	return strings.TrimSpace(value)
}

func stringFromMap(data map[string]any, key string) string {
	value, _ := data[key].(string)
	return strings.TrimSpace(value)
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
