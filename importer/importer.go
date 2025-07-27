package importer

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/xuri/excelize/v2"
)

// ReadFile reads a CSV or Excel file and returns rows as maps keyed by header name.
func ReadFile(path string) ([]map[string]string, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".csv":
		return readCSV(path)
	case ".xls", ".xlsx":
		return readXLS(path)
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

func readXLS(path string) ([]map[string]string, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	sheet := f.GetSheetName(0)
	if sheet == "" {
		return nil, fmt.Errorf("no sheets found")
	}
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	headers := rows[0]
	var result []map[string]string
	for _, rec := range rows[1:] {
		row := map[string]string{}
		for i, h := range headers {
			if i < len(rec) {
				row[h] = rec[i]
			} else {
				row[h] = ""
			}
		}
		result = append(result, row)
	}
	return result, nil
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
