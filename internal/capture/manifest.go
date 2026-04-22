package capture

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type manifestEntry struct {
	URL   string  `json:"url"`
	Name  string  `json:"name"`
	File  string  `json:"file"`
	Path  string  `json:"path"`
	Error *string `json:"error"`
}

type manifestDoc struct {
	Generated string          `json:"generated"`
	Count     int             `json:"count"`
	Routes    []manifestEntry `json:"routes"`
}

// WriteManifest writes <outDir>/manifest.json summarising all BulkResults.
// Entries where both Result and Err are nil are skipped.
func WriteManifest(outDir string, results []BulkResult) error {
	var entries []manifestEntry
	for _, br := range results {
		if br.Result == nil && br.Err == nil {
			continue
		}
		e := manifestEntry{
			URL:  br.Job.URL,
			Name: br.Job.Name,
		}
		if br.Err != nil {
			msg := br.Err.Error()
			e.Error = &msg
		}
		if br.Result != nil {
			e.File = filepath.Base(br.Result.Path)
			e.Path = br.Result.Path
		}
		entries = append(entries, e)
	}

	doc := manifestDoc{
		Generated: time.Now().UTC().Format(time.RFC3339),
		Count:     len(entries),
		Routes:    entries,
	}

	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(outDir, "manifest.json"), data, 0o644)
}
