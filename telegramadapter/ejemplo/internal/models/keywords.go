package models

import "time"

/*
The model describes keywords for the peer
*/

type KeyWords struct {
	ID int64 `gorm:"primaryKey"`

	PeerID int64
	Peer   Peer `gorm:"constraint:OnUpdate:CASCADE,OnDelete:NO ACTION;"`

	UserID int64
	User   User `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:NO ACTION;"`

	Word      string
	CreatedAt time.Time
}
