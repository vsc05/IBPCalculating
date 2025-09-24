package handler

import (
	"Lab1/internal/app/ds"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handler) GetComponents(ctx *gin.Context) {
	var components []ds.Component
	var err error

	searchQuery := ctx.Query("query")
	if searchQuery == "" {
		components, err = h.Repository.GetComponents()
		if err != nil {
			logrus.Error(err)
		}
	} else {
		components, err = h.Repository.GetComponentsByTitle(searchQuery)
		if err != nil {
			logrus.Error(err)
		}
	}

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"orders": components,
		"query":  searchQuery,
		"count":  h.Repository.GetBinCount(1),
	})
}

func (h *Handler) GetComponent(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
	}

	component, err := h.Repository.GetComponent(id)
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "component.html", gin.H{
		"order": component,
	})
}

func (h *Handler) AddComponentToBid(ctx *gin.Context) {
	bidID := 7 // пока захардкодим черновик
	compIDStr := ctx.PostForm("component_id")

	compID, err := strconv.Atoi(compIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid component id"})
		return
	}

	err = h.Repository.AddComponentToBid(bidID, compID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Redirect(http.StatusFound, "/")
}
