package metamonster

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
)

type Metamonster struct {
	URL         string
	Title       string // optimized
	Keywords    string // optimized
	Description string // optimized
}

func Report(ctx context.Context, path string) (map[string]Metamonster, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("Error opening csv file %q: %s", path, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("Error reading CSV: %s", err)
	}

	var report = make(map[string]Metamonster)

	// handle the parsed CSV data
	// url issues crawled_page_title optimized_page_title crawled_meta_description optimized_meta_description primary_keyword
	for _, row := range records {
		report[row[0]] = Metamonster{
			URL:         row[0],
			Title:       row[3],
			Keywords:    row[6],
			Description: row[5],
		}
	}

	return report, nil
}
