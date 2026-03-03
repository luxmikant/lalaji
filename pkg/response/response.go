package response

import (
	"github.com/gin-gonic/gin"
)

// APIResponse is the standard envelope for all API responses.
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Fields    interface{} `json:"fields,omitempty"`
	RequestID string      `json:"requestId,omitempty"`
}

// Success returns a success envelope for the caller to pass to c.JSON.
func Success(c *gin.Context, data interface{}) APIResponse {
	return APIResponse{
		Success:   true,
		Data:      data,
		RequestID: getRequestID(c),
	}
}

// Error returns an error envelope for the caller to pass to c.JSON.
func Error(c *gin.Context, statusCode int, message string) APIResponse {
	return APIResponse{
		Success:   false,
		Error:     message,
		RequestID: getRequestID(c),
	}
}

// getRequestID extracts the requestId from the Gin context.
func getRequestID(c *gin.Context) string {
	if id, exists := c.Get("requestId"); exists {
		return id.(string)
	}
	return ""
}
