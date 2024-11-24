package models

import (
	"time"
)

/*
The model describes the data about the peer
It can be a telegram channel/chat
PeerClass is tg.InputPeerClass
*/

type Peer struct {
	ID int64 `gorm:"primaryKey"`

	UserName string `gorm:"type:varchar(64);"`
	Title    string `gorm:"type:varchar(128);"`

	TelegramID int64
	AccessHash int64 // client hash to peer
	IsChannel  bool

	CreatedAt     time.Time
	UpdatedAt     time.Time
	LastMessageID int

	Users []User `gorm:"many2many:user_peers;"`
}
