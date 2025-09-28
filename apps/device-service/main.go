package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

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
	SerialNumber     string `json:"serial_number" binding:"required"`
	ConnectionString string `json:"connection_string"`
	Description      string `json:"description"`
	DeviceTypeID     int    `json:"device_type_id" binding:"required"`
	RoomID           string `json:"room_id"`
}

// DeviceUpdate represents the data that can be updated for a device
type DeviceUpdate struct {
	ConnectionString string `json:"connection_string"`
	Description      string `json:"description"`
	RoomID           string `json:"room_id"`
	Status           string `json:"status"`
}

// DeviceCommand represents a command to be sent to a device
type DeviceCommand struct {
	Command    string                 `json:"command" binding:"required"`
	Parameters map[string]interface{} `json:"parameters"`
}

// DeviceService handles device operations
type DeviceService struct {
	Devices map[string]Device
}

// NewDeviceService creates a new device service
func NewDeviceService() *DeviceService {
	return &DeviceService{
		Devices: make(map[string]Device),
	}
}

// CreateDevice creates a new device
func (s *DeviceService) CreateDevice(device DeviceCreate) Device {
	id := uuid.New().String()
	now := time.Now()

	deviceRecord := Device{
		ID:               id,
		SerialNumber:     device.SerialNumber,
		ConnectionString: device.ConnectionString,
		Description:      device.Description,
		DeviceTypeID:     device.DeviceTypeID,
		Status:           "offline",
		RoomID:           device.RoomID,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	s.Devices[id] = deviceRecord
	return deviceRecord
}

// GetDevices retrieves devices with optional filtering
func (s *DeviceService) GetDevices(houseID, status string) []Device {
	var devices []Device

	for _, device := range s.Devices {
		// Filter by house ID if provided (simplified - using room_id as house_id)
		if houseID != "" && device.RoomID != houseID {
			continue
		}

		// Filter by status if provided
		if status != "" && device.Status != status {
			continue
		}

		devices = append(devices, device)
	}

	return devices
}

// GetDeviceByID retrieves a device by ID
func (s *DeviceService) GetDeviceByID(id string) (Device, bool) {
	device, exists := s.Devices[id]
	return device, exists
}

// UpdateDevice updates an existing device
func (s *DeviceService) UpdateDevice(id string, update DeviceUpdate) (Device, bool) {
	device, exists := s.Devices[id]
	if !exists {
		return Device{}, false
	}

	// Update fields if provided
	if update.ConnectionString != "" {
		device.ConnectionString = update.ConnectionString
	}
	if update.Description != "" {
		device.Description = update.Description
	}
	if update.RoomID != "" {
		device.RoomID = update.RoomID
	}
	if update.Status != "" {
		device.Status = update.Status
	}

	device.UpdatedAt = time.Now()
	s.Devices[id] = device

	return device, true
}

// UpdateDeviceStatus updates the status of a device
func (s *DeviceService) UpdateDeviceStatus(id, status string) (Device, bool) {
	device, exists := s.Devices[id]
	if !exists {
		return Device{}, false
	}

	device.Status = status
	device.UpdatedAt = time.Now()
	s.Devices[id] = device

	return device, true
}

// SendCommand sends a command to a device
func (s *DeviceService) SendCommand(id string, command DeviceCommand) bool {
	device, exists := s.Devices[id]
	if !exists {
		return false
	}

	// Simulate command processing
	log.Printf("Sending command %s to device %s (serial: %s)", command.Command, id, device.SerialNumber)

	// Update device status to indicate command processing
	device.Status = "processing"
	device.UpdatedAt = time.Now()
	s.Devices[id] = device

	// Simulate async processing
	go func() {
		time.Sleep(2 * time.Second)
		device.Status = "online"
		device.UpdatedAt = time.Now()
		s.Devices[id] = device
		log.Printf("Command %s completed for device %s", command.Command, id)
	}()

	return true
}

// GenerateSampleData generates sample device data for testing
func (s *DeviceService) GenerateSampleData() {
	// Generate some sample devices
	devices := []DeviceCreate{
		{
			SerialNumber:     "SN-1001",
			ConnectionString: "mqtt://broker:1883",
			Description:      "Living Room Temperature Sensor",
			DeviceTypeID:     1,
			RoomID:           "room-001",
		},
		{
			SerialNumber:     "SN-1002",
			ConnectionString: "mqtt://broker:1883",
			Description:      "Bedroom Light Controller",
			DeviceTypeID:     2,
			RoomID:           "room-002",
		},
		{
			SerialNumber:     "SN-1003",
			ConnectionString: "mqtt://broker:1883",
			Description:      "Kitchen Humidity Sensor",
			DeviceTypeID:     1,
			RoomID:           "room-003",
		},
	}

	for _, device := range devices {
		s.CreateDevice(device)
	}
}

func main() {
	// Initialize device service
	deviceService := NewDeviceService()

	// Generate sample data
	deviceService.GenerateSampleData()

	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "device-service",
		})
	})

	// API routes
	api := router.Group("/devices")
	{
		// Get devices with optional filtering
		api.GET("", func(c *gin.Context) {
			houseID := c.Query("house_id")
			status := c.Query("status")

			devices := deviceService.GetDevices(houseID, status)
			c.JSON(http.StatusOK, devices)
		})

		// Create device
		api.POST("", func(c *gin.Context) {
			var device DeviceCreate
			if err := c.ShouldBindJSON(&device); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			createdDevice := deviceService.CreateDevice(device)
			c.JSON(http.StatusCreated, createdDevice)
		})

		// Get device by ID
		api.GET("/:id", func(c *gin.Context) {
			id := c.Param("id")
			device, exists := deviceService.GetDeviceByID(id)
			if !exists {
				c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
				return
			}

			c.JSON(http.StatusOK, device)
		})

		// Update device
		api.PATCH("/:id", func(c *gin.Context) {
			id := c.Param("id")
			var update DeviceUpdate
			if err := c.ShouldBindJSON(&update); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			device, exists := deviceService.UpdateDevice(id, update)
			if !exists {
				c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
				return
			}

			c.JSON(http.StatusOK, device)
		})

		// Update device status
		api.PATCH("/:id/status", func(c *gin.Context) {
			id := c.Param("id")
			var request struct {
				Status string `json:"status" binding:"required"`
			}

			if err := c.ShouldBindJSON(&request); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			device, exists := deviceService.UpdateDeviceStatus(id, request.Status)
			if !exists {
				c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
				return
			}

			c.JSON(http.StatusOK, device)
		})

		// Send command to device
		api.POST("/:id/commands", func(c *gin.Context) {
			id := c.Param("id")
			var command DeviceCommand
			if err := c.ShouldBindJSON(&command); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			success := deviceService.SendCommand(id, command)
			if !success {
				c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
				return
			}

			c.JSON(http.StatusAccepted, gin.H{"message": "Command accepted"})
		})
	}

	// Start server
	port := getEnv("PORT", ":8083")
	log.Printf("Device Service starting on %s", port)
	if err := router.Run(port); err != nil {
		log.Fatalf("Failed to start Device Service: %v", err)
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
