package pipeline

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type Step interface {
	Name() string
	Ext() string
	Run(in []byte) ([]byte, error)
}

type Runner struct {
	OutDir string
	LogW   io.Writer
}

type stepLog struct {
	Time      string `json:"time"`
	Step      string `json:"step"`
	InBytes   int    `json:"in_bytes"`
	OutBytes  int    `json:"out_bytes"`
	InSHA256  string `json:"in_sha256"`
	OutSHA256 string `json:"out_sha256"`
	Artefact  string `json:"artefact"`
	Notes     string `json:"notes,omitempty"`
}

func (r *Runner) Run(steps []Step, input []byte, prefix string) ([]byte, error) {
	if err := os.MkdirAll(r.OutDir, 0o755); err != nil {
		return nil, err
	}

	current := input

	for i, step := range steps {
		out, err := step.Run(current)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", step.Name(), err)
		}

		artefact := filepath.Join(
			r.OutDir,
			fmt.Sprintf("%s%02d.%s", prefix, i+1, step.Ext()),
		)

		if err := os.WriteFile(artefact, out, 0o644); err != nil {
			return nil, err
		}

		if r.LogW != nil {
			payload := stepLog{
				Time:      time.Now().UTC().Format(time.RFC3339Nano),
				Step:      step.Name(),
				InBytes:   len(current),
				OutBytes:  len(out),
				InSHA256:  sha256Hex(current),
				OutSHA256: sha256Hex(out),
				Artefact:  artefact,
			}
			enc := json.NewEncoder(r.LogW)
			if err := enc.Encode(payload); err != nil {
				return nil, err
			}
		}

		current = out
	}

	return current, nil
}

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
