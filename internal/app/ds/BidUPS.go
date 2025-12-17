package ds

import (
	"database/sql"
	"time"
)

type BidUPS struct {
	ID          uint         `gorm:"primaryKey" json:"id"`
	Status      string       `json:"status"`
	DateUpdate  time.Time    `json:"date_update"`
	DateFinish  sql.NullTime `gorm:"default:null" json:"date_finish"`
	CreatorID   uint         `gorm:"not null" json:"creator_id"`
	ModeratorID uint         `gorm:"default:null" json:"moderator_id"`

	IncomingCurrent int `gorm:"type:int;default:0"; json:"incoming_power"`

	Creator    User      `gorm:"foreignKey:CreatorID" json:"creator"`
	Moderator  User      `gorm:"foreignKey:ModeratorID" json:"moderator"`
	Components []CalcUPS `gorm:"foreignKey:BidID" json:"components"`

	CalculatedPowerCount int `gorm:"-" json:"calculated_power_count"`
}
