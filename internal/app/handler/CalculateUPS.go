package handler

import (
	"Lab1/internal/app/ds"
	"fmt"
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

func (h *Handler) GetBid(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
		ctx.Status(http.StatusBadRequest)
		return
	}

	components, err := h.Repository.GetCalcUPSs(id)
	if err != nil {
		logrus.Error(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}
	var total int

	total += int(h.Repository.GetCalcPower(id))

	ctx.HTML(http.StatusOK, "calcUPS.html", gin.H{
		"components": components,
		"bid":        id,
		"result":     total,
	})
}

func (h *Handler) DeleteBidByStatus(ctx *gin.Context) {
	strBidID := ctx.Param("id")
	bidID, _ := strconv.Atoi(strBidID)
	if err := h.Repository.UpdateBidStatus(bidID, "удален"); err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, "Ne polychilos")
		return
	}
}

func (h *Handler) DeleteComponent(ctx *gin.Context) {
	strBidID := ctx.Param("id")
	bidID, err := strconv.Atoi(strBidID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "неверный bidUPS_id"})
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

	ctx.Redirect(http.StatusFound, fmt.Sprintf("/calcups/%d", bidID))
}

func (h *Handler) RegisterUser(ctx *gin.Context) {
	var request ds.RegisterRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный формат данных")
		return
	}

	existingUser, _ := h.Repository.GetUserByUsername(request.Login)
	if existingUser != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Пользователь с таким именем уже существует")
		return
	}

	// тут захэширую потом

	user := &ds.User{
		Login:       request.Login,
		Password:    request.Password, // хэШ!!!!!!!
		IsModerator: false,
	}

	if request.IsModerator != false {
		user.IsModerator = request.IsModerator
	}

	if err := h.Repository.CreateUser(user); err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, "Ошибка создания пользователя: "+err.Error())
		return
	}

	userResponse := gin.H{
		"id":          user.ID,
		"login":       user.Login,
		"isModerator": user.IsModerator,
	}

	h.successResponse(ctx, gin.H{
		"message": "Пользователь успешно зарегистрирован",
		"user":    userResponse,
	})
}

func (h *Handler) GetUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный ID пользователя")
		return
	}

	user, err := h.Repository.GetUserByID(id)
	if err != nil {
		h.errorResponse(ctx, http.StatusNotFound, "Пользователь не найден")
		return
	}

	userResponse := gin.H{
		"id":          user.ID,
		"login":       user.Login,
		"isModerator": user.IsModerator,
	}

	h.successResponse(ctx, userResponse)
}

func (h *Handler) SetUserChanges(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный ID пользователя")
		return
	}

	var request ds.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный формат данных")
		return
	}

	updates := make(map[string]interface{})
	if request.Login != "" {
		existingUser, _ := h.Repository.GetUserByUsername(request.Login)
		if existingUser != nil && existingUser.ID != uint(id) {
			h.errorResponse(ctx, http.StatusBadRequest, "Имя пользователя уже занято")
			return
		}
		updates["login"] = request.Login
	}
	if request.IsModerator != false {
		updates["isModerator"] = request.IsModerator
	}

	if err := h.Repository.UpdateUser(id, updates); err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, "Ошибка обновления пользователя: "+err.Error())
		return
	}

	h.successResponse(ctx, gin.H{
		"message": "Данные пользователя успешно обновлены",
	})
}

func (h *Handler) LoginUser(ctx *gin.Context) {
	var request ds.LoginRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный формат данных")
		return
	}

	user, err := h.Repository.GetUserByUsername(request.Login)
	if err != nil {
		h.errorResponse(ctx, http.StatusUnauthorized, "Неверное имя пользователя или пароль")
		return
	}

	if request.Password != user.Password {
		h.errorResponse(ctx, http.StatusUnauthorized, "Неверное имя пользователя или пароль")
		return
	}

	token := "generated-jwt-token-here" // джейвити

	userResponse := gin.H{
		"id":          user.ID,
		"login":       user.Login,
		"isModerator": user.IsModerator,
	}

	h.successResponse(ctx, gin.H{
		"message": "Успешный вход",
		"token":   token,
		"user":    userResponse,
	})
}

func (h *Handler) LogoutUser(ctx *gin.Context) {
	// пока заглушка
	h.successResponse(ctx, gin.H{
		"message": "Успешный выход",
	})
}

func (h *Handler) createComponent(ctx *gin.Context) {
	var resource ds.Component
	if err := ctx.ShouldBindJSON(&resource); err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный формат данных: "+err.Error())
		return
	}

	// Устанавливаем дефолтные значения
	resource.Image = "/images/default.jpg"

	if err := h.Repository.CreateComponent(&resource); err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, "Ошибка создания компонента: "+err.Error())
		return
	}

	h.successResponse(ctx, gin.H{
		"message":  "Компонент успешно создан",
		"resource": resource,
	})
}

