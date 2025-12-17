package handler

import (
	"DIA3Course/internal/app/ds"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	hardcodedUserInstance *HardcodedUser
	hardcodedUserOnce     sync.Once
)

type HardcodedUser struct {
	Login      string
	BidID      int
	ItemsCount int
}

func GetHardcodedUser() *HardcodedUser {
	hardcodedUserOnce.Do(func() {
		hardcodedUserInstance = &HardcodedUser{
			Login:      "hardcoded_user",
			BidID:      -1,
			ItemsCount: 0,
		}
	})

	return hardcodedUserInstance
}

func (h *Handler) GetUserCart2(ctx *gin.Context) {
	hardcodedUser := GetHardcodedUser()

	h.successResponse(ctx, gin.H{
		"bid_id":      hardcodedUser.BidID,
		"items_count": hardcodedUser.ItemsCount,
	})
}

// @Summary Получение списка компонентовfffff
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

// GetComponent получает компонент по ID
// @Summary Получить компонент по ID
// @Description Возвращает детальную информацию о компоненте по указанному ID
// @Tags Components
// @Accept json
// @Produce json
// @Param id path integer true "ID компонента" format(int) minimum(1)
// @Param Authorization header string true "Bearer токен авторизации" default(Bearer )
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Успешный ответ с данными компонента"
// @Failure 400 {object} map[string]interface{} "Неверный формат ID"
// @Failure 401 "Пользователь не авторизован"
// @Failure 404 {object} map[string]interface{} "Компонент не найден"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /api/components/{id} [get]
func (h *Handler) GetComponent(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"handler": "GetComponent",
			"id":      idStr,
		}).Error("Invalid ID format: ", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid ID format",
		})
		return
	}

	component, err := h.Repository.GetComponent(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logrus.WithFields(logrus.Fields{
				"handler": "GetComponent",
				"id":      id,
			}).Warn("Component not found")
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "Component not found",
			})
			return
		}

		logrus.WithFields(logrus.Fields{
			"handler": "GetComponent",
			"id":      id,
		}).Error("Database error: ", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
		return
	}

	// Возвращаем компонент в формате JSON
	ctx.JSON(http.StatusOK, gin.H{
		"component": component,
	})
}

// AddComponentToBid добавляет компонент в черновик заявки
// @Summary Добавить компонент в черновик заявки
// @Description Добавляет указанный компонент в текущую черновую заявку пользователя. После успешного добавления происходит редирект на главную страницу.
// @Tags Components, bidUPS
// @Accept json
// @Produce json
// @Param id path integer true "ID компонента" format(int) minimum(1)
// @Param Authorization header string true "Bearer токен авторизации" default(Bearer )
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Компонент успешно добавлен"
// @Failure 302 "Редирект на главную страницу после успешного добавления"
// @Failure 400 {object} map[string]interface{} "Неверный ID компонента"
// @Failure 401 "Пользователь не авторизован"
// @Failure 404 {object} map[string]interface{} "Черновая заявка не найдена"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /api/component/{id} [post]
func (h *Handler) AddComponentToBid(ctx *gin.Context) {
	// Получаем логин пользователя из контекста (добавляется в middleware WithAuthCheck)
	userLogin, exists := ctx.Get("userLogin")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	compIDStr := ctx.Param("id")
	compID, err := strconv.Atoi(compIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid component id"})
		return
	}

	// Передаем логин пользователя в функцию GetDraftBid
	draftBid, err := h.Repository.GetOrCreateDraftBid(userLogin.(string))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "не найдена заявка-черновик: " + err.Error()})
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
// @Tags User
// @Param Authorization header string true "Bearer <access_token>"
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param input body ds.RegisterRequest true "Register user"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/register [post]
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
// @Tags User
// @Param Authorization header string true "Bearer <access_token>"
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string "Invalid ID"
// @Failure 404 {object} map[string]string "User not found"
// @Router /api/users/{id} [get]
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
// @Tags User
// @Param Authorization header string true "Bearer <access_token>"
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param input body ds.UpdateUserRequest true "Update user"
// @Success 200 {object} map[string]string "Success"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/users/{id} [put]
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

	// 1. Обработка Логина
	if request.Login != "" {
		// Проверка на уникальность логина
		existingUser, _ := h.Repository.GetUserByUsername(request.Login)
		if existingUser != nil && existingUser.ID != uint(id) {
			h.errorResponse(ctx, http.StatusBadRequest, "Имя пользователя уже занято")
			return
		}
		updates["login"] = request.Login
	}

	// 2. Обработка Пароля (Без хеширования, прямое обновление)
	if request.Password != "" {
		// Предполагаем, что поле в БД называется 'password' (или 'password_hash' если ты хочешь его обновлять)
		updates["password"] = request.Password
	}

	// Если ни логин, ни пароль не переданы, нет смысла обновлять
	if len(updates) == 0 {
		h.errorResponse(ctx, http.StatusBadRequest, "Нет данных для обновления")
		return
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
// @Tags User
// @Accept json
// @Produce json
// @Param input body ds.LoginRequest true "Login credentials"
// @Success 200 {object} ds.LoginResponse
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Invalid credentials"
// @Router /api/login [post]
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
		UserDBID:    user.ID, // <-- Устанавливаем ID пользователя из БД
		UserUUID:    uuid.New(),
		Scopes:      []string{"read", "write"},
		IsModerator: user.IsModerator,
		Login:       user.Login,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(h.Config.JWT.Secret))

	gCtx.JSON(http.StatusOK, gin.H{
		"expires_in":   expiration.Seconds(),
		"access_token": tokenString,
		"token_type":   "Bearer",
	})
}

