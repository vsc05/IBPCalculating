package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handler) GetBid(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
		ctx.Status(http.StatusBadRequest)
		return
	}

	components, err := h.Repository.GetBidComponents(id)
	if err != nil {
		logrus.Error(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}
	var total int

	total += int(h.Repository.GetCalcPower(id))

	ctx.HTML(http.StatusOK, "bid.html", gin.H{
		"components": components,
		"bid":        id,
		"result":     total,
	})
}

func (h *Handler) DeleteBidByStatus(ctx *gin.Context) {
	strBidID := ctx.Param("id")
	bidID, _ := strconv.Atoi(strBidID)
	if err := h.Repository.UpdateBidStatus(bidID, "удален"); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
}

func (h *Handler) DeleteComponent(ctx *gin.Context) {
	strBidID := ctx.Param("id")
	bidID, err := strconv.Atoi(strBidID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "неверный bid_id"})
		return
	}

	strCompID := ctx.PostForm("component_id")
	compID, err := strconv.Atoi(strCompID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "неверный component_id"})
		return
	}

	err = h.Repository.DeleteComponent(uint(bidID), uint(compID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Redirect(http.StatusFound, fmt.Sprintf("/bid/%d", bidID))
}

// // POST /save-bid
// func (h *Handler) SaveBid(ctx *gin.Context) {
// 	// считываем bid_id как в примере удаления
// 	strId := ctx.PostForm("bid_id")
// 	bidID, err := strconv.Atoi(strId)
// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// часы работы
// 	hoursStr := ctx.PostForm("hours")
// 	hours, err := strconv.Atoi(hoursStr)
// 	if err != nil || hours <= 0 {
// 		comps, _ := h.Repository.GetBidComponents(bidID)
// 		ctx.HTML(http.StatusOK, "bid.html", gin.H{
// 			"components": comps,
// 			"bid":        bidID,
// 			"error":      "Введите все необходимые данные для подсчета мощности ИБП",
// 		})
// 		return
// 	}

// 	components, err := h.Repository.GetBidComponents(bidID)
// 	if err != nil {
// 		h.errorHandler(ctx, http.StatusInternalServerError, err)
// 		return
// 	}

// 	var totalPower int

// 	for _, c := range components {
// 		compID := int(c.ID) // приведение uint->int если нужно
// 		field := fmt.Sprintf("incoming_current_%d", compID)
// 		currentStr := ctx.PostForm(field)
// 		if currentStr == "" {
// 			ctx.HTML(http.StatusOK, "bid.html", gin.H{
// 				"components": components,
// 				"bid":        bidID,
// 				"error":      "Введите все необходимые данные для подсчета мощности ИБП",
// 			})
// 			return
// 		}

// 		current, err := strconv.Atoi(currentStr)
// 		if err != nil {
// 			ctx.HTML(http.StatusOK, "bid.html", gin.H{
// 				"components": components,
// 				"bid":        bidID,
// 				"error":      "Неверный формат входящего тока",
// 			})
// 			return
// 		}

// 		// рассчитываем: hours * (current + component.Power) * component.Coeff
// 		power := int(c.Power)
// 		coeff := int(math.Ceil(float64(c.Coeff)))
// 		calc := hours * (current + power) * coeff
// 		totalPower += calc

// 		// сохраняем в bid_components
// 		if err := h.Repository.UpdateBidComponent(bidID, compID, current, totalPower, hours); err != nil {
// 			h.errorHandler(ctx, http.StatusInternalServerError, err)
// 			return
// 		}
// 	}

// 	// обновляем статус заявки (как в примере UpdateColumn)
// 	if err := h.Repository.UpdateBidStatus(bidID, "рассчитан"); err != nil {
// 		h.errorHandler(ctx, http.StatusInternalServerError, err)
// 		return
// 	}

// 	// PRG — редирект на /bid/:id чтобы обновлённая информация подтянулась GET-ом
// 	ctx.Redirect(http.StatusFound, fmt.Sprintf("/bid/%d", bidID))
// }
