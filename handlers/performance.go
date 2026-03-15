package handlers

import (
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"

	"lumina-blog/cache"
	"lumina-blog/config"
	"lumina-blog/middleware"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// PerformanceStats holds performance monitoring data
type PerformanceStats struct {
	QPS            int64   `json:"qps"`
	Uptime         string  `json:"uptime"`
	CacheHitRate   float64 `json:"cache_hit_rate"`
	CacheSize      int     `json:"cache_size"`
	TotalHits      int64   `json:"total_cache_hits"`
	TotalMisses    int64   `json:"total_cache_misses"`
	MemoryUsage    uint64  `json:"memory_usage_mb"`
	Goroutines     int     `json:"goroutines"`
	DBOpenConns    int     `json:"db_open_connections"`
	DBIdleConns    int     `json:"db_idle_connections"`
	DBInUseConns   int     `json:"db_in_use_connections"`
	AvgResponseMs  float64 `json:"avg_response_time_ms"`
}

// ResponseTimeStats tracks response times
var (
	responseTimes  []time.Duration
	responseMutex  sync.Mutex
	startTime     = time.Now()
)

func init() {
	// Initialize response time tracking
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			responseMutex.Lock()
			if len(responseTimes) > 1000 {
				responseTimes = responseTimes[len(responseTimes)-500:]
			}
			responseMutex.Unlock()
		}
	}()
}

// TrackResponseTime tracks the response time for averaging
func TrackResponseTime(duration time.Duration) {
	responseMutex.Lock()
	defer responseMutex.Unlock()
	responseTimes = append(responseTimes, duration)
	if len(responseTimes) > 1000 {
		responseTimes = responseTimes[1:]
	}
}

// GetAverageResponseTime returns the average response time in milliseconds
func GetAverageResponseTime() float64 {
	responseMutex.Lock()
	defer responseMutex.Unlock()
	
	if len(responseTimes) == 0 {
		return 0
	}
	
	var total time.Duration
	for _, t := range responseTimes {
		total += t
	}
	return float64(total.Milliseconds()) / float64(len(responseTimes))
}

// GetPerformanceStats returns current performance metrics
func GetPerformanceStats(c *gin.Context) {
	middleware.IncrementRequestCount()
	start := time.Now()
	defer func() {
		TrackResponseTime(time.Since(start))
	}()

	// Get cache stats
	hits, misses, size := cache.ArticleCache.Stats()
	
	// Get memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	memUsageMB := memStats.Alloc / 1024 / 1024

	// Get DB stats
	dbStats := config.GetDBStats()

	// Calculate uptime
	uptime := time.Since(startTime)

	// Get QPS
	qps := middleware.GetQPS()

	// Get average response time
	avgResponseMs := GetAverageResponseTime()

	stats := PerformanceStats{
		QPS:            qps,
		Uptime:         uptime.Round(time.Second).String(),
		CacheHitRate:   cache.ArticleCache.HitRate(),
		CacheSize:      size,
		TotalHits:      hits,
		TotalMisses:    misses,
		MemoryUsage:    memUsageMB,
		Goroutines:     runtime.NumGoroutine(),
		DBOpenConns:    dbStats.OpenConnections,
		DBIdleConns:    dbStats.Idle,
		DBInUseConns:   dbStats.InUse,
		AvgResponseMs:  avgResponseMs,
	}

	c.JSON(http.StatusOK, stats)
}

// ClearCache clears the article cache
func ClearCache(c *gin.Context) {
	middleware.IncrementRequestCount()
	
	// Clear caches
	cache.ArticleCache.Clear()
	cache.DefaultCache.Clear()
	
	// Reset performance counters
	middleware.ResetQPS()
	
	response := gin.H{
		"message":       "Cache cleared successfully",
		"cleared_items": "all",
		"timestamp":     time.Now().Format(time.RFC3339),
	}
	
	c.JSON(http.StatusOK, response)
}

// ImageProcessRequest represents a request for image processing
type ImageProcessRequest struct {
	URL         string `json:"url" binding:"required"`
	Format      string `json:"format"` // webp, jpeg, png
	Quality     int    `json:"quality"` // 1-100
	MaxWidth    int    `json:"max_width"`
	MaxHeight   int    `json:"max_height"`
}

// ImageProcessResult represents the result of image processing
type ImageProcessResult struct {
	OriginalURL   string `json:"original_url"`
	ProcessedURL  string `json:"processed_url"`
	Format        string `json:"format"`
	Quality       int    `json:"quality"`
	WebPURL       string `json:"webp_url,omitempty"`
	WebPAvailable bool   `json:"webp_available"`
	Note          string `json:"note"`
}

