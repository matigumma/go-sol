package cache

import (
	"sync"
	"time"
)

const (
	StateNone = iota
	WaitingPeerName
	WaitingKeyWord
)

type PeerKeyWords struct {
	/* keyword id -> word */
	Keywords map[int64]string
}

type PeerData struct {
	TelegramID int64
	AccessHash int64

	DatabaseID int64
	UserName   string
	Title      string
	LastMsgID  int
	IsChannel  bool

	CreatedAt time.Time
	UpdatedAt time.Time

	/* telegram user id -> keywords */
	UsersKeyWords map[int64]*PeerKeyWords
}

type UserData struct {
	DatabaseID         int64
	TelegramID         int64
	AccessHash         int64
	State              uint32
	ActiveMessageID    int
	ActivePeerID       int64
	SecretButtonClicks int

	/* telegram peer id */
	Peers map[int64]*PeerData
}

type PeersManager struct {
	/* telegram peer id */
	Peers map[int64]*PeerData
	Users map[int64]*UserData

	Mutex sync.RWMutex
}
