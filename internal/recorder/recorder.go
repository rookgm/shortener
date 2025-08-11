package recorder

import (
	"bufio"
	"encoding/json"
	"io"
)

// Record is record entity
type Record struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// Recorder is recorder
type Recorder struct{}

// NewRecorder creates new Recorder
func NewRecorder() (*Recorder, error) {
	return &Recorder{}, nil
}

// WriteRecord writes record
func (r *Recorder) WriteRecord(writer io.Writer, rec *Record) error {
	encoder := json.NewEncoder(writer)
	return encoder.Encode(&rec)
}

// ReadAllRecords reading all records
func (r *Recorder) ReadAllRecords(reader io.Reader) (map[string]string, error) {

	m := make(map[string]string)

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		rec := Record{}
		err := json.Unmarshal(scanner.Bytes(), &rec)
		if err != nil {
			return nil, err
		}
		m[rec.ShortURL] = rec.OriginalURL
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return m, nil
}
