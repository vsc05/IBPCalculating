package handler

import (
	"Lab1/internal/app/repository"
	"net/http"
	"strconv"

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

func (h *Handler) Trash(ctx *gin.Context) {
	var trash []repository.Trash
	var err error
	var activeTrash repository.Trash

	trash, err = h.Repository.GetTrash()
	if err != nil {
		logrus.Error(err)
	}

	for _, t := range trash {
		if t.Status == "Active" {
			activeTrash = t
			break
		}
	}

	ctx.HTML(http.StatusOK, "trash.html", gin.H{
		"trash": activeTrash,
	})
}

func (h *Handler) GetOrders(ctx *gin.Context) {
	var orders []repository.Order
	var err error
	var trash []repository.Trash
	var activeTrash repository.Trash
	var trashCount int

	trash, err = h.Repository.GetTrash()
	if err != nil {
		logrus.Error(err)
	}

	for _, t := range trash {
		if t.Status == "Active" {
			activeTrash = t
			break
		}
	}

	searchQuery := ctx.Query("query")
	if searchQuery == "" {
		orders, err = h.Repository.GetOrders()
		if err != nil {
			logrus.Error(err)
		}
	} else {
		orders, err = h.Repository.GetOrdersByTitle(searchQuery)
		if err != nil {
			logrus.Error(err)
		}
	}

	trashCount = len(trash)

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"trash":      activeTrash,
		"orders":     orders,
		"query":      searchQuery,
		"trashCount": trashCount,
	})
}

func (h *Handler) GetOrder(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
	}

	order, err := h.Repository.GetOrder(id)
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "order.html", gin.H{
		"order": order,
	})
}

func (h *Handler) GetCartCount(ctx *gin.Context) {

}
