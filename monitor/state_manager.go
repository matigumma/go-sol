package monitor

import (
	"matu/gosol/types"
	"matu/gosol/storage"
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

type StateManager struct {
	IndexedMints  []string
	MintState     map[string][]types.Report
	StatusHistory []StatusMessage
	Mu            sync.RWMutex
}

func NewStateManager(dbPath string) *StateManager {
	storage, err := storage.NewStorage(dbPath)
	if err != nil {
		log.Fatalf("Error al inicializar el almacenamiento: %v", err)
	}

	mintState, err := storage.GetMintState()
	if err != nil {
		log.Fatalf("Error al obtener el estado de los mints: %v", err)
	}

	return &StateManager{
		MintState:     mintState,
		StatusHistory: make([]StatusMessage, 0),
		storage:       storage,
	}
}

func (sm *StateManager) AddMint(mint string) {
	sm.Mu.Lock()
	defer sm.Mu.Unlock()
	if _, exists := sm.MintState[mint]; !exists {
		sm.MintState[mint] = []types.Report{}
		sm.IndexedMints = append(sm.IndexedMints, mint)
	}
}

func (sm *StateManager) UpdateMintState(mint string, report types.Report) {
	sm.Mu.Lock()
	defer sm.Mu.Unlock()
	sm.MintState[mint] = append(sm.MintState[mint], report)

	// Persistir el reporte en la base de datos
	err := sm.storage.AddReport(mint, report)
	if err != nil {
		sm.AddStatusMessage(StatusMessage{Level: ERR, Message: "Error al persistir el reporte: " + err.Error()})
	}
}

func (sm *StateManager) SendTokenUpdates(tokenUpdates chan<- []types.TokenInfo) {
	sm.Mu.RLock()
	defer sm.Mu.RUnlock()

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
	sm.Mu.Lock()
	defer sm.Mu.Unlock()
	sm.StatusHistory = append(sm.StatusHistory, msg)
}

func (sm *StateManager) GetStatusHistory() []StatusMessage {
	sm.Mu.RLock()
	defer sm.Mu.RUnlock()
	return sm.StatusHistory
}

func (sm *StateManager) GetMintState() map[string][]types.Report {
	sm.Mu.RLock()
	defer sm.Mu.RUnlock()
	return sm.MintState
}
