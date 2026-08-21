package telemetry

import (
	"errors"
	"testing"

	"go-base/internal/domain"
)

func TestAnnotatingOneDeviceFailureDoesNotPolluteBatchPeersOrSnapshots(t *testing.T) {
	base, err := NewFailure(
		"sequence_conflict",
		"reading sequence already committed",
		fmtConflict(),
		map[string]string{"barn": "north"},
		[]string{"ingest"},
	)
	if err != nil {
		t.Fatalf("new failure: %v", err)
	}
	result, err := NewBatchResult("batch-14", "farm-east", []DeviceResult{
		{SensorID: "sensor-a", Failure: base},
		{SensorID: "sensor-b", Failure: base},
	})
	if err != nil {
		t.Fatalf("new batch result: %v", err)
	}
	retained := result.Snapshot()

	if err := result.AnnotateFailure("sensor-a", "gateway-7", "2"); err != nil {
		t.Fatalf("annotate: %v", err)
	}
	first := result.Devices[0].Failure
	second := result.Devices[1].Failure
	if first.Labels["station"] != "gateway-7" || first.Path[len(first.Path)-1] != "attempt-2" {
		t.Fatalf("selected failure was not annotated: %+v", first)
	}
	if _, polluted := second.Labels["station"]; polluted || len(second.Path) != 1 {
		t.Fatalf("peer failure was polluted: %+v", second)
	}
	if _, polluted := retained.Devices[0].Failure.Labels["station"]; polluted || len(retained.Devices[0].Failure.Path) != 1 {
		t.Fatalf("retained snapshot was polluted: %+v", retained.Devices[0].Failure)
	}
	if _, polluted := base.Labels["station"]; polluted || len(base.Path) != 1 {
		t.Fatalf("caller's original failure was polluted: %+v", base)
	}
	for _, failure := range []*Failure{first, second, retained.Devices[0].Failure, base} {
		if !errors.Is(failure, domain.ErrConflict) || !failure.Retryable() {
			t.Fatalf("failure lost error classification: %v", failure)
		}
	}
}

func fmtConflict() error {
	return errors.Join(contextDeadlineError, domain.ErrConflict)
}
