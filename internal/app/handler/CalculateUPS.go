package handler

import (
	"DIA3Course/internal/app/ds"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sirupsen/logrus"
)

// @Summary Получение списка компонентов
// @Description Возвращает все компоненты (HTML)
// @Tags Components
// @Param Authorization header string true "Bearer <access_token>"
// @Security BearerAuth
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

// @Summary Получение одного компонента
// @Description Возвращает компонент по ID
// @Tags Components
// @Param Authorization header string true "Bearer <access_token>"
// @Security BearerAuth
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

// @Summary Добавление компонента в заявку
// @Tags Bid
// @Param Authorization header string true "Bearer <access_token>"
// @Security BearerAuth
func (h *Handler) AddComponentToBid(ctx *gin.Context) {
	compIDStr := ctx.Param("id")
	compID, err := strconv.Atoi(compIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid component id"})
		return
	}
	draftBid, err := h.Repository.GetDraftBid()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "не найдена заявка-черновик: " + err.Error()})
		return
	}
	err = h.Repository.AddComponentToBid(int(draftBid.ID), compID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.successResponse(ctx, gin.H{
		"message": "Компонент успешно добавлен в текущую заявку",
	})
	ctx.Redirect(http.StatusFound, "/")
}

// @Summary Получение заявки
// @Description Возвращает заявку с компонентами
// @Tags Bid
// @Param Authorization header string true "Bearer <access_token>"
// @Security BearerAuth
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

// @Summary Удаление заявки по статусу
// @Tags Bid
// @Param Authorization header string true "Bearer <access_token>"
// @Security BearerAuth
func (h *Handler) DeleteBidByStatus(ctx *gin.Context) {
	strBidID := ctx.Param("id")
	bidID, _ := strconv.Atoi(strBidID)
	if err := h.Repository.UpdateBidStatus(bidID, "удален"); err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, "Ne polychilos")
		return
	}
}

// @Summary Удаление компонента из заявки
// @Tags Bid
// @Param Authorization header string true "Bearer <access_token>"
// @Security BearerAuth
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

// @Summary Регистрация пользователя
// @Description Регистрирует нового пользователя
// @Tags Auth
// @Param Authorization header string true "Bearer <access_token>"
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param input body ds.RegisterRequest true "Register user"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /register [post]
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

// @Summary Получение информации о пользователе
// @Description Возвращает данные пользователя по ID
// @Tags Users
// @Param Authorization header string true "Bearer <access_token>"
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string "Invalid ID"
// @Failure 404 {object} map[string]string "User not found"
// @Router /users/{id} [get]
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

// @Summary Обновление данных пользователя
// @Description Обновляет данные пользователя
// @Tags Users
// @Param Authorization header string true "Bearer <access_token>"
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param input body ds.UpdateUserRequest true "Update user"
// @Success 200 {object} map[string]string "Success"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /users/{id} [put]
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

// LoginUser godoc
// @Summary User login
// @Description Authenticates user and returns JWT token
// @Tags Auth
// @Accept json
// @Produce json
// @Param input body ds.LoginRequest true "Login credentials"
// @Success 200 {object} ds.LoginResponse
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Invalid credentials"
// @Router /login [post]
func (h *Handler) LoginUser(gCtx *gin.Context) {
	var req ds.LoginRequest
	if err := gCtx.ShouldBindJSON(&req); err != nil {
		gCtx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат"})
		return
	}

	user, err := h.Repository.GetUserByUsername(req.Login)
	if err != nil || user.Password != req.Password {
		gCtx.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный логин или пароль"})
		return
	}

	expiration := time.Hour * 24
	claims := ds.JWTClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(expiration).Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "bitop-admin",
		},
		UserUUID:    uuid.New(),
		Scopes:      []string{"read", "write"},
		IsModerator: user.IsModerator,
		Login:       user.Login,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("my-key"))

	gCtx.JSON(http.StatusOK, gin.H{
		"expires_in":   expiration.Seconds(),
		"access_token": tokenString,
		"token_type":   "Bearer",
	})
}

