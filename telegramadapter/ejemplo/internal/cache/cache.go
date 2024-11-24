package cache

import (
	"fmt"
	"tg_reader_bot/internal/app"
	"tg_reader_bot/internal/models"

	"github.com/gotd/td/tg"
)

func (manager *PeersManager) LoadUsersData() {
	db := app.GetDatabase()

	var users []models.User
	db.Find(&users)
	for _, user := range users {
		manager.addToCacheUser(&user)
	}

	var peers []models.Peer
	db.Preload("Users").Find(&peers)
	for _, peer := range peers {
		db.Model(&peer).Association("Users").Find(&peer.Users)
		manager.addToCachePeers(&peer)
	}

	var keyWords []models.KeyWords
	db.Preload("Peer").Preload("User").Find(&keyWords)
	for _, keyWord := range keyWords {
		manager.addToCacheKeyWords(&keyWord)
	}

	fmt.Println("Cache loaded")
}

func (manager *PeersManager) createUser(username string, telegramID int64, accessHash int64) *UserData {
	db := app.GetDatabase()

	assignData := models.User{
		UserName:   username,
		TelegramID: telegramID,
		AccessHash: accessHash,
	}

	var dbuser models.User
	db.Where(models.User{TelegramID: telegramID}).Assign(assignData).FirstOrCreate(&dbuser)

	user := &UserData{
		DatabaseID: dbuser.ID,
		TelegramID: telegramID,
		AccessHash: accessHash,
		Peers:      make(map[int64]*PeerData),
	}
	manager.Users[telegramID] = user

	return user
}

func (manger *PeersManager) AddPeerToUser(user *UserData, tgpeer *tg.Channel) error {
	db := app.GetDatabase()

	var peerInfo models.Peer

	peerName := tgpeer.Username
	if len(peerName) == 0 {
		peerName = tgpeer.Usernames[0].Username
	}

	peerCache, peerInCache := manger.Peers[tgpeer.ID]
	if !peerInCache {
		peerInfo = models.Peer{
			UserName:   peerName,
			Title:      tgpeer.Title,
			TelegramID: tgpeer.ID,
			AccessHash: tgpeer.AccessHash,
		}
		if err := db.Create(&peerInfo).Error; err != nil {
			return err
		}
	} else {
		peerInfo = models.Peer{ID: peerCache.DatabaseID}
	}

	manger.Mutex.Lock()

	if !peerInCache {
		peerCache = manger.addToCachePeers(&peerInfo)
	}

	userCache := manger.Users[user.TelegramID]

	_, userHavePeer := userCache.Peers[tgpeer.ID]
	if !userHavePeer {
		userCache.Peers[tgpeer.ID] = peerCache
	}

	manger.Mutex.Unlock()

	if !userHavePeer {
		db.Model(&models.User{ID: userCache.DatabaseID}).Association("Peers").Append(&peerInfo)
	}

	return nil
}

func (manager *PeersManager) RemovePeerFromUser(user *UserData, peer *PeerData) error {
	db := app.GetDatabase()

	dbpeer := models.Peer{ID: peer.DatabaseID}
	dbuser := models.User{ID: user.DatabaseID}

	if err := db.Model(&dbuser).Association("Peers").Delete(&dbpeer); err != nil {
		return err
	}

	if err := db.Where("peer_id = ? AND user_id = ?", peer.DatabaseID, user.DatabaseID).Delete(&models.KeyWords{}).Error; err != nil {
		return err
	}

	manager.Mutex.Lock()
	defer manager.Mutex.Unlock()

	delete(user.Peers, peer.TelegramID)

	if peer, ok := manager.Peers[peer.TelegramID]; ok {
		delete(peer.UsersKeyWords, user.TelegramID)
	}

	return nil
}

