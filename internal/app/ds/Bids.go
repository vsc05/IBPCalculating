package ds

import (
	"database/sql"
	"time"
)

type Bid struct {
	ID          uint   `gorm:"primaryKey"`
	Status      string `gorm:"type:varchar(15);not null"`
	DateUpdate  time.Time
	DateFinish  sql.NullTime `gorm:"default:null"`
	CreatorID   uint         `gorm:"not null"`
	ModeratorID uint

	Creator    User           `gorm:"foreignKey:CreatorID"`
	Moderator  User           `gorm:"foreignKey:ModeratorID"`
	Components []BidComponent `gorm:"foreignKey:BidID"`
}
