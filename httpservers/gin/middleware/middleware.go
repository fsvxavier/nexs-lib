package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/fsvxavier/nexs-lib/httpservers/common"
)

// ErrorHandler handles errors and returns appropriate response
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			// Check if it's our custom APIError
			if apiErr, ok := err.(*common.APIError); ok {
				c.JSON(apiErr.StatusCode, apiErr)
				return
			}

			// Default error
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
		}
	}
}

// Logger middleware logs HTTP requests
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Store the request ID in context
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
			c.Header("X-Request-ID", requestID)
		}

		// Process request
		c.Next()

		// Measure execution time
		execTime := time.Since(start).String()

		// Set response header
		c.Header("X-Response-Time", execTime)

		// Log format: method path status response-time request-id
		fmt.Printf("[%s] %s %s - %d - %s - %s\n",
			time.Now().Format("2006-01-02 15:04:05"),
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			execTime,
			requestID,
		)
	}
}

// Tracing adds tracing capability
func Tracing() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate trace ID if not present
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = uuid.New().String()
			c.Header("X-Trace-ID", traceID)
		}

		// Set span ID
		spanID := uuid.New().String()
		c.Header("X-Span-ID", spanID)

		c.Next()
	}
}

// CORS middleware for handling CORS requests
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RegisterPprof registers pprof routes on the router
func RegisterPprof(router *gin.Engine) {
	// Implementation depends on project dependencies
	// Example:
	// pprof.Register(router, "/debug/pprof")
	router.GET("/debug/pprof", func(c *gin.Context) {
		c.String(http.StatusOK, "Pprof enabled")
	})
}

// RegisterMetrics registers metric routes on the router
func RegisterMetrics(router *gin.Engine) {
	// Implementation depends on project dependencies
	// Example:
	// router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.GET("/metrics", func(c *gin.Context) {
		c.String(http.StatusOK, "Metrics endpoint enabled")
	})
}

// RegisterSwagger registers swagger routes on the router
func RegisterSwagger(router *gin.Engine) {
	// Implementation depends on project dependencies
	// Example:
	// router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/swagger/*any", func(c *gin.Context) {
		c.String(http.StatusOK, "Swagger enabled")
	})
}
