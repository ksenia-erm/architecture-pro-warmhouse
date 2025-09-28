package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

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
	DeviceID      string    `json:"device_id" binding:"required"`
	MetricsNames  []string  `json:"metrics_names" binding:"required"`
	MetricsValues []float64 `json:"metrics_values" binding:"required"`
}

// TelemetryService handles telemetry data operations
type TelemetryService struct {
	Records map[string]TelemetryRecord
}

// NewTelemetryService creates a new telemetry service
func NewTelemetryService() *TelemetryService {
	return &TelemetryService{
		Records: make(map[string]TelemetryRecord),
	}
}

// CreateTelemetryRecord creates a new telemetry record
func (s *TelemetryService) CreateTelemetryRecord(record TelemetryRecordCreate) TelemetryRecord {
	id := uuid.New().String()
	telemetryRecord := TelemetryRecord{
		ID:            id,
		DeviceID:      record.DeviceID,
		MetricsNames:  record.MetricsNames,
		MetricsValues: record.MetricsValues,
		CreatedAt:     time.Now(),
	}

	s.Records[id] = telemetryRecord
	return telemetryRecord
}

// GetTelemetryRecords retrieves telemetry records with optional filtering
func (s *TelemetryService) GetTelemetryRecords(deviceID, from, to string) []TelemetryRecord {
	var records []TelemetryRecord

	for _, record := range s.Records {
		// Filter by device ID if provided
		if deviceID != "" && record.DeviceID != deviceID {
			continue
		}

		// Filter by time range if provided
		if from != "" {
			if fromTime, err := time.Parse(time.RFC3339, from); err == nil {
				if record.CreatedAt.Before(fromTime) {
					continue
				}
			}
		}

		if to != "" {
			if toTime, err := time.Parse(time.RFC3339, to); err == nil {
				if record.CreatedAt.After(toTime) {
					continue
				}
			}
		}

		records = append(records, record)
	}

	return records
}

// GetTelemetryRecordByID retrieves a telemetry record by ID
func (s *TelemetryService) GetTelemetryRecordByID(id string) (TelemetryRecord, bool) {
	record, exists := s.Records[id]
	return record, exists
}

// CreateBulkTelemetryRecords creates multiple telemetry records
func (s *TelemetryService) CreateBulkTelemetryRecords(records []TelemetryRecordCreate) []TelemetryRecord {
	var createdRecords []TelemetryRecord

	for _, record := range records {
		createdRecord := s.CreateTelemetryRecord(record)
		createdRecords = append(createdRecords, createdRecord)
	}

	return createdRecords
}

// GenerateSampleData generates sample telemetry data for testing
func (s *TelemetryService) GenerateSampleData() {
	// Generate some sample data
	devices := []string{"device-001", "device-002", "device-003"}
	metrics := []string{"temperature", "humidity", "pressure", "light_level"}

	for i := 0; i < 10; i++ {
		deviceID := devices[i%len(devices)]
		values := []float64{
			20.0 + float64(i%10),
			45.0 + float64(i%20),
			1013.0 + float64(i%5),
			300.0 + float64(i%100),
		}

		record := TelemetryRecordCreate{
			DeviceID:      deviceID,
			MetricsNames:  metrics,
			MetricsValues: values,
		}

		s.CreateTelemetryRecord(record)
	}
}

func main() {
	// Initialize telemetry service
	telemetryService := NewTelemetryService()

	// Generate sample data
	telemetryService.GenerateSampleData()

	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "telemetry-service",
		})
	})

	// API routes
	api := router.Group("/telemetry")
	{
		// Get telemetry records with optional filtering
		api.GET("", func(c *gin.Context) {
			deviceID := c.Query("device_id")
			from := c.Query("from")
			to := c.Query("to")

			records := telemetryService.GetTelemetryRecords(deviceID, from, to)
			c.JSON(http.StatusOK, records)
		})

		// Create single telemetry record
		api.POST("", func(c *gin.Context) {
			var record TelemetryRecordCreate
			if err := c.ShouldBindJSON(&record); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			createdRecord := telemetryService.CreateTelemetryRecord(record)
			c.JSON(http.StatusCreated, createdRecord)
		})

		// Create bulk telemetry records
		api.POST("/bulk", func(c *gin.Context) {
			var records []TelemetryRecordCreate
			if err := c.ShouldBindJSON(&records); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			createdRecords := telemetryService.CreateBulkTelemetryRecords(records)
			c.JSON(http.StatusCreated, createdRecords)
		})

		// Get telemetry record by ID
		api.GET("/:id", func(c *gin.Context) {
			id := c.Param("id")
			record, exists := telemetryService.GetTelemetryRecordByID(id)
			if !exists {
				c.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
				return
			}

			c.JSON(http.StatusOK, record)
		})
	}

	// Start server
	port := getEnv("PORT", ":8082")
	log.Printf("Telemetry Service starting on %s", port)
	if err := router.Run(port); err != nil {
		log.Fatalf("Failed to start Telemetry Service: %v", err)
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
