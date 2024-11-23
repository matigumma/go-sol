package monitor

import (
	"context"
	"gosol/types"
	"os"
)

var websocketURL string
var apiKey string

type App struct {
	wsClient       *WebSocketClient
	logProcessor   *LogProcessor
	transactionMgr *TransactionManager
	apiClient      *APIClient
	stateManager   *StateManager
	statusUpdates  chan StatusMessage
	tokenUpdates   chan []types.TokenInfo
	ctx            context.Context
	cancel         context.CancelFunc
}

func NewApp(pubkey, apiBaseURL string) *App {
	ctx, cancel := context.WithCancel(context.Background())

	statusCh := make(chan StatusMessage, 100)
	tokenCh := make(chan []types.TokenInfo, 100)

	websocketURL = os.Getenv("WEBSOCKET_URL")
	apiKey = os.Getenv("API_KEY")

	stateMgr := NewStateManager()
	apiCli := NewAPIClient(apiBaseURL, stateMgr, statusCh, tokenCh)
	transMgr := NewTransactionManager(apiCli, stateMgr, statusCh, tokenCh)
	logProc := NewLogProcessor(transMgr, statusCh)
	wsCli := NewWebSocketClient(websocketURL, apiKey, pubkey, transMgr.logCh, statusCh)

	return &App{
		wsClient:       wsCli,
		logProcessor:   logProc,
		transactionMgr: transMgr,
		apiClient:      apiCli,
		stateManager:   stateMgr,
		statusUpdates:  statusCh,
		tokenUpdates:   tokenCh,
		ctx:            ctx,
		cancel:         cancel,
	}
}

func (app *App) Run() {
	go app.wsClient.Reconnect(app.ctx)
	// Aquí puedes iniciar otras rutinas necesarias
}

func (app *App) Stop() {
	app.cancel()
	app.transactionMgr.Wait()
	close(app.statusUpdates)
	close(app.tokenUpdates)
}
