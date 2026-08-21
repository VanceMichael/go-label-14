package telemetry

import (
	"fmt"
	"sort"

	"go-base/internal/domain"
)

type DeviceResult struct {
	SensorID string
	Accepted int
	Failure  *Failure
}

type BatchResult struct {
	BatchID  string
	TenantID string
	Devices  []DeviceResult
}

func NewBatchResult(batchID, tenantID string, devices []DeviceResult) (BatchResult, error) {
	if batchID == "" || tenantID == "" {
		return BatchResult{}, fmt.Errorf("%w: telemetry batch identity", domain.ErrInvalid)
	}
	seen := make(map[string]struct{}, len(devices))
	copyDevices := make([]DeviceResult, len(devices))
	for index, device := range devices {
		if device.SensorID == "" || device.Accepted < 0 {
			return BatchResult{}, fmt.Errorf("%w: telemetry device result", domain.ErrInvalid)
		}
		if _, exists := seen[device.SensorID]; exists {
			return BatchResult{}, fmt.Errorf("%w: duplicate sensor result", domain.ErrConflict)
		}
		seen[device.SensorID] = struct{}{}
		copyDevices[index] = DeviceResult{
			SensorID: device.SensorID,
			Accepted: device.Accepted,
			Failure:  device.Failure.Clone(),
		}
	}
	sort.Slice(copyDevices, func(i, j int) bool { return copyDevices[i].SensorID < copyDevices[j].SensorID })
	return BatchResult{BatchID: batchID, TenantID: tenantID, Devices: copyDevices}, nil
}

func (r BatchResult) Snapshot() BatchResult {
	devices := make([]DeviceResult, len(r.Devices))
	for index, device := range r.Devices {
		devices[index] = DeviceResult{
			SensorID: device.SensorID,
			Accepted: device.Accepted,
			Failure:  device.Failure.Clone(),
		}
	}
	return BatchResult{BatchID: r.BatchID, TenantID: r.TenantID, Devices: devices}
}

func (r *BatchResult) AnnotateFailure(sensorID, station, attempt string) error {
	if r == nil || sensorID == "" {
		return domain.ErrInvalid
	}
	for index := range r.Devices {
		device := &r.Devices[index]
		if device.SensorID != sensorID {
			continue
		}
		if device.Failure == nil {
			return domain.ErrConflict
		}
		device.Failure.AddContext("station", station, "attempt-"+attempt)
		return nil
	}
	return domain.ErrNotFound
}