func (h *Handler) UpdateComponent(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный ID компонента")
		return
	}

	var updates ds.Component
	if err := ctx.ShouldBindJSON(&updates); err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный формат данных")
		return
	}

	updateMap := make(map[string]interface{})

	if updates.Title != "" {
		updateMap["title"] = updates.Title
	}
	if updates.Power != 0 {
		updateMap["power"] = updates.Power
	}
	if updates.Coeff != 0 {
		updateMap["coeff"] = updates.Coeff
	}
	if updates.Image != "" {
		updateMap["image"] = updates.Image
	}
	if updates.IsDelete != false {
		updateMap["is_delete"] = updates.IsDelete
	}

	if len(updateMap) == 0 {
		h.errorResponse(ctx, http.StatusBadRequest, "Нет данных для обновления")
		return
	}

	if err := h.Repository.UpdateComponent(id, updateMap); err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, "Ошибка обновления Компонента: "+err.Error())
		return
	}

	h.successResponse(ctx, gin.H{
		"message": "Компонент успешно обновлен",
	})
}

func (h *Handler) DeleteComponentPostman(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный ID Компонента")
		return
	}

	if err := h.Repository.DeleteComponentPostman(id); err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, "Ошибка удаления Компонента: "+err.Error())
		return
	}

	h.successResponse(ctx, gin.H{
		"message": "Компонент успешно удален",
	})
}

func (h *Handler) SetComponentImage(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный ID Компонента")
		return
	}

	var request struct {
		Image string `json:"image" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный формат данных")
		return
	}

	if err := h.Repository.SetComponentImage(id, request.Image); err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, "Ошибка установки изображения: "+err.Error())
		return
	}

	h.successResponse(ctx, gin.H{
		"message": "Изображение успешно установлено",
	})
}

func (h *Handler) GetUserCart(ctx *gin.Context) {
	userIDStr := ctx.Param("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный ID пользователя")
		return
	}

	applications, err := h.Repository.GetUserCart(userID)
	if err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, "Ошибка получения корзины: "+err.Error())
		return
	}

	h.successResponse(ctx, gin.H{
		"applications": applications,
	})
}

func (h *Handler) GetBidUPS(ctx *gin.Context) {
	applications, err := h.Repository.GetAllBidUPS()
	if err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, "Ошибка получения заявок: "+err.Error())
		return
	}

	h.successResponse(ctx, gin.H{
		"applications": applications,
		"total":        len(applications),
	})
}

func (h *Handler) FormBidUPS(ctx *gin.Context) {
	applicationIDStr := ctx.Param("id")
	applicationID, err := strconv.Atoi(applicationIDStr)
	if err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный ID заявки")
		return
	}

	err = h.Repository.FormBidUPS(applicationID)
	if err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	h.successResponse(ctx, gin.H{
		"message": "Заявка успешно отправлена",
	})
}

func (h *Handler) DeclineBidUPS(ctx *gin.Context) {
	applicationIDStr := ctx.Param("id")
	applicationID, err := strconv.Atoi(applicationIDStr)
	if err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный ID заявки")
		return
	}

	var request struct {
		ModeratorID int `json:"moderator_id" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный формат данных")
		return
	}

	if err := h.Repository.DeclineBidUPS(applicationID, request.ModeratorID); err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, "Ошибка отклонения заявки: "+err.Error())
		return
	}

	h.successResponse(ctx, gin.H{
		"message": "Заявка успешно отклонена",
	})
}

func (h *Handler) SetBidUPS(ctx *gin.Context) {
	applicationIDStr := ctx.Param("id")
	applicationID, err := strconv.Atoi(applicationIDStr)
	if err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный ID заявки")
		return
	}

	var request struct {
		Weight       int `json:"creator_id" binding:"required"`
		Productivity int `json:"moderator_id" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный формат данных")
		return
	}

	err = h.Repository.SetBidUPS(applicationID, request.Weight, request.Productivity)
	if err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	h.successResponse(ctx, gin.H{
		"message": "Изменения успешно сохранены",
	})
}

func (h *Handler) DeleteCalcUPS(ctx *gin.Context) {
	resourceIDStr := ctx.Param("id")
	resourceID, err := strconv.Atoi(resourceIDStr)
	if err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный ID ресурса")
		return
	}

	var request struct {
		ApplicationID int `json:"bidId" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный формат данных")
		return
	}

	if err := h.Repository.DeleteCalcUPS(request.ApplicationID, resourceID); err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, "Ошибка удаления компонента из заявки: "+err.Error())
		return
	}

	h.successResponse(ctx, gin.H{
		"message": "Компонент успешно удален из заявки",
	})
}

func (h *Handler) SetCalcUPS(ctx *gin.Context) {
	resourceIDStr := ctx.Param("id")
	resourceID, err := strconv.Atoi(resourceIDStr)
	if err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный ID ресурса")
		return
	}

	var request struct {
		ApplicationID int `json:"bidId" binding:"required"`
		ComponentID   int `json:"componentId" binding:"required"`
		Coeff         int `json:"battery_life" binding:"required"`
		Power         int `json:"incoming_power" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный формат данных")
		return
	}

	if err := h.Repository.SetCalcUPS(request.ApplicationID, resourceID, request.ComponentID, request.Coeff, request.Power); err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, "Ошибка установки коэффициента: "+err.Error())
		return
	}

	h.successResponse(ctx, gin.H{
		"message": "Коэффициент успешно установлен",
	})
}
