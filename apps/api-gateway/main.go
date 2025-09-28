package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ServiceConfig represents configuration for a microservice
type ServiceConfig struct {
	Name string
	URL  string
}

// APIGateway handles routing requests to microservices
type APIGateway struct {
	Services map[string]ServiceConfig
	Client   *http.Client
}

// NewAPIGateway creates a new API Gateway instance
func NewAPIGateway() *APIGateway {
	return &APIGateway{
		Services: map[string]ServiceConfig{
			"telemetry": {
				Name: "telemetry-service",
				URL:  getEnv("TELEMETRY_SERVICE_URL", "http://telemetry-service:8082"),
			},
			"device": {
				Name: "device-service",
				URL:  getEnv("DEVICE_SERVICE_URL", "http://device-service:8083"),
			},
			"temperature": {
				Name: "temperature-api",
				URL:  getEnv("TEMPERATURE_API_URL", "http://temperature-api:8081"),
			},
		},
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ProxyRequest proxies a request to the appropriate microservice
func (gw *APIGateway) ProxyRequest(c *gin.Context) {
	// Extract service name from the URL path
	path := c.Param("path")
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	if len(pathParts) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invalid path"})
		return
	}

	// Determine service name based on the route that matched
	var serviceName string
	if strings.HasPrefix(c.Request.URL.Path, "/api/temperature") {
		serviceName = "temperature"
	} else if strings.HasPrefix(c.Request.URL.Path, "/api/device") {
		serviceName = "device"
	} else if strings.HasPrefix(c.Request.URL.Path, "/api/telemetry") {
		serviceName = "telemetry"
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
		return
	}

	service, exists := gw.Services[serviceName]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
		return
	}

	// Build target URL
	targetURL, err := url.Parse(service.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid service URL"})
		return
	}

	// Create reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// Modify the request
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.URL.Scheme = targetURL.Scheme
		req.URL.Host = targetURL.Host
		// Handle different service path patterns
		if serviceName == "temperature" {
			// Temperature service expects /temperature/:id
			req.URL.Path = "/" + serviceName + "/" + strings.TrimPrefix(path, "/")
		} else {
			// Other services expect /{serviceName} directly
			if path == "" || path == "/" {
				req.URL.Path = "/" + serviceName
			} else {
				req.URL.Path = "/" + serviceName + "/" + strings.TrimPrefix(path, "/")
			}
		}
		req.Header.Set("X-Forwarded-For", c.ClientIP())
		req.Header.Set("X-Forwarded-Proto", c.Request.Header.Get("X-Forwarded-Proto"))
	}

	// Add error handling
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Proxy error for service %s: %v", serviceName, err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "Service unavailable"})
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}

// HealthCheck returns the health status of the API Gateway
func (gw *APIGateway) HealthCheck(c *gin.Context) {
	// Check health of all services
	servicesStatus := make(map[string]string)

	for name, service := range gw.Services {
		status := "healthy"
		resp, err := gw.Client.Get(service.URL + "/health")
		if err != nil || resp.StatusCode != http.StatusOK {
			status = "unhealthy"
		}
		if resp != nil {
			resp.Body.Close()
		}
		servicesStatus[name] = status
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "ok",
		"services": servicesStatus,
	})
}

// AggregateTelemetry aggregates telemetry data from multiple sources
func (gw *APIGateway) AggregateTelemetry(c *gin.Context) {
	// Get temperature data from temperature service
	tempResp, err := gw.Client.Get(gw.Services["temperature"].URL + "/temperature?location=" + c.Query("location"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch temperature data"})
		return
	}
	defer tempResp.Body.Close()

	var tempData map[string]interface{}
	if err := json.NewDecoder(tempResp.Body).Decode(&tempData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode temperature data"})
		return
	}

	// Get telemetry data from telemetry service
	telemetryResp, err := gw.Client.Get(gw.Services["telemetry"].URL + "/telemetry?device_id=" + c.Query("device_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch telemetry data"})
		return
	}
	defer telemetryResp.Body.Close()

	var telemetryData []map[string]interface{}
	if err := json.NewDecoder(telemetryResp.Body).Decode(&telemetryData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode telemetry data"})
		return
	}

	// Aggregate data
	aggregatedData := gin.H{
		"temperature": tempData,
		"telemetry":   telemetryData,
		"timestamp":   time.Now(),
	}

	c.JSON(http.StatusOK, aggregatedData)
}

func main() {
	// Initialize API Gateway
	gateway := NewAPIGateway()

	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check endpoint
	router.GET("/health", gateway.HealthCheck)

	// API routes
	api := router.Group("/api")
	{
		// Proxy routes for microservices
		api.Any("/telemetry/*path", gateway.ProxyRequest)
		api.Any("/device/*path", gateway.ProxyRequest)
		api.Any("/temperature/*path", gateway.ProxyRequest)

		// Aggregated endpoints
		api.GET("/aggregate/telemetry", gateway.AggregateTelemetry)
	}

	// Start server
	port := getEnv("PORT", ":8084")
	log.Printf("API Gateway starting on %s", port)
	if err := router.Run(port); err != nil {
		log.Fatalf("Failed to start API Gateway: %v", err)
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
