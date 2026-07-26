// Package prism provides a datasource that reads PRISM Control JSONL export
// files produced by `prismctl export`. It loads initiative and RMI data for
// use in devfolio contributor profiles.
package prism

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// Initiative represents a PRISM Control initiative as exported by prismctl.
type Initiative struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status"`
	CreatedAt   string `json:"createdAt,omitempty"`
	UpdatedAt   string `json:"updatedAt,omitempty"`
}

// RMI represents a PRISM Control Roadmap Item as exported by prismctl.
type RMI struct {
	ID           string `json:"id"`
	Repo         string `json:"repo"`
	Initiative   string `json:"initiative"`
	Phase        string `json:"phase"`
	Title        string `json:"title"`
	Type         string `json:"type"`
	Status       string `json:"status"`
	Required     bool   `json:"required"`
	AssignedTo   string `json:"assignedTo,omitempty"`
	CompletedAt  string `json:"completedAt,omitempty"`
	CreatedAt    string `json:"createdAt,omitempty"`
	UpdatedAt    string `json:"updatedAt,omitempty"`
}

// ExportRecord represents a single line in a PRISM Control JSONL export.
// Each line has a "kind" discriminator ("initiative" or "rmi") and the
// corresponding data in the matching field.
type ExportRecord struct {
	Kind       string      `json:"kind"`
	Initiative *Initiative `json:"initiative,omitempty"`
	RMI        *RMI        `json:"rmi,omitempty"`
	ExportedAt string      `json:"exportedAt,omitempty"`
}

// ExportData holds the parsed result of a PRISM Control JSONL export.
type ExportData struct {
	Initiatives []Initiative
	RMIs        []RMI
	ExportedAt  time.Time
}

// RMIsByInitiative returns a map from initiative ID to its RMIs.
func (d *ExportData) RMIsByInitiative() map[string][]RMI {
	m := make(map[string][]RMI, len(d.Initiatives))
	for _, rmi := range d.RMIs {
		m[rmi.Initiative] = append(m[rmi.Initiative], rmi)
	}
	return m
}

// RMIsByRepo returns a map from repository to its RMIs.
func (d *ExportData) RMIsByRepo() map[string][]RMI {
	m := make(map[string][]RMI)
	for _, rmi := range d.RMIs {
		m[rmi.Repo] = append(m[rmi.Repo], rmi)
	}
	return m
}

// LoadFile reads a PRISM Control JSONL export file and returns parsed data.
func LoadFile(path string) (*ExportData, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("prism: open export file: %w", err)
	}
	defer func() { _ = f.Close() }()
	return Load(f)
}

// Load reads a PRISM Control JSONL export from the given reader.
func Load(r io.Reader) (*ExportData, error) {
	data := &ExportData{}
	scanner := bufio.NewScanner(r)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var rec ExportRecord
		if err := json.Unmarshal(line, &rec); err != nil {
			return nil, fmt.Errorf("prism: line %d: %w", lineNum, err)
		}

		switch rec.Kind {
		case "initiative":
			if rec.Initiative == nil {
				return nil, fmt.Errorf("prism: line %d: kind is 'initiative' but initiative field is nil", lineNum)
			}
			data.Initiatives = append(data.Initiatives, *rec.Initiative)
		case "rmi":
			if rec.RMI == nil {
				return nil, fmt.Errorf("prism: line %d: kind is 'rmi' but rmi field is nil", lineNum)
			}
			data.RMIs = append(data.RMIs, *rec.RMI)
		default:
			return nil, fmt.Errorf("prism: line %d: unknown kind %q", lineNum, rec.Kind)
		}

		if rec.ExportedAt != "" {
			t, err := time.Parse(time.RFC3339, rec.ExportedAt)
			if err == nil {
				data.ExportedAt = t
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("prism: reading export: %w", err)
	}

	return data, nil
}
