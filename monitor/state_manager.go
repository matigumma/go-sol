package monitor

import (
	"matu/gosol/types"
	"sync"
	"time"
)

const (
	INFO LogLevel = iota
	WARN
	ERR
	NONE
	DEBUG
	TRACE
	PANIC
)

type LogLevel int

type StatusMessage struct {
	Level   LogLevel
	Message string
}

type StateManager struct {
	MintState     map[string][]types.Report
	StatusHistory []StatusMessage
	mu            sync.RWMutex
}

func NewStateManager() *StateManager {
	return &StateManager{
		MintState:     make(map[string][]types.Report),
		StatusHistory: make([]StatusMessage, 0),
	}
}

func (sm *StateManager) AddMint(mint string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if _, exists := sm.MintState[mint]; !exists {
		sm.MintState[mint] = []types.Report{}
	}
}

func (sm *StateManager) UpdateMintState(mint string, report types.Report) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.MintState[mint] = append(sm.MintState[mint], report)
}

func (sm *StateManager) SendTokenUpdates(tokenUpdates chan<- []types.TokenInfo) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var allTokens []types.TokenInfo
	for mint, reports := range sm.MintState {
		if len(reports) == 0 {
			continue
		}
		latestReport := reports[len(reports)-1]
		token := types.TokenInfo{
			Symbol:    latestReport.TokenMeta.Symbol,
			Address:   mint,
			CreatedAt: latestReport.DetectedAt.In(time.Local).Format("15:04"),
			Score:     int64(latestReport.Score),
		}
		allTokens = append(allTokens, token)
	}

	// sort.Slice(allTokens, func(i, j int) bool {
	// 	return allTokens[i].CreatedAt < allTokens[j].CreatedAt
	// })

	select {
	case tokenUpdates <- allTokens:
		// Successfully sent token updates
	default:
		sm.AddStatusMessage(StatusMessage{Level: WARN, Message: "Token updates channel is full or blocked"})
	}
}

func (sm *StateManager) AddStatusMessage(msg StatusMessage) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.StatusHistory = append(sm.StatusHistory, msg)
}

func (sm *StateManager) GetStatusHistory() []StatusMessage {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.StatusHistory
}

func (sm *StateManager) GetMintState() map[string][]types.Report {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.MintState
}
