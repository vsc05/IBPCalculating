package main

import (
	"DIA3Course/internal/app/config"
	"DIA3Course/internal/app/docs"
	"DIA3Course/internal/app/dsn"
	"DIA3Course/internal/app/handler"
	"DIA3Course/internal/app/redis"
	"DIA3Course/internal/app/repository"
	"DIA3Course/internal/pkg"
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// @title BITOP API
// @version 1.0
// @description Bmstu Open IT Platform
// @contact.name API Support
// @contact.url https://vk.com/bmstu_schedule
// @contact.email bitop@spatecon.ru
// @license.name AS IS (NO WARRANTY)
// @host 127.0.0.1:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	router := gin.Default()
	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	postgresString := dsn.FromEnv()
	fmt.Println(postgresString)

	rep, errRep := repository.New(postgresString)
	if errRep != nil {
		logrus.Fatalf("error initializing repository: %v", errRep)
	}

	ctx := context.Background()
	redisClient, err := redis.New(ctx, conf.Redis)
	if err != nil {
		logrus.Fatalf("error initializing redis: %v", err)
	}

	hand := handler.NewHandler(rep, redisClient, conf)

	docs.SwaggerInfo.Title = "UPS calc"
	docs.SwaggerInfo.Description = "API SERVER"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "localhost:8080"
	docs.SwaggerInfo.BasePath = "/"

	application := pkg.NewApp(conf, router, hand, redisClient)
	application.RunApp()
}
