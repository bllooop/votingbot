package api

import (
	"github.com/bllooop/votingbot/internal/usecase"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	Usecases *usecase.Usecase
}

func NewHandler(usecases *usecase.Usecase) *Handler {
	return &Handler{Usecases: usecases}
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	router.POST("/vote", h.VoteHandler)
	return router
}
