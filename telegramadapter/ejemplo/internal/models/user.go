package models

import "time"

/*
The model describes data about the telegram user
*/

type User struct {
	ID         int64  `gorm:"primaryKey"`
	UserName   string `gorm:"type:varchar(32);"`
	TelegramID int64
	AccessHash int64 // bot hash to user

	CreatedAt time.Time

	Peers []Peer `gorm:"many2many:user_peers;"`
}
