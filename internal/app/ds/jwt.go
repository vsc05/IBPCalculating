package ds

import (
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

type JWTClaims struct {
	jwt.StandardClaims           // все что точно необходимо по RFC
	UserDBID           uint      `json:"user_db_id"` // <-- Добавлено
	UserUUID           uuid.UUID `json:"user_uuid"`  // наши данные - uuid этого пользователя в базе данных
	Scopes             []string  `json:"scopes"`     // список доступов в нашей системе
	IsModerator        bool
	Login              string
}

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginResponse struct {
	ExpiresIn   int         `json:"expires_in"`
	AccessToken string      `json:"access_token"`
	TokenType   string      `json:"token_type"`
	User        interface{} `json:"user"`
}
