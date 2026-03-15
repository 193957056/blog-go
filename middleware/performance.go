package middleware

import (
	"context"
	"log"
	"strconv"
	"sync"
	"time"

	"lumina-blog/cache"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SlowQueryThreshold is the threshold in milliseconds for logging slow queries
const SlowQueryThreshold = 500 // ms

// QueryLogger is a custom GORM logger that logs slow queries
type QueryLogger struct {
	SlowThreshold time.Duration
}

// LogMode sets the log mode
func (l *QueryLogger) LogMode(level logger.LogLevel) logger.Interface {
	return l
}

// Info logs info messages
func (l *QueryLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	log.Printf("[GORM] "+msg, data...)
}

// Warn logs warning messages
func (l *QueryLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	log.Printf("[GORM WARN] "+msg, data...)
}

// Error logs error messages
func (l *QueryLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	log.Printf("[GORM ERROR] "+msg, data...)
}

// Trace traces SQL execution
func (l *QueryLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	
	// Log slow queries
	if elapsed.Milliseconds() > SlowQueryThreshold {
		log.Printf("[SLOW QUERY] duration=%dms rows=%d sql=%s", 
			elapsed.Milliseconds(), rows, truncateSQL(sql, 500))
	}
	
	// Log all queries in debug mode
	if gin.Mode() == gin.DebugMode {
		log.Printf("[QUERY] duration=%dms rows=%d sql=%s", 
			elapsed.Milliseconds(), rows, truncateSQL(sql, 300))
	}
}

func truncateSQL(sql string, maxLen int) string {
	if len(sql) > maxLen {
		return sql[:maxLen] + "..."
	}
	return sql
}

// PerformanceMiddleware tracks request performance metrics
func PerformanceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		c.Next()
		
		duration := time.Since(start)
		durationMs := duration.Milliseconds()
		
		// Log slow requests (> 500ms)
		if durationMs > 500 {
			log.Printf("[SLOW REQUEST] method=%s path=%s duration=%dms status=%d",
				c.Request.Method, c.Request.URL.Path, durationMs, c.Writer.Status())
		}
		
		// Add performance headers in debug mode
		if gin.Mode() == gin.DebugMode {
			c.Header("X-Response-Time-Ms", strconv.FormatInt(durationMs, 10))
		}
	}
}

// CacheMiddleware adds caching to GET requests
func CacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only cache GET requests
		if c.Request.Method != "GET" {
			c.Next()
			return
		}
		
		// Skip if user is authenticated (for admin endpoints)
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			c.Next()
			return
		}
		
		// Build cache key from path and query
		cacheKey := c.Request.URL.Path
		if query := c.Request.URL.RawQuery; query != "" {
			cacheKey += "?" + query
		}
		
		// Check cache
		if value, found := cache.ArticleCache.Get(cacheKey); found {
			c.JSON(200, value)
			c.Abort()
			return
		}
		
		// Process request and cache response
		c.Next()
		
		// Cache successful responses (placeholder for response capture)
		// In production, you'd capture the actual response body via gin.Context.Writer
	}
}

// DBStatsMiddleware adds database stats to response headers
func DBStatsMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		
		// Get database stats if available
		sqlDB, err := db.DB()
		if err == nil {
			stats := sqlDB.Stats()
			c.Header("X-DB-Idle-Connections", strconv.Itoa(stats.Idle))
			c.Header("X-DB-Open-Connections", strconv.Itoa(stats.OpenConnections))
			c.Header("X-DB-InUse-Connections", strconv.Itoa(stats.InUse))
		}
	}
}

// RequestCounter tracks request counts for QPS calculation
var (
	requestCount    int64 = 0
	requestCountLock sync.Mutex
	resetTime       time.Time
)

// IncrementRequestCount increments the request counter
func IncrementRequestCount() {
	requestCountLock.Lock()
	defer requestCountLock.Unlock()
	
	now := time.Now()
	// Reset counter if it's been more than a second
	if now.Sub(resetTime) > time.Second {
		resetTime = now
		requestCount = 0
	}
	requestCount++
}

// GetQPS returns the current queries per second
func GetQPS() int64 {
	requestCountLock.Lock()
	defer requestCountLock.Unlock()
	return requestCount
}

// ResetQPS resets the QPS counter
func ResetQPS() {
	requestCountLock.Lock()
	defer requestCountLock.Unlock()
	resetTime = time.Now()
	requestCount = 0
}
