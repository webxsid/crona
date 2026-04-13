package export

import (
	"fmt"
	"reflect"
	"strings"
)

type node interface{}

type textNode struct {
	text string
}

type varNode struct {
	path string
}

type ifNode struct {
	path     string
	children []node
}

type eachNode struct {
	path     string
	children []node
}

func RenderTemplate(source string, data map[string]any) (string, error) {
	nodes, pos, err := parseNodes(source, 0, "")
	if err != nil {
		return "", err
	}
	if pos != len(source) {
		return "", fmt.Errorf("unexpected template tail at %d", pos)
	}
	rendered, err := renderNodes(nodes, data, data)
	if err != nil {
		return "", err
	}
	return compactExcessBlankLines(rendered), nil
}

func parseNodes(source string, pos int, endTag string) ([]node, int, error) {
	nodes := make([]node, 0)
	for pos < len(source) {
		open := strings.Index(source[pos:], "{{")
		if open < 0 {
			nodes = append(nodes, textNode{text: source[pos:]})
			return nodes, len(source), nil
		}
		open += pos
		if open > pos {
			nodes = append(nodes, textNode{text: source[pos:open]})
		}
		closeIdx := strings.Index(source[open+2:], "}}")
		if closeIdx < 0 {
			return nil, 0, fmt.Errorf("unterminated handlebars token")
		}
		closeIdx += open + 2
		token := strings.TrimSpace(source[open+2 : closeIdx])
		pos = closeIdx + 2
		if isControlToken(token) {
			if nextPos, ok := standaloneControlLineEnd(source, open, closeIdx+2); ok {
				trimStandalonePrefix(&nodes)
				pos = nextPos
			}
		}

		if token == endTag {
			return nodes, pos, nil
		}
		switch {
		case strings.HasPrefix(token, "#if "):
			path := strings.TrimSpace(strings.TrimPrefix(token, "#if "))
			children, nextPos, err := parseNodes(source, pos, "/if")
			if err != nil {
				return nil, 0, err
			}
			nodes = append(nodes, ifNode{path: path, children: children})
			pos = nextPos
		case strings.HasPrefix(token, "#each "):
			path := strings.TrimSpace(strings.TrimPrefix(token, "#each "))
			children, nextPos, err := parseNodes(source, pos, "/each")
			if err != nil {
				return nil, 0, err
			}
			nodes = append(nodes, eachNode{path: path, children: children})
			pos = nextPos
		case strings.HasPrefix(token, "/"):
			return nil, 0, fmt.Errorf("unexpected closing tag %q", token)
		default:
			nodes = append(nodes, varNode{path: token})
		}
	}
	if endTag != "" {
		return nil, 0, fmt.Errorf("missing closing tag %q", endTag)
	}
	return nodes, pos, nil
}

func isControlToken(token string) bool {
	return strings.HasPrefix(token, "#if ") ||
		strings.HasPrefix(token, "#each ") ||
		token == "/if" ||
		token == "/each"
}

func standaloneControlLineEnd(source string, open int, closeEnd int) (int, bool) {
	lineStart := strings.LastIndex(source[:open], "\n") + 1
	for _, r := range source[lineStart:open] {
		if r != ' ' && r != '\t' && r != '\r' {
			return closeEnd, false
		}
	}
	pos := closeEnd
	for pos < len(source) {
		switch source[pos] {
		case ' ', '\t':
			pos++
		case '\r':
			if pos+1 < len(source) && source[pos+1] == '\n' {
				return pos + 2, true
			}
			return pos + 1, true
		case '\n':
			return pos + 1, true
		default:
			return closeEnd, false
		}
	}
	return pos, true
}

func trimStandalonePrefix(nodes *[]node) {
	if len(*nodes) == 0 {
		return
	}
	last, ok := (*nodes)[len(*nodes)-1].(textNode)
	if !ok {
		return
	}
	idx := strings.LastIndex(last.text, "\n")
	prefix := last.text[idx+1:]
	if strings.TrimSpace(prefix) != "" {
		return
	}
	last.text = last.text[:idx+1]
	if last.text == "" {
		*nodes = (*nodes)[:len(*nodes)-1]
		return
	}
	(*nodes)[len(*nodes)-1] = last
}

func compactExcessBlankLines(value string) string {
	for strings.Contains(value, "\n\n\n") {
		value = strings.ReplaceAll(value, "\n\n\n", "\n\n")
	}
	return value
}

func renderNodes(nodes []node, current any, root any) (string, error) {
	var out strings.Builder
	for _, item := range nodes {
		switch node := item.(type) {
		case textNode:
			out.WriteString(node.text)
		case varNode:
			value := resolvePath(node.path, current, root)
			out.WriteString(formatValue(value))
		case ifNode:
			value := resolvePath(node.path, current, root)
			if isTruthy(value) {
				rendered, err := renderNodes(node.children, current, root)
				if err != nil {
					return "", err
				}
				out.WriteString(rendered)
			}
		case eachNode:
			values := resolvePath(node.path, current, root)
			rv := reflect.ValueOf(values)
			if !rv.IsValid() {
				continue
			}
			if rv.Kind() == reflect.Pointer {
				if rv.IsNil() {
					continue
				}
				rv = rv.Elem()
			}
			if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
				continue
			}
			for i := 0; i < rv.Len(); i++ {
				rendered, err := renderNodes(node.children, rv.Index(i).Interface(), root)
				if err != nil {
					return "", err
				}
				out.WriteString(rendered)
			}
		default:
			return "", fmt.Errorf("unsupported template node %T", item)
		}
	}
	return out.String(), nil
}

func resolvePath(path string, current any, root any) any {
	switch path {
	case "", "this", ".":
		return current
	}
	if strings.HasPrefix(path, "@root.") {
		return resolveSegments(strings.Split(strings.TrimPrefix(path, "@root."), "."), root)
	}
	return resolveSegments(strings.Split(path, "."), current)
}

func resolveSegments(parts []string, value any) any {
	current := value
	for _, part := range parts {
		if part == "" {
			continue
		}
		current = resolveSegment(current, part)
		if current == nil {
			return nil
		}
	}
	return current
}

func resolveSegment(value any, part string) any {
	if value == nil {
		return nil
	}
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}
	switch rv.Kind() {
	case reflect.Map:
		key := reflect.ValueOf(part)
		item := rv.MapIndex(key)
		if item.IsValid() {
			return item.Interface()
		}
	case reflect.Struct:
		field := rv.FieldByNameFunc(func(name string) bool { return strings.EqualFold(name, part) })
		if field.IsValid() {
			return field.Interface()
		}
	}
	return nil
}

func formatValue(value any) string {
	if value == nil {
		return ""
	}
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return ""
		}
		return formatValue(rv.Elem().Interface())
	}
	return fmt.Sprint(value)
}

func isTruthy(value any) bool {
	if value == nil {
		return false
	}
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Pointer {
		return !rv.IsNil() && isTruthy(rv.Elem().Interface())
	}
	switch rv.Kind() {
	case reflect.Bool:
		return rv.Bool()
	case reflect.String:
		return strings.TrimSpace(rv.String()) != ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv.Uint() != 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() != 0
	case reflect.Slice, reflect.Array, reflect.Map:
		return rv.Len() > 0
	default:
		return true
	}
}