func (peer *PeerData) AddKeyword(user *UserData, keyword string) error {
	db := app.GetDatabase()

	usersKeywords, ok := peer.UsersKeyWords[user.TelegramID]
	if ok {
		for _, v := range usersKeywords.Keywords {
			if v == keyword {
				return nil
			}
		}
	}

	newKeyWord := models.KeyWords{
		PeerID: peer.DatabaseID,
		UserID: user.DatabaseID,
		Word:   keyword,
	}

	res := db.Create(&newKeyWord)
	if res.Error != nil {
		return res.Error
	}

	if ok {
		usersKeywords.Keywords[newKeyWord.ID] = keyword
	} else {
		usersKeywords = &PeerKeyWords{Keywords: make(map[int64]string)}
		usersKeywords.Keywords[newKeyWord.ID] = keyword
		peer.UsersKeyWords[user.TelegramID] = usersKeywords
	}

	return nil
}

func (peer *PeerData) RemoveKeyword(userID int64, keywordID int64) error {
	db := app.GetDatabase()

	if err := db.Delete(&models.KeyWords{}, keywordID).Error; err != nil {
		return err
	}

	if usersKeywords, ok := peer.UsersKeyWords[userID]; ok {
		delete(usersKeywords.Keywords, keywordID)
	}

	return nil
}

func (manager *PeersManager) GetUserData(tguser *tg.User, create bool) *UserData {
	if user, ok := manager.Users[tguser.ID]; ok {
		return user
	} else if create {
		return manager.createUser(tguser.Username, tguser.ID, tguser.AccessHash)
	}
	return nil
}

func (manager *PeersManager) addToCacheUser(user *models.User) {
	data := &UserData{
		DatabaseID: user.ID,
		TelegramID: user.TelegramID,
		AccessHash: user.AccessHash,
		Peers:      make(map[int64]*PeerData),
	}
	manager.Users[user.TelegramID] = data
}

func (manager *PeersManager) addToCachePeers(peer *models.Peer) *PeerData {
	data := &PeerData{
		TelegramID:    peer.TelegramID,
		AccessHash:    peer.AccessHash,
		DatabaseID:    peer.ID,
		UserName:      peer.UserName,
		Title:         peer.Title,
		LastMsgID:     peer.LastMessageID,
		IsChannel:     peer.IsChannel,
		CreatedAt:     peer.CreatedAt,
		UpdatedAt:     peer.UpdatedAt,
		UsersKeyWords: make(map[int64]*PeerKeyWords),
	}
	manager.Peers[peer.TelegramID] = data

	for _, userData := range peer.Users {
		user := manager.Users[userData.TelegramID]
		user.Peers[peer.TelegramID] = data
	}

	return data
}

func (manager *PeersManager) addToCacheKeyWords(keyword *models.KeyWords) {
	peerData := &keyword.Peer

	peer, ok := manager.Peers[peerData.TelegramID]
	if !ok {
		return
	}

	peerOwnerTelegramID := keyword.User.TelegramID
	peerKeyWords, ok := peer.UsersKeyWords[peerOwnerTelegramID]
	if !ok {
		peerKeyWords = &PeerKeyWords{Keywords: make(map[int64]string)}
		peer.UsersKeyWords[peerOwnerTelegramID] = peerKeyWords
	}

	peerKeyWords.Keywords[keyword.ID] = keyword.Word
}

func (peer *PeerData) GetUserKeyWords(userID int64) map[int64]string {
	if user, ok := peer.UsersKeyWords[userID]; ok {
		return user.Keywords
	}
	return nil
}

func (peer *PeerData) GetUserKeyWordsCount(userID int64) int {
	keyWords := peer.GetUserKeyWords(userID)
	if keyWords != nil {
		return len(keyWords)
	}

	return 0
}

func (user *UserData) HasPeerByID(ID int64) bool {
	_, ok := user.Peers[ID]
	return ok
}

func (user *UserData) GetActivePeerID() int64 {
	return user.ActivePeerID
}

func (user *UserData) GetActivePeer() *PeerData {
	return user.Peers[user.GetActivePeerID()]
}

func (user *UserData) GetID() int64 {
	return user.TelegramID
}
