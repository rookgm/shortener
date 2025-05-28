package recorder

import (
	"bufio"
	"encoding/json"
	"io"
)

type Record struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type Recorder struct{}

func NewRecorder() (*Recorder, error) {
	return &Recorder{}, nil
}

func (r *Recorder) WriteRecord(writer io.Writer, rec *Record) error {
	encoder := json.NewEncoder(writer)
	return encoder.Encode(&rec)
}

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
