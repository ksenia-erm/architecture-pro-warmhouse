package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// TelemetryService handles telemetry data operations
type TelemetryService struct {
	BaseURL    string
	HTTPClient *http.Client
}

// TelemetryRecord represents a telemetry data record
type TelemetryRecord struct {
	ID            string    `json:"id"`
	DeviceID      string    `json:"device_id"`
	MetricsNames  []string  `json:"metrics_names"`
	MetricsValues []float64 `json:"metrics_values"`
	CreatedAt     time.Time `json:"created_at"`
}

// TelemetryRecordCreate represents the data needed to create a new telemetry record
type TelemetryRecordCreate struct {
	DeviceID      string    `json:"device_id"`
	MetricsNames  []string  `json:"metrics_names"`
	MetricsValues []float64 `json:"metrics_values"`
}

// NewTelemetryService creates a new telemetry service
func NewTelemetryService(baseURL string) *TelemetryService {
	return &TelemetryService{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetTelemetryRecords retrieves telemetry records with optional filtering
func (s *TelemetryService) GetTelemetryRecords(deviceID, from, to string) ([]TelemetryRecord, error) {
	url := fmt.Sprintf("%s/telemetry", s.BaseURL)
	params := make([]string, 0)

	if deviceID != "" {
		params = append(params, fmt.Sprintf("device_id=%s", deviceID))
	}
	if from != "" {
		params = append(params, fmt.Sprintf("from=%s", from))
	}
	if to != "" {
		params = append(params, fmt.Sprintf("to=%s", to))
	}

	if len(params) > 0 {
		url += "?" + params[0]
		for i := 1; i < len(params); i++ {
			url += "&" + params[i]
		}
	}

	resp, err := s.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching telemetry records: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var records []TelemetryRecord
	if err := json.NewDecoder(resp.Body).Decode(&records); err != nil {
		return nil, fmt.Errorf("error decoding telemetry records: %w", err)
	}

	return records, nil
}

// CreateTelemetryRecord creates a new telemetry record
func (s *TelemetryService) CreateTelemetryRecord(record TelemetryRecordCreate) (*TelemetryRecord, error) {
	url := fmt.Sprintf("%s/telemetry", s.BaseURL)

	jsonData, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("error marshaling telemetry record: %w", err)
	}

	resp, err := s.HTTPClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating telemetry record: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var createdRecord TelemetryRecord
	if err := json.NewDecoder(resp.Body).Decode(&createdRecord); err != nil {
		return nil, fmt.Errorf("error decoding created telemetry record: %w", err)
	}

	return &createdRecord, nil
}

// CreateBulkTelemetryRecords creates multiple telemetry records
func (s *TelemetryService) CreateBulkTelemetryRecords(records []TelemetryRecordCreate) ([]TelemetryRecord, error) {
	url := fmt.Sprintf("%s/telemetry/bulk", s.BaseURL)

	jsonData, err := json.Marshal(records)
	if err != nil {
		return nil, fmt.Errorf("error marshaling telemetry records: %w", err)
	}

	resp, err := s.HTTPClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating bulk telemetry records: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var createdRecords []TelemetryRecord
	if err := json.NewDecoder(resp.Body).Decode(&createdRecords); err != nil {
		return nil, fmt.Errorf("error decoding created telemetry records: %w", err)
	}

	return createdRecords, nil
}
