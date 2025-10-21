package ds

type User struct {
	ID          uint   `gorm:"primary_key" json:"id"`
	Login       string `gorm:"type:varchar(25);unique;not null" json:"login"`
	Password    string `gorm:"type:varchar(100);not null" json:"-"`
	IsModerator bool   `gorm:"type:boolean;default:false" json:"is_moderator"`
}

type RegisterRequest struct {
	Login       string `json:"login" binding:"required"`
	Password    string `json:"password" binding:"required"`
	IsModerator bool   `json:"isModerator,omitempty"`
}

type UpdateUserRequest struct {
	Login       string `json:"login,omitempty"`
	IsModerator bool   `json:"isModerator,omitempty"`
}