// @Summary Выход пользователя
// @Description Логаут и добавление JWT в черный список
// @Tags User
// @Param Authorization header string true "Bearer <access_token>"
// @Security BearerAuth
// @Success 200 {string} string "Success"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/logout [post]
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
// @Router /api/component [post]
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
// @Router /api/component/{id} [put]
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
// @Router /api/components/{id} [delete]
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
// @Router /api/component/{id}/setComponentImage [post]
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

// GetUserCart получает корзину пользователя
// @Summary Получить корзину пользователя
// @Description Возвращает данные корзины (ID бида и количество товаров) для текущего авторизованного пользователя
// @Tags bidUPS
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Успешный ответ"
// @Failure 401 {object} map[string]interface{} "Пользователь не авторизован"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /api/bidUPS [get]
// @Param Authorization header string true "Bearer токен авторизации" default(Bearer )
func (h *Handler) GetUserCart(ctx *gin.Context) {
	userLogin, exists := ctx.Get("userLogin")
	if !exists {
		h.errorResponse(ctx, http.StatusUnauthorized, "User not authenticated")
		return
	}

	bidID, itemsCount, err := h.Repository.GetUserCartDetails(userLogin.(string))
	if err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, "Ошибка получения корзины: "+err.Error())
		return
	}

	h.successResponse(ctx, gin.H{
		"bid_id":      bidID,
		"items_count": itemsCount,
	})
}

// @Summary Получение списка заявок UPS
// @Description Возвращает все заявки UPS из системы с общим количеством
// @Tags bidUPS
// @Produce json
// @Param Authorization header string true "Bearer <access_token>"
// @Success 200 {object} map[string]interface{} "Успешный ответ"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 500 {object} map[string]string "Ошибка сервера"
// @Security BearerAuth
// @Router /api/bidUPSAll [get]
func (h *Handler) GetBidUPS(ctx *gin.Context) {
	// Создаем карту фильтров
	filters := make(map[string]interface{})

	// Получаем параметры фильтрации из query string
	if startDate := ctx.Query("start_date"); startDate != "" {
		filters["start_date"] = startDate
	}
	if endDate := ctx.Query("end_date"); endDate != "" {
		filters["end_date"] = endDate
	}
	if status := ctx.Query("status"); status != "" {
		filters["status"] = status
	}

	applications, err := h.Repository.GetAllBidUPS(filters)
	if err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, "Ошибка получения заявок: "+err.Error())
		return
	}

	// Преобразуем в response DTO чтобы вернуть только нужные поля
	type BidUPSResponse struct {
		ID                   uint      `json:"id"`
		Status               string    `json:"status"`
		DateUpdate           time.Time `json:"date_update"`
		DateFinish           time.Time `json:"date_finish"`
		IncomingCurrent      int       `json:"incoming_current"`
		CreatorLogin         string    `json:"creator_login"`
		ModeratorLogin       string    `json:"moderator_login"`
		CalculatedPowerCount int       `json:"calculated_power_count"`
	}

	var response []BidUPSResponse
	for _, app := range applications {
		moderatorLogin := ""
		if app.Moderator.ID != 0 {
			moderatorLogin = app.Moderator.Login
		}

		dateFinish := time.Time{}
		if app.DateFinish.Valid {
			dateFinish = app.DateFinish.Time
		}

		response = append(response, BidUPSResponse{
			ID:                   app.ID,
			Status:               app.Status,
			DateUpdate:           app.DateUpdate,
			DateFinish:           dateFinish,
			IncomingCurrent:      app.IncomingCurrent,
			CreatorLogin:         app.Creator.Login,
			ModeratorLogin:       moderatorLogin,
			CalculatedPowerCount: app.CalculatedPowerCount,
		})
	}

	h.successResponse(ctx, gin.H{
		"bid_ups": response,
	})
}

