package main

import (
	"Lab1/internal/app/ds"
	"Lab1/internal/app/dsn"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	_ = godotenv.Load()
	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	err = db.AutoMigrate(
		&ds.Component{},
		&ds.User{},
		&ds.Bid{},
		&ds.BidComponent{},
	)
	if err != nil {
		panic("cant migrate db")
	}
}
