package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

// LogLevel represents different log levels
type LogLevel int

const (
	LogLevelInfo LogLevel = iota
	LogLevelWarn
	LogLevelError
)

// Logger provides structured logging for the wishlist service
type Logger struct {
	serviceName string
}

// NewLogger creates a new logger instance
func NewLogger(serviceName string) *Logger {
	return &Logger{serviceName: serviceName}
}

// LogRequest logs HTTP request information
func (l *Logger) LogRequest(r *http.Request, userID *openapi_types.UUID, action string) {
	userStr := "anonymous"
	if userID != nil {
		userStr = userID.String()
	}

	log.Printf("[%s] %s request from %s (user: %s) - %s %s",
		l.serviceName, action, r.RemoteAddr, userStr, r.Method, r.URL.Path)
}

// LogSuccess logs successful operations
func (l *Logger) LogSuccess(userID *openapi_types.UUID, action string, details string) {
	userStr := "anonymous"
	if userID != nil {
		userStr = userID.String()
	}

	log.Printf("[%s] SUCCESS - %s (user: %s): %s",
		l.serviceName, action, userStr, details)
}

// LogValidationError logs validation errors with details
func (l *Logger) LogValidationError(userID *openapi_types.UUID, action string, errors ValidationErrors) {
	userStr := "anonymous"
	if userID != nil {
		userStr = userID.String()
	}

	log.Printf("[%s] VALIDATION_ERROR - %s (user: %s): %v",
		l.serviceName, action, userStr, errors)
}

// LogError logs general errors
func (l *Logger) LogError(userID *openapi_types.UUID, action string, err error, context string) {
	userStr := "anonymous"
	if userID != nil {
		userStr = userID.String()
	}

	log.Printf("[%s] ERROR - %s (user: %s): %s - %v",
		l.serviceName, action, userStr, context, err)
}

// LogUnauthorized logs unauthorized access attempts
func (l *Logger) LogUnauthorized(r *http.Request, action string, reason string) {
	log.Printf("[%s] UNAUTHORIZED - %s from %s: %s",
		l.serviceName, action, r.RemoteAddr, reason)
}

// LogNotFound logs resource not found errors
func (l *Logger) LogNotFound(userID *openapi_types.UUID, resourceType string, resourceID string) {
	userStr := "anonymous"
	if userID != nil {
		userStr = userID.String()
	}

	log.Printf("[%s] NOT_FOUND - %s '%s' (user: %s)",
		l.serviceName, resourceType, resourceID, userStr)
}

// LogBadRequest logs bad request errors with context
func (l *Logger) LogBadRequest(userID *openapi_types.UUID, action string, reason string) {
	userStr := "anonymous"
	if userID != nil {
		userStr = userID.String()
	}

	log.Printf("[%s] BAD_REQUEST - %s (user: %s): %s",
		l.serviceName, action, userStr, reason)
}

// LogDatabaseOperation logs database operations with performance info
func (l *Logger) LogDatabaseOperation(operation string, duration time.Duration, success bool, details string) {
	status := "SUCCESS"
	if !success {
		status = "FAILED"
	}

	log.Printf("[%s] DB_%s - %s in %v: %s",
		l.serviceName, status, operation, duration, details)
}

// LogStartup logs service startup information
func (l *Logger) LogStartup(port string, dbConnection string) {
	log.Printf("[%s] STARTUP - Service starting on port %s, connected to database: %s",
		l.serviceName, port, dbConnection)
}

// LogShutdown logs service shutdown information
func (l *Logger) LogShutdown(reason string) {
	log.Printf("[%s] SHUTDOWN - Service stopping: %s",
		l.serviceName, reason)
}

// Helper function to format resource details
func FormatResourceDetails(resourceType string, resourceID interface{}, additionalInfo ...string) string {
	details := fmt.Sprintf("%s: %v", resourceType, resourceID)
	if len(additionalInfo) > 0 {
		details += fmt.Sprintf(" (%s)", additionalInfo[0])
	}
	return details
}
