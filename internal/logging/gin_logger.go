package logging

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func GinLogrusLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		if strings.HasPrefix(path, "/v1/management") || strings.HasPrefix(path, "/v0/management") || strings.HasPrefix(path, "/management") {
			c.Next()
			return
		}

		start := time.Now()
		raw := maskSensitiveQuery(c.Request.URL.RawQuery)

		c.Next()

		if raw != "" {
			path = path + "?" + raw
		}

		latency := time.Since(start)
		if latency > time.Minute {
			latency = latency.Truncate(time.Second)
		} else {
			latency = latency.Truncate(time.Millisecond)
		}

		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()
		timestamp := time.Now().Format("2006/01/02 - 15:04:05")
		logLine := fmt.Sprintf("[GIN] %s | %3d | %13v | %15s | %-7s \"%s\"", timestamp, statusCode, latency, clientIP, method, path)

		if selectedAuth, exists := c.Get("selected_auth"); exists {
			if authID, ok := selectedAuth.(string); ok && authID != "" {
				logLine = logLine + " | auth=" + authID
			}
		}

		if errorMessage != "" {
			logLine = logLine + " | " + errorMessage
		}

		switch {
		case statusCode >= http.StatusInternalServerError:
			Error(logLine)
		case statusCode >= http.StatusBadRequest:
			Warn(logLine)
		default:
			Info(logLine)
		}
	}
}

func GinLogrusRecovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		WithFields(Fields{
			"panic": recovered,
			"stack": string(debug.Stack()),
			"path":  c.Request.URL.Path,
		}).Error("recovered from panic")

		c.AbortWithStatus(http.StatusInternalServerError)
	})
}
