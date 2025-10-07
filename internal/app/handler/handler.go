package handler

import (
	"Lab1/internal/app/repository"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

func (h *Handler) RegisterHandler(router *gin.Engine) {
	router.GET("/", h.GetComponents)
	router.GET("/component/:id", h.GetComponent)
	router.GET("/calcups/:id", h.GetBid)
	router.POST("/calcups/:id/delete-component", h.DeleteComponent)
	router.POST("/add-to-bid", h.AddComponentToBid)

	//user
	router.POST("/users/register", h.RegisterUser)
	router.GET("/users/:id", h.GetUser)
	router.PUT("/users/:id/setChanges", h.SetUserChanges)
	router.POST("/users/login", h.LoginUser)
	router.POST("/users/logout", h.LogoutUser)

	//component
	router.POST("/component/createComponent", h.createComponent)
	router.PUT("/component/:id/update", h.UpdateComponent)
	router.DELETE("/component/:id/delete", h.DeleteComponentPostman)
	router.POST("/component/:id/setComponentImage", h.SetComponentImage)

	//bidUPS
	router.GET("/users/:id/bidUPS", h.GetUserCart)
	router.GET("/bidUPS", h.GetBidUPS)
	router.POST("/bidUPS/:id/setChanges", h.SetBidUPS)
	router.POST("/bidUPS/:id/form", h.FormBidUPS)
	router.PUT("/bidUPS/:id/decline", h.DeclineBidUPS)

	//calcUPS
	router.PUT("/calcUPS/:id/deleteCalcUPSComponent", h.DeleteCalcUPS)
	router.POST("/calcUPS/:id/setCalcUPS", h.SetCalcUPS)
}

func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("../../templates/*")
	router.Static("../../static", "../../resources")
}

func (h *Handler) errorResponse(ctx *gin.Context, statusCode int, message string) {
	logrus.Error(message)
	ctx.JSON(statusCode, gin.H{
		"status":  "error",
		"message": message,
	})
}

func (h *Handler) successResponse(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   data,
	})
}
