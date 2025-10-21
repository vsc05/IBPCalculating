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

type pingReq struct{}
type pingResp struct {
	Status string `json:"status"`
}

// @Summary      Show hello text
// @Description  very very friendly response
// @Tags         Tests
// @Produce      json
// @Param        name path string true "User name"
// @Success      200  {object}  pingResp
// @Router       /ping/{name} [get]
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
			return []byte("my-key"), nil
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
	router.GET("/api/getComponent/:id", h.GetComponent)
	router.GET("/api/getComponents", h.GetComponents2)
	router.GET("/api/calcups/:id", h.GetBid)
	router.POST("/api/calcups/:id/delete-component", h.DeleteComponent)
	router.POST("/api/:id/add-to-bid", h.AddComponentToBid)

	//user
	router.POST("/api/register", h.RegisterUser)
	router.GET("/api/users/:id", h.GetUser)
	router.PUT("/api/users/:id", h.SetUserChanges)
	router.POST("/login", h.LoginUser)
	router.POST("/logout", h.Logout)

	//component
	router.POST("/api/component/createComponent", h.createComponent)
	router.PUT("/api/component/:id", h.UpdateComponent)
	router.DELETE("/api/component/:id", h.DeleteComponentPostman)
	router.POST("/api/component/:id/setComponentImage", h.SetComponentImage)

	//bidUPS
	router.GET("/api/users/bidUPS", h.WithAuthCheck(role.User), h.GetUserCart)
	router.GET("/api/bidUPS", h.WithAuthCheck(role.IsModerator), h.GetBidUPS)
	router.POST("/api/bidUPS/:id", h.SetBidUPS)
	router.POST("/api/bidUPS/:id/form", h.FormBidUPS)
	router.PUT("/api/bidUPS/:id/decline", h.WithAuthCheck(role.IsModerator), h.DeclineBidUPS)
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
