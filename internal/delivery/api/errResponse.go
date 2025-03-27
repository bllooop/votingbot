package api

import (
	logger "github.com/bllooop/votingbot/pkg/logging"
	"github.com/gin-gonic/gin"
)

type errorResponse struct {
	Message string `json:"message"`
}

func newErrorResponse(c *gin.Context, statusCode int, message string) {
	logger.Log.Error().Msg(message)
	c.AbortWithStatusJSON(statusCode, errorResponse{message})
}