// GetComponents2 godoc
// @Summary Get all components (alternative endpoint)
// @Description Get list of all components - alternative version
// @Tags Components
// @Produce json
// @Success 200 {array} ds.Component
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/component [get]
func (h *Handler) GetComponents2(ctx *gin.Context) {
	nameFilter := ctx.Query("query") // получаем параметр query для фильтрации по имени

	components, err := h.Repository.GetAllComponents(nameFilter)
	if err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, "Ошибка получения компонентов: "+err.Error())
		return
	}

	h.successResponse(ctx, gin.H{
		"Components": components,
	})
}

// @Summary Сформировать заявку ИБП
// @Description Завершает формирование заявки на расчет ИБП
// @Tags bidUPS
// @Produce json
// @Param id path int true "ID заявки"
// @Success 200 {object} map[string]string "Заявка успешно сформирована"
// @Failure 400 {object} map[string]string "Неверный ID заявки"
// @Failure 500 {object} map[string]string "Ошибка формирования заявки"
// @Router /api/bidUPS/{id}/form [put]
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

// @Summary Отклонить заявку ИБП
// @Description Отклоняет заявку на расчет ИБП с указанием модератора и статуса
// @Tags bidUPS
// @Accept json
// @Produce json
// @Param id path int true "ID заявки"
// @Param request body object true "Данные для отклонения заявки" example:{"moderator_id": 1, "status": "declined"}
// @Success 200 {object} map[string]string "Заявка успешно отклонена"
// @Failure 400 {object} map[string]string "Неверный ID заявки или формат данных"
// @Failure 500 {object} map[string]string "Ошибка отклонения заявки"
// @Router /api/bidUPS/{id}/decline [put]
func (h *Handler) ProcessBidUPS(ctx *gin.Context) {
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

	if err := h.Repository.ProcessBidUPS(applicationID, request.ModeratorID, request.Status); err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, "Ошибка отклонения заявки: "+err.Error())
		return
	}

	h.successResponse(ctx, gin.H{
		"message": "Заявка успешно обработана",
	})
}

// @Summary Удалить заявку ИБП
// @Description Полностью удаляет заявку на расчет ИБП
// @Tags bidUPS
// @Accept json
// @Produce json
// @Param id path int true "ID заявки"
// @Param request body object true "ID модератора" example:{"moderator_id": 1}
// @Success 200 {object} map[string]string "Заявка успешно удалена"
// @Failure 400 {object} map[string]string "Неверный ID заявки или формат данных"
// @Failure 500 {object} map[string]string "Ошибка удаления заявки"
// @Router /api/bidUPS/{id} [delete]
func (h *Handler) DeleteBidUPS(ctx *gin.Context) {
	applicationIDStr := ctx.Param("id")
	applicationID, err := strconv.Atoi(applicationIDStr)
	if err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный ID заявки")
		return
	}

	if err := h.Repository.DeleteBidUPS(applicationID); err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, "Ошибка отклонения заявки: "+err.Error())
		return
	}

	h.successResponse(ctx, gin.H{
		"message": "Заявка успешно удалена",
	})
}

