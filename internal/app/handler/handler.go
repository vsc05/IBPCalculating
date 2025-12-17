package handler

import (
	"DIA3Course/internal/app/config"
	"DIA3Course/internal/app/ds"
	myredis "DIA3Course/internal/app/redis"
	"DIA3Course/internal/app/repository"
	"DIA3Course/internal/app/role"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	goredis "github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func (h *Handler) Ping(gCtx *gin.Context) {
	gCtx.JSON(http.StatusOK, gin.H{"status": "Hello!"})
}

type Handler struct {
	Repository *repository.Repository
	Redis      *myredis.Client
	Config     *config.Config
}

func NewHandler(r *repository.Repository, redisClient *myredis.Client, cfg *config.Config) *Handler {
	return &Handler{
		Repository: r,
		Redis:      redisClient,
		Config:     cfg,
	}
}

const jwtPrefix = "Bearer "

func (h *Handler) WithAuthCheck(allowedRoles ...role.Role) gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		jwtStr := gCtx.GetHeader("Authorization")
		if !strings.HasPrefix(jwtStr, jwtPrefix) {
			gCtx.AbortWithStatus(http.StatusForbidden)
			return
		}

		// убираем "Bearer "
		jwtStr = jwtStr[len(jwtPrefix):]

		err := h.Redis.CheckJWTInBlacklist(gCtx.Request.Context(), jwtStr)
		if err == nil { // значит что токен в блеклисте
			gCtx.AbortWithStatus(http.StatusForbidden)
			return
		}
		if !errors.Is(err, goredis.Nil) { // значит что это не ошибка отсуствия - внутренняя ошибка
			gCtx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		token, err := jwt.ParseWithClaims(jwtStr, &ds.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(h.Config.JWT.Secret), nil
		})
		if err != nil {
			log.Println("JWT parse error:", err)
			gCtx.AbortWithStatus(http.StatusForbidden)
			return
		}

		claims, ok := token.Claims.(*ds.JWTClaims)
		if !ok || !token.Valid {
			gCtx.AbortWithStatus(http.StatusForbidden)
			return
		}

		// Сохраняем логин пользователя в контексте
		gCtx.Set("userLogin", claims.Login)
		gCtx.Set("userUUID", claims.UserUUID.String())
		gCtx.Set("userID", claims.UserDBID) // <-- Сохраняем User.ID в контекст

		// Проверяем, разрешена ли роль (IsModerator)
		for _, allowed := range allowedRoles {
			if claims.IsModerator == bool(allowed) {
				gCtx.Next()
				return
			}
		}

		log.Printf("Access denied for role (IsModerator=%v), allowed: %v", claims.IsModerator, allowedRoles)
		gCtx.AbortWithStatus(http.StatusForbidden)
	}
}

func (h *Handler) RegisterHandler(router *gin.Engine) {
	router.GET("/ping", h.WithAuthCheck(role.IsModerator), h.Ping)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/api/", h.GetComponents)
	router.GET("/api/getComponents", h.GetComponents2)
	router.GET("/api/calcups/:id", h.GetBid)
	router.POST("/api/calcups/:id/delete-component", h.DeleteComponent)
	router.GET("/api/UPSbid", h.GetUserCart2)

	//user
	router.POST("/api/register", h.RegisterUser)
	router.GET("/api/users/:id", h.GetUser)
	router.PUT("/api/users/:id", h.SetUserChanges)
	router.POST("/api/login", h.LoginUser)
	router.POST("/api/logout", h.Logout)

	//component
	router.GET("/api/component/:id", h.GetComponent)
	router.GET("/api/component", h.GetComponents2)
	router.POST("/api/component", h.WithAuthCheck(role.User), h.createComponent)
	router.PUT("/api/component/:id", h.WithAuthCheck(role.User), h.UpdateComponent)
	router.DELETE("/api/component/:id", h.WithAuthCheck(role.User), h.DeleteComponentPostman)
	router.POST("/api/component/:id/setComponentImage", h.WithAuthCheck(role.User), h.SetComponentImage)
	router.POST("/api/component/:id", h.WithAuthCheck(role.User), h.AddComponentToBid)

	//bidUPS
	router.GET("/api/bidUPS", h.WithAuthCheck(role.User), h.GetUserCart)
	router.GET("/api/bidUPS/:id", h.GetBidUPSByID)
	router.GET("/api/bidUPSAll", h.GetBidUPS)
	router.PUT("/api/bidUPS/:id", h.SetBidUPS)
	router.PUT("/api/bidUPS/:id/form", h.FormBidUPS)
	router.PUT("/api/bidUPS/:id/decline", h.ProcessBidUPS)
	router.DELETE("/api/bidUPS/:id", h.DeleteBidUPS)
	router.PUT("/api/bidUPS/updateups", h.UpdateCalculatedPower)

	//calcUPS
	router.DELETE("/api/calcUPS", h.DeleteCalcUPS)
	router.PUT("/api/calcUPS", h.SetCalcUPS)
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
