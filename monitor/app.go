package monitor

import (
	"context"
	"gosol/types"
	"os"

	"github.com/gagliardetto/solana-go/rpc/ws"
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
	logCh          chan *ws.LogResult
	tokenUpdates   chan []types.TokenInfo
	ctx            context.Context
	cancel         context.CancelFunc
}

func NewApp() *App {
	ctx, cancel := context.WithCancel(context.Background())

	statusCh := make(chan StatusMessage, 100)
	tokenCh := make(chan []types.TokenInfo, 100)
	logCh := make(chan *ws.LogResult, 100)

	pubkey := os.Getenv("RAY_FEE_PUBKEY")
	apiBaseURL := os.Getenv("API_BASE_URL")

	websocketURL = os.Getenv("WEBSOCKET_URL")
	apiKey = os.Getenv("API_KEY")

	stateMgr := NewStateManager()
	apiCli := NewAPIClient(apiBaseURL, stateMgr, statusCh, tokenCh)
	transMgr := NewTransactionManager(apiCli, stateMgr, statusCh, tokenCh)
	logProc := NewLogProcessor(transMgr, statusCh)
	wsCli := NewWebSocketClient(websocketURL, apiKey, pubkey, logCh, statusCh)

	return &App{
		wsClient:       wsCli,
		logProcessor:   logProc,
		transactionMgr: transMgr,
		apiClient:      apiCli,
		stateManager:   stateMgr,
		statusUpdates:  statusCh,
		tokenUpdates:   tokenCh,
		logCh:          logCh,
		ctx:            ctx,
		cancel:         cancel,
	}
}

func (app *App) Run() {
	go app.wsClient.Reconnect(app.ctx)
	// Aquí puedes iniciar otras rutinas necesarias

	// Iniciar una goroutine para procesar los logs
	go func() {
		for {
			select {
			case <-app.ctx.Done():
				return
			case logMsg := <-app.logCh:
				app.logProcessor.ProcessLog(logMsg)
			}
		}
	}()
}

func (app *App) Stop() {
	app.cancel()
	app.transactionMgr.Wait()
	close(app.statusUpdates)
	close(app.tokenUpdates)
	close(app.logCh)
}