// @Summary Обновить данные заявки ИБП
// @Description Обновляет вес и производительность для заявки ИБП
// @Tags bidUPS
// @Accept json
// @Produce json
// @Param id path int true "ID заявки"
// @Param request body object true "Данные для обновления" example:{"creator_id": 1, "moderator_id": 2}
// @Success 200 {object} map[string]string "Изменения успешно сохранены"
// @Failure 400 {object} map[string]string "Неверный ID заявки или формат данных"
// @Failure 500 {object} map[string]string "Ошибка обновления заявки"
// @Router /api/bidUPS/{id}/set [put]
func (h *Handler) SetBidUPS(ctx *gin.Context) {
	applicationIDStr := ctx.Param("id")
	applicationID, err := strconv.Atoi(applicationIDStr)
	if err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный ID заявки")
		return
	}

	var request struct {
		IncomingCurrent int `json:"incoming_current" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный формат данных")
		return
	}

	err = h.Repository.SetBidUPS(applicationID, request.IncomingCurrent)
	if err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	h.successResponse(ctx, gin.H{
		"message": "Изменения успешно сохранены",
	})
}

// @Summary Удалить компонент из расчета
// @Description Удаляет компонент из расчета ИБП
// @Tags Calculations UPS
// @Accept json
// @Produce json
// @Param id path int true "ID ресурса"
// @Param request body object true "ID заявки" example:{"bidId": 123}
// @Success 200 {object} map[string]string "Компонент успешно удален из заявки"
// @Failure 400 {object} map[string]string "Неверный ID ресурса или формат данных"
// @Failure 500 {object} map[string]string "Ошибка удаления компонента"
// @Router /api/calcUPS [delete]
func (h *Handler) DeleteCalcUPS(ctx *gin.Context) {
	var request struct {
		ApplicationID int `json:"bidId" binding:"required"`
		ComponentID   int `json:"componentId" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный формат данных")
		return
	}

	if err := h.Repository.DeleteCalcUPS(request.ApplicationID, request.ComponentID); err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, "Ошибка удаления компонента из заявки: "+err.Error())
		return
	}

	h.successResponse(ctx, gin.H{
		"message": "Компонент успешно удален из заявки",
	})
}

