package middleware

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"vsq-oper-manpower/backend/internal/errors"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)
		c.Next()
	}
}

// ErrorHandlerMiddleware handles errors and returns structured error responses
func ErrorHandlerMiddleware() gin.HandlerFunc {
	isDevelopment := os.Getenv("ENVIRONMENT") != "production"

	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			// Get request ID
			requestID, _ := c.Get("request_id")
			requestIDStr := ""
			if id, ok := requestID.(string); ok {
				requestIDStr = id
			}

			// Get user ID from session if available
			userID := ""
			session := sessions.Default(c)
			if uid := session.Get("user_id"); uid != nil {
				if uidStr, ok := uid.(string); ok {
					userID = uidStr
				}
			}

			// Determine error response
			var appErr *errors.AppError
			var statusCode int
			var errorMessage string
			var errorCode errors.ErrorCode

			if errors.IsAppError(err.Err) {
				appErr, _ = errors.AsAppError(err.Err)
				statusCode = appErr.StatusCode
				errorMessage = appErr.Message
				errorCode = appErr.Code
			} else {
				// Unknown error - treat as internal server error
				statusCode = http.StatusInternalServerError
				errorCode = errors.ErrorCodeInternal
				errorMessage = errors.SanitizeError(err.Err, isDevelopment)
			}

			// Log error with context
			logError(c, err.Err, statusCode, requestIDStr, userID, isDevelopment)

			// Build error response
			errorResponse := gin.H{
				"error":      errorMessage,
				"code":       string(errorCode),
				"request_id": requestIDStr,
			}

			// Add details if available
			if appErr != nil && len(appErr.Details) > 0 {
				errorResponse["details"] = appErr.Details
			}

			// In development, include more details
			if isDevelopment && err.Err != nil {
				errorResponse["debug"] = gin.H{
					"message": err.Err.Error(),
					"type":    fmt.Sprintf("%T", err.Err),
				}
			}

			c.JSON(statusCode, errorResponse)
		}
	}
}

// logError logs error information with context
func logError(c *gin.Context, err error, statusCode int, requestID, userID string, isDevelopment bool) {
	timestamp := time.Now().Format(time.RFC3339)
	method := c.Request.Method
	path := c.Request.URL.Path

	log.Printf("[ERROR] [%s] %s %s | Status: %d | RequestID: %s | UserID: %s | Error: %v",
		timestamp, method, path, statusCode, requestID, userID, err)

	// In development, log more details
	if isDevelopment {
		log.Printf("[ERROR-DEBUG] Request Body: %s", getRequestBody(c))
		log.Printf("[ERROR-DEBUG] Query Params: %s", c.Request.URL.RawQuery)
	}
}

// getRequestBody safely extracts request body for logging
func getRequestBody(c *gin.Context) string {
	// Don't log request body for large requests or sensitive endpoints
	if c.Request.ContentLength > 10000 {
		return "[body too large]"
	}

	// Don't log sensitive endpoints
	sensitivePaths := []string{"/auth/login", "/auth/register"}
	for _, path := range sensitivePaths {
		if c.Request.URL.Path == path {
			return "[sensitive data]"
		}
	}

	return "[body logged in development only]"
}