// ProcessImage processes an image (compression hints and WebP conversion info)
func ProcessImage(c *gin.Context) {
	middleware.IncrementRequestCount()
	start := time.Now()
	defer func() {
		TrackResponseTime(time.Since(start))
	}()

	// Load .env to check for image service configuration
	_ = godotenv.Load()

	var req ImageProcessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL is required"})
		return
	}

	// Set defaults
	if req.Quality == 0 {
		req.Quality = 80
	}
	if req.Format == "" {
		req.Format = "webp"
	}

	// Get image service configuration from environment
	// In production, you might integrate with Cloudinary, Imgix, or similar services
	imageServiceURL := ""
	imageServiceEnabled := false
	
	// Check environment for image service configuration
	// Examples:
	// IMAGE_SERVICE=cloudinary
	// IMAGE_CLOUDINARY_CLOUD_NAME=xxx
	// IMAGE_CLOUDINARY_API_KEY=xxx
	// IMAGE_CLOUDINARY_API_SECRET=xxx
	
	// For now, return hints and info about WebP conversion
	result := ImageProcessResult{
		OriginalURL:   req.URL,
		Format:        req.Format,
		Quality:       req.Quality,
		WebPURL:       "",
		WebPAvailable: imageServiceEnabled,
		Note:          getImageProcessingNote(req, imageServiceEnabled),
	}

	// If image service is configured, generate the processed URL
	if imageServiceEnabled {
		result.ProcessedURL = generateProcessedImageURL(req, imageServiceURL)
		result.WebPURL = generateWebPURL(req, imageServiceURL)
		result.WebPAvailable = true
	} else {
		// Return suggestions without actual processing
		result.ProcessedURL = req.URL
		result.WebPURL = getWebPSuggestion(req.URL)
	}

	c.JSON(http.StatusOK, result)
}

func getImageProcessingNote(req ImageProcessRequest, serviceEnabled bool) string {
	if serviceEnabled {
		return "Image processed successfully using external image service"
	}
	
	switch req.Format {
	case "webp":
		return "WebP conversion is available. Configure an image service (Cloudinary, Imgix, Sharp, etc.) in environment variables to enable automatic conversion."
	case "jpeg":
		return "JPEG optimization is available. Configure an image service in environment variables to enable automatic compression."
	default:
		return "Configure an image service in environment variables to enable image processing. Suggested services: Cloudinary, Imgix, or local Sharp processing."
	}
}

func generateProcessedImageURL(req ImageProcessRequest, serviceBase string) string {
	// Placeholder for actual service integration
	return serviceBase + "/process?url=" + req.URL + "&format=" + req.Format
}

func generateWebPURL(req ImageProcessRequest, serviceBase string) string {
	// Placeholder for actual service integration
	return serviceBase + "/process?url=" + req.URL + "&format=webp"
}

func getWebPSuggestion(originalURL string) string {
	// Suggest WebP URL transformation
	// This is a common pattern used by many image CDNs
	parts := strings.Split(originalURL, ".")
	if len(parts) >= 2 {
		// Insert webp before extension
		baseURL := strings.Join(parts[:len(parts)-1], ".")
		
		// Common patterns:
		// - Replace extension: image.jpg -> image.webp
		// - Add query param: image.jpg?format=webp
		// - Path prefix: /images/webp/image.jpg
		
		return baseURL + ".webp"
	}
	return originalURL + "?format=webp"
}

// HealthCheck returns a simple health check
func HealthCheck(c *gin.Context) {
	middleware.IncrementRequestCount()
	
	// Check database connection
	db, err := config.DB.DB()
	dbHealthy := err == nil && db.Ping() == nil
	
	// Check cache
	_, _, cacheSize := cache.ArticleCache.Stats()
	
	// Memory usage
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	status := "healthy"
	httpStatus := http.StatusOK
	
	if !dbHealthy {
		status = "unhealthy"
		httpStatus = http.StatusServiceUnavailable
	}
	
	c.JSON(httpStatus, gin.H{
		"status":         status,
		"timestamp":     time.Now().Format(time.RFC3339),
		"database":      dbHealthy,
		"cache_items":   cacheSize,
		"memory_mb":     memStats.Alloc / 1024 / 1024,
		"goroutines":    runtime.NumGoroutine(),
	})
}
