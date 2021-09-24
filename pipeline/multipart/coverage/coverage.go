package coverage

import (
	"encoding/json"
	"fmt"
	"strconv"

	"mime/multipart"
	"net/textproto"
)

// MimeType used by coverage reports.
const MimeType = "application/json+coverage"

type (
	// Report represents a coverage report.
	Report struct {
		Timestamp int64   `json:"timestmp,omitempty"`
		Command   string  `json:"command_name,omitempty"`
		Files     []File  `json:"files"`
		Metrics   Metrics `json:"metrics"`
	}

	// File represents a coverage report for a single file.
	File struct {
		Name            string  `json:"filename"`
		Digest          string  `json:"checksum,omitempty"`
		Coverage        []*int  `json:"coverage"`
		Covered         float64 `json:"covered_percent,omitempty"`
		CoveredStrength float64 `json:"covered_strength,omitempty"`
		CoveredLines    int     `json:"covered_lines,omitempty"`
		TotalLines      int     `json:"lines_of_code"`
	}

	// Metrics represents total coverage metrics for all files.
	Metrics struct {
		Covered         float64 `json:"covered_percent"`
		CoveredStrength float64 `json:"covered_strength"`
		CoveredLines    int     `json:"covered_lines"`
		TotalLines      int     `json:"total_lines"`
	}
)

// WriteTo writes the report to multipart.Writer w.
func (r *Report) WriteTo(w *multipart.Writer) error {
	header := textproto.MIMEHeader{}
	header.Set("Content-Type", MimeType)
	header.Set("X-Covered", fmt.Sprintf("%.2f", r.Metrics.Covered))
	header.Set("X-Covered-Lines", strconv.Itoa(r.Metrics.CoveredLines))
	header.Set("X-Total-Lines", strconv.Itoa(r.Metrics.TotalLines))
	part, err := w.CreatePart(header)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(part)
	encoder.SetIndent("", "  ")
	return encoder.Encode(r)
}
