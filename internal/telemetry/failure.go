package telemetry

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"go-base/internal/domain"
)

type Failure struct {
	Code    string
	Message string
	Cause   error
	Labels  map[string]string
	Path    []string
}

func NewFailure(code, message string, cause error, labels map[string]string, path []string) (*Failure, error) {
	if strings.TrimSpace(code) == "" || strings.TrimSpace(message) == "" || cause == nil {
		return nil, fmt.Errorf("%w: telemetry failure", domain.ErrInvalid)
	}
	return &Failure{
		Code:    code,
		Message: message,
		Cause:   cause,
		Labels:  labels,
		Path:    path,
	}, nil
}

func (f *Failure) Error() string {
	if f == nil {
		return "telemetry failure"
	}
	if len(f.Path) == 0 {
		return f.Code + ": " + f.Message
	}
	return f.Code + ": " + f.Message + " at " + strings.Join(f.Path, "/")
}

func (f *Failure) Unwrap() error {
	if f == nil {
		return nil
	}
	return f.Cause
}

func (f *Failure) Clone() *Failure {
	if f == nil {
		return nil
	}
	return &Failure{
		Code:    f.Code,
		Message: f.Message,
		Cause:   f.Cause,
		Labels:  f.Labels,
		Path:    f.Path,
	}
}

func (f *Failure) AddContext(label, value, segment string) {
	if f == nil {
		return
	}
	if f.Labels == nil {
		f.Labels = make(map[string]string)
	}
	if label != "" {
		f.Labels[label] = value
	}
	if segment != "" {
		f.Path = append(f.Path, segment)
	}
}

func (f *Failure) Retryable() bool {
	return f != nil && (errors.Is(f.Cause, contextDeadlineError) || errors.Is(f.Cause, domain.ErrConflict))
}

func (f *Failure) SortedLabels() []string {
	if f == nil {
		return nil
	}
	keys := make([]string, 0, len(f.Labels))
	for key := range f.Labels {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

var contextDeadlineError = errors.New("telemetry downstream deadline")
