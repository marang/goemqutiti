package steps

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ReadFile reads a CSV file and returns rows as maps keyed by header name.
func ReadFile(path string) ([]map[string]string, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".csv":
		return readCSV(path)
	default:
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}
}

func readCSV(path string) ([]map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	headers, err := r.Read()
	if err != nil {
		return nil, err
	}
	var rows []map[string]string
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		row := map[string]string{}
		for i, h := range headers {
			if i < len(rec) {
				row[h] = rec[i]
			} else {
				row[h] = ""
			}
		}
		rows = append(rows, row)
	}
	return rows, nil
}

var placeholder = regexp.MustCompile(`\{([^}]+)\}`)

// BuildTopic replaces placeholders in tmpl with values from fields.
func BuildTopic(tmpl string, fields map[string]string) string {
	return placeholder.ReplaceAllStringFunc(tmpl, func(s string) string {
		key := strings.Trim(s, "{}")
		if v, ok := fields[key]; ok {
			return v
		}
		return ""
	})
}

// RowToJSON converts a row map to a JSON payload using the provided field mapping.
func RowToJSON(row map[string]string, mapping map[string]string) ([]byte, error) {
	out := map[string]interface{}{}
	for k, v := range row {
		path := k
		if mapped, ok := mapping[k]; ok && strings.TrimSpace(mapped) != "" {
			path = mapped
		}
		setNested(out, strings.Split(path, "."), v)
	}
	return json.Marshal(out)
}

func setNested(m map[string]interface{}, path []string, value string) {
	for i, p := range path {
		if i == len(path)-1 {
			m[p] = value
			return
		}
		if _, ok := m[p]; !ok {
			m[p] = map[string]interface{}{}
		}
		child, _ := m[p].(map[string]interface{})
		m = child
	}
}
