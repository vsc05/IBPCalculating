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
	router.GET("/api/", h.GetComponents)
	router.GET("/api/getComponent/:id", h.GetComponent)
	router.GET("/api/getComponents", h.GetComponents2)
	router.GET("/api/calcups/:id", h.GetBid)
	router.POST("/api/calcups/:id/delete-component", h.DeleteComponent)
	router.POST("/api/:id/add-to-bid", h.AddComponentToBid)

	//user
	router.POST("/api/users/register", h.RegisterUser)
	router.GET("/api/users/:id", h.GetUser)
	router.PUT("/api/users/:id", h.SetUserChanges)
	router.POST("/api/users/login", h.LoginUser)
	router.POST("/api/users/logout", h.LogoutUser)

	//component
	router.POST("/api/component/createComponent", h.createComponent)
	router.PUT("/api/component/:id", h.UpdateComponent)
	router.DELETE("/api/component/:id", h.DeleteComponentPostman)
	router.POST("/api/component/:id/setComponentImage", h.SetComponentImage)

	//bidUPS
	router.GET("/api/users/bidUPS", h.GetUserCart)
	router.GET("/api/bidUPS", h.GetBidUPS)
	router.POST("/api/bidUPS/:id", h.SetBidUPS)
	router.POST("/api/bidUPS/:id/form", h.FormBidUPS)
	router.PUT("/api/bidUPS/:id/decline", h.DeclineBidUPS)
	router.DELETE("/api/bidUPS/:id", h.DeleteBidUPS)

	//calcUPS
	router.PUT("/api/calcUPS/:id/deleteCalcUPSComponent", h.DeleteCalcUPS)
	router.POST("/api/calcUPS/:id", h.SetCalcUPS)
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
		"data": data,
	})
}