// @Summary Установить параметры расчета
// @Description Устанавливает коэффициенты и мощность для расчета ИБП
// @Tags Calculations UPS
// @Accept json
// @Produce json
// @Param id path int true "ID ресурса"
// @Param request body object true "Параметры расчета" example:{"bidId": 123, "componentId": 456, "battery_life": 2, "incoming_power": 1500}
// @Success 200 {object} map[string]string "Успешно установлено"
// @Failure 400 {object} map[string]string "Неверный ID ресурса или формат данных"
// @Failure 500 {object} map[string]string "Ошибка установки коэффициента"
// @Router /api/calcUPS/{id} [put]
func (h *Handler) SetCalcUPS(ctx *gin.Context) {
	// ⭐️ УДАЛЕНО: Извлечение resourceID из ctx.Param("id") ⭐️
	// Теперь ID записи CalcUPS не требуется в пути.

	// ⭐️ ОБНОВЛЕННАЯ СТРУКТУРА ЗАПРОСА ⭐️
	var request struct {
		// ID заявки/корзины (BidUPS) - обязателен
		BidID int `json:"bid_id" binding:"required"`
		// ID компонента (Component) - обязателен для поиска в таблице CalcUPS
		ComponentID int `json:"component_id" binding:"required"`
		// Время работы (часы)
		BatteryLife int `json:"battery_life" binding:"required"`
		// Количество
		Count int `json:"count" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный формат данных: требуются 'bid_id', 'component_id', 'battery_life' и 'count'")
		return
	}

	// ⭐️ Измененный вызов репозитория: Передаем ComponentID и BidID ⭐️
	if err := h.Repository.SetCalcUPS(request.BidID, request.ComponentID, request.BatteryLife, request.Count); err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, "Ошибка обновления данных компонента: "+err.Error())
		return
	}

	h.successResponse(ctx, gin.H{
		"message": "Успешно обновлено",
	})
}

// GetBidUPSByID получает заявку UPS по ID
// @Summary Получить заявку UPS по ID
// @Description Возвращает детальную информацию о заявке UPS по указанному ID, включая компоненты
// @Tags bidUPS
// @Accept json
// @Produce json
// @Param id path integer true "ID заявки UPS" format(uint32)
// @Success 200 {object} map[string]interface{} "Успешный ответ с данными заявки"
// @Failure 400 {object} map[string]interface{} "Неверный формат ID"
// @Failure 404 {object} map[string]interface{} "Заявка не найдена"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /api/bidUPS/{id} [get]
func (h *Handler) GetBidUPSByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный ID заявки")
		return
	}

	bidUPS, err := h.Repository.GetBidUPSByID(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.errorResponse(ctx, http.StatusNotFound, "Заявка не найдена")
		} else {
			h.errorResponse(ctx, http.StatusInternalServerError, "Ошибка получения заявки: "+err.Error())
		}
		return
	}

	// Создаем response DTO с нужными полями
	type ComponentResponse struct {
		ID              uint    `json:"id"`
		Image           string  `json:"image,omitempty"`
		Title           string  `json:"title"`
		Power           int     `json:"power"`
		Coeff           float32 `json:"coeff"`
		CalculatedPower int     `json:"calculated_power"`
		Count           int     `json:"count"`
		// ⭐️ ДОБАВЛЕНО: Время работы (BatteryLife) для компонента ⭐️
		BatteryLife int `json:"battery_life"`
	}

	type BidUPSResponse struct {
		ID         uint      `json:"id"`
		Status     string    `json:"status"`
		DateUpdate time.Time `json:"date_update"`
		DateFinish time.Time `json:"date_finish,omitempty"`
		// Поле BatteryLife на уровне заявки не используется, так как оно теперь в компонентах
		// BatteryLife     int                 `json:"battery_life"`
		IncomingCurrent int                 `json:"incoming_current"`
		CreatorLogin    string              `json:"creator_login"`
		ModeratorLogin  string              `json:"moderator_login,omitempty"`
		Components      []ComponentResponse `json:"components"`
	}

	// Преобразуем компоненты в response формат
	var componentsResponse []ComponentResponse
	for _, calcUPS := range bidUPS.Components {
		componentsResponse = append(componentsResponse, ComponentResponse{
			ID:              uint(calcUPS.Component.ID),
			Image:           calcUPS.Component.Image,
			Title:           calcUPS.Component.Title,
			Power:           calcUPS.Component.Power,
			Coeff:           calcUPS.Component.Coeff,
			CalculatedPower: calcUPS.CalculatedPower,
			Count:           calcUPS.Count,
			// ⭐️ ЗАПОЛНЕНИЕ: Используем calcUPS.BatteryLife ⭐️
			BatteryLife: calcUPS.BatteryLife,
		})
	}

	// Создаем основной response
	response := BidUPSResponse{
		ID:              bidUPS.ID,
		Status:          bidUPS.Status,
		DateUpdate:      bidUPS.DateUpdate,
		IncomingCurrent: bidUPS.IncomingCurrent,
		CreatorLogin:    bidUPS.Creator.Login,
		Components:      componentsResponse,
	}

	// Добавляем опциональные поля
	if bidUPS.Moderator.ID != 0 {
		response.ModeratorLogin = bidUPS.Moderator.Login
	}
	if bidUPS.DateFinish.Valid {
		response.DateFinish = bidUPS.DateFinish.Time
	}

	h.successResponse(ctx, response)
}

const InterServiceSecretToken = "SECRET_8B"

func (h *Handler) UpdateCalculatedPower(ctx *gin.Context) {
	// 1. ПРОВЕРКА АВТОРИЗАЦИИ
	authHeader := ctx.GetHeader("Authorization")

	// Ожидаем формат "Bearer SECRET_8B"
	expectedToken := "Bearer " + InterServiceSecretToken

	if authHeader != expectedToken {
		// Отказ в доступе, если токен не совпадает
		h.errorResponse(ctx, http.StatusUnauthorized, "Неверный или отсутствующий токен авторизации")
		return
	}

	// 2. ОБРАБОТКА ТЕЛА ЗАПРОСА (остается без изменений)
	var request struct {
		BidID            int                `json:"bid_id" binding:"required"`
		Results          []ds.BidCalcResult `json:"results" binding:"required"`
		ModeratorIDFinal int                `json:"moderator_id_final" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный формат данных колбэка: "+err.Error())
		return
	}

	if err := h.Repository.UpdateCalculatedPower(request.BidID, request.Results, request.ModeratorIDFinal); err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, "Ошибка обновления расчетов: "+err.Error())
		return
	}

	h.successResponse(ctx, gin.H{"message": "Расчеты успешно обновлены"})
}
