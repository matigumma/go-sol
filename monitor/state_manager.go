package monitor

import (
	"gosol/types"
	"sort"
	"sync"
	"time"
)

const (
	INFO LogLevel = iota
	WARN
	ERR
	NONE
)

type LogLevel int

type StatusMessage struct {
	Level   LogLevel
	Message string
}

type StateManager struct {
	mintState     map[string][]types.Report
	statusHistory []StatusMessage
	mu            sync.RWMutex
}

func NewStateManager() *StateManager {
	return &StateManager{
		mintState:     make(map[string][]types.Report),
		statusHistory: make([]StatusMessage, 0),
	}
}

func (sm *StateManager) AddMint(mint string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if _, exists := sm.mintState[mint]; !exists {
		sm.mintState[mint] = []types.Report{}
	}
}

func (sm *StateManager) UpdateMintState(mint string, report types.Report) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.mintState[mint] = append(sm.mintState[mint], report)
}

func (sm *StateManager) SendTokenUpdates(tokenUpdates chan<- []types.TokenInfo) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var allTokens []types.TokenInfo
	for mint, reports := range sm.mintState {
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

	sort.Slice(allTokens, func(i, j int) bool {
		return allTokens[i].CreatedAt < allTokens[j].CreatedAt
	})

	tokenUpdates <- allTokens
}

func (sm *StateManager) AddStatusMessage(msg StatusMessage) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.statusHistory = append(sm.statusHistory, msg)
}

func (sm *StateManager) GetStatusHistory() []StatusMessage {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.statusHistory
}
