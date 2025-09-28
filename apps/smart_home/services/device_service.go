package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// DeviceService handles device management operations
type DeviceService struct {
	BaseURL    string
	HTTPClient *http.Client
}

// Device represents a smart device
type Device struct {
	ID               string    `json:"id"`
	SerialNumber     string    `json:"serial_number"`
	ConnectionString string    `json:"connection_string"`
	Description      string    `json:"description"`
	DeviceTypeID     int       `json:"device_type_id"`
	Status           string    `json:"status"`
	RoomID           string    `json:"room_id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// DeviceCreate represents the data needed to create a new device
type DeviceCreate struct {
	SerialNumber     string `json:"serial_number"`
	ConnectionString string `json:"connection_string"`
	Description      string `json:"description"`
	DeviceTypeID     int    `json:"device_type_id"`
	RoomID           string `json:"room_id"`
}

// DeviceCommand represents a command to be sent to a device
type DeviceCommand struct {
	Command    string                 `json:"command"`
	Parameters map[string]interface{} `json:"parameters"`
}

// NewDeviceService creates a new device service
func NewDeviceService(baseURL string) *DeviceService {
	return &DeviceService{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetDevices retrieves devices with optional filtering
func (s *DeviceService) GetDevices(houseID, status string) ([]Device, error) {
	url := fmt.Sprintf("%s/devices", s.BaseURL)
	if houseID != "" {
		url += fmt.Sprintf("?house_id=%s", houseID)
	}
	if status != "" {
		if houseID != "" {
			url += "&"
		} else {
			url += "?"
		}
		url += fmt.Sprintf("status=%s", status)
	}

	resp, err := s.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching devices: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var devices []Device
	if err := json.NewDecoder(resp.Body).Decode(&devices); err != nil {
		return nil, fmt.Errorf("error decoding devices: %w", err)
	}

	return devices, nil
}

// CreateDevice creates a new device
func (s *DeviceService) CreateDevice(device DeviceCreate) (*Device, error) {
	url := fmt.Sprintf("%s/devices", s.BaseURL)

	jsonData, err := json.Marshal(device)
	if err != nil {
		return nil, fmt.Errorf("error marshaling device: %w", err)
	}

	resp, err := s.HTTPClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating device: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var createdDevice Device
	if err := json.NewDecoder(resp.Body).Decode(&createdDevice); err != nil {
		return nil, fmt.Errorf("error decoding created device: %w", err)
	}

	return &createdDevice, nil
}

// SendCommand sends a command to a device
func (s *DeviceService) SendCommand(deviceID string, command DeviceCommand) error {
	url := fmt.Sprintf("%s/devices/%s/commands", s.BaseURL, deviceID)

	jsonData, err := json.Marshal(command)
	if err != nil {
		return fmt.Errorf("error marshaling command: %w", err)
	}

	resp, err := s.HTTPClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error sending command: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