// @Summary Выход пользователя
// @Description Логаут и добавление JWT в черный список
// @Tags Auth
// @Param Authorization header string true "Bearer <access_token>"
// @Security BearerAuth
// @Success 200 {string} string "Success"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /logout [post]
func (h *Handler) Logout(gCtx *gin.Context) {
	jwtStr := gCtx.GetHeader("Authorization")
	if !strings.HasPrefix(jwtStr, jwtPrefix) {
		gCtx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	jwtStr = jwtStr[len(jwtPrefix):]

	_, err := jwt.ParseWithClaims(jwtStr, &ds.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("my-key"), nil
	})
	if err != nil {
		gCtx.AbortWithError(http.StatusBadRequest, err)
		log.Println(err)
		return
	}

	// сохраняем в чёрный список
	err = h.Redis.WriteJWTToBlacklist(gCtx.Request.Context(), jwtStr, h.Config.JWT.Expiration)
	if err != nil {
		gCtx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	gCtx.Status(http.StatusOK)
}

// @Summary Создание компонента
// @Description Создает новый компонент
// @Tags Components
// @Param Authorization header string true "Bearer <access_token>"
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param input body ds.Component true "Component data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /components [post]
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

// @Summary Обновление компонента
// @Description Обновляет данные компонента
// @Tags Components
// @Param Authorization header string true "Bearer <access_token>"
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Component ID"
// @Param input body ds.Component true "Updated component"
// @Success 200 {object} map[string]string "Success"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /components/{id} [put]
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

// @Summary Удаление компонента
// @Description Удаляет компонент из системы
// @Tags Components
// @Param Authorization header string true "Bearer <access_token>"
// @Security BearerAuth
// @Param id path int true "Component ID"
// @Success 200 {object} map[string]string "Success"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /components/{id} [delete]
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

// @Summary Установка изображения компонента
// @Description Загружает и сохраняет изображение компонента на MinIO
// @Tags Components
// @Param Authorization header string true "Bearer <access_token>"
// @Security BearerAuth
// @Param id path int true "Component ID"
// @Param image formData file true "Image file"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /components/{id}/image [post]
func (h *Handler) SetComponentImage(ctx *gin.Context) {
	// Загружаем env, если ещё не
	_ = godotenv.Load()

	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный ID компонента")
		return
	}

	file, err := ctx.FormFile("image")
	if err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Не удалось получить файл")
		return
	}

	src, err := file.Open()
	if err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, "Ошибка открытия файла")
		return
	}
	defer src.Close()

	// Подключаемся к MinIO
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	secretKey := os.Getenv("MINIO_SECRET_KEY")
	bucketName := os.Getenv("MINIO_BUCKET")
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, "Ошибка подключения к MinIO: "+err.Error())
		return
	}

	// Проверяем / создаём бакет
	ctxMinio := context.Background()
	exists, err := minioClient.BucketExists(ctxMinio, bucketName)
	if err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, "Ошибка проверки бакета: "+err.Error())
		return
	}
	if !exists {
		err = minioClient.MakeBucket(ctxMinio, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			h.errorResponse(ctx, http.StatusInternalServerError, "Ошибка создания бакета: "+err.Error())
			return
		}
	}

	// Имя файла
	objectName := fmt.Sprintf("component_%d_%d%s", id, time.Now().Unix(), filepath.Ext(file.Filename))

	// Загружаем файл
	uploadInfo, err := minioClient.PutObject(ctxMinio, bucketName, objectName, src, file.Size, minio.PutObjectOptions{
		ContentType: file.Header.Get("Content-Type"),
	})
	if err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, "Ошибка загрузки в MinIO: "+err.Error())
		return
	}

	// Формируем публичную ссылку
	imageURL := fmt.Sprintf("http://%s/%s/%s", endpoint, bucketName, uploadInfo.Key)

	// Сохраняем в БД
	if err := h.Repository.SetComponentImage(id, imageURL); err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, "Ошибка записи в БД: "+err.Error())
		return
	}

	h.successResponse(ctx, gin.H{
		"message":  "Изображение успешно загружено в MinIO",
		"imageURL": imageURL,
	})
}

func (h *Handler) GetUserCart(ctx *gin.Context) {
	userLogin, exists := ctx.Get("userLogin")
	if !exists {
		h.errorResponse(ctx, http.StatusUnauthorized, "User not authenticated")
		return
	}

	bids, err := h.Repository.GetUserCartDetails(userLogin.(string))
	if err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, "Ошибка получения корзины: "+err.Error())
		return
	}

	h.successResponse(ctx, gin.H{
		"cart_count":   len(bids),
		"applications": bids,
		"user_login":   userLogin,
	})
}

// @Summary Получение списка заявок UPS
// @Description Возвращает все заявки UPS из системы с общим количеством
// @Tags Bid UPS
// @Produce json
// @Param Authorization header string true "Bearer <access_token>"
// @Success 200 {object} map[string]interface{} "Успешный ответ"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 500 {object} map[string]string "Ошибка сервера"
// @Security BearerAuth
// @Router /api/bidUPS [get]
func (h *Handler) GetBidUPS(ctx *gin.Context) {
	applications, err := h.Repository.GetAllBidUPS()
	if err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, "Ошибка получения заявок: "+err.Error())
		return
	}

	h.successResponse(ctx, gin.H{
		"BidUPS": applications,
		"total":  len(applications),
	})
}

// GetComponents2 godoc
// @Summary Get all components (alternative endpoint)
// @Description Get list of all components - alternative version
// @Tags Components
// @Produce json
// @Success 200 {array} ds.Component
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/getComponents [get]
func (h *Handler) GetComponents2(ctx *gin.Context) {
	applications, err := h.Repository.GetAllComponents()
	if err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, "Ошибка получения компонентов: "+err.Error())
		return
	}

	h.successResponse(ctx, gin.H{
		"Components": applications,
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
		"message": "Заявка успешно сформирована",
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
		ModeratorID int    `json:"moderator_id" binding:"required"`
		Status      string `json:"status" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный формат данных")
		return
	}

	if err := h.Repository.DeclineBidUPS(applicationID, request.ModeratorID, request.Status); err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, "Ошибка отклонения заявки: "+err.Error())
		return
	}

	h.successResponse(ctx, gin.H{
		"message": "Заявка успешно отклонена",
	})
}

func (h *Handler) DeleteBidUPS(ctx *gin.Context) {
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

	if err := h.Repository.DeleteBidUPS(applicationID, request.ModeratorID); err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, "Ошибка отклонения заявки: "+err.Error())
		return
	}

	h.successResponse(ctx, gin.H{
		"message": "Заявка успешно удалена",
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
		"message": "Успешно установлено",
	})
}
