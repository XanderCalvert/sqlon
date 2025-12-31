package pipeline

import (
	"encoding/json"
	"io"
	"time"
)

type Logger struct {
	w io.Writer
}

func NewLogger(w io.Writer) *Logger {
	return &Logger{w: w}
}

func (l *Logger) Log(stepName string, in []byte, out []byte, artefact string, notes string) error {
	payload := stepLog{
		Time:      time.Now().UTC().Format(time.RFC3339Nano),
		Step:      stepName,
		InBytes:   len(in),
		OutBytes:  len(out),
		InSHA256:  sha256Hex(in),
		OutSHA256: sha256Hex(out),
		Artefact:  artefact,
		Notes:     notes,
	}

	enc := json.NewEncoder(l.w)
	return enc.Encode(payload)
}
