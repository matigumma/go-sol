package monitor

import (
	"context"
	"fmt"
	"matu/gosol/types"
	_ "net/http/pprof"
	"os"

	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/joho/godotenv"
)

var websocketURL string
var apiKey string
var pubkey string
var apiBaseURL string

type App struct {
	wsClient       *WebSocketClient
	logProcessor   *LogProcessor
	transactionMgr *TransactionManager
	ApiClient      *APIClient
	StateManager   *StateManager
	StatusUpdates  chan StatusMessage
	LogCh          chan *ws.LogResult
	TokenUpdates   chan []types.TokenInfo
	Ctx            context.Context
	Cancel         context.CancelFunc
}

func NewApp() *App {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	ctx, cancel := context.WithCancel(context.Background())

	statusCh := make(chan StatusMessage, 100)
	tokenCh := make(chan []types.TokenInfo, 100)
	logCh := make(chan *ws.LogResult, 100)

	pubkey = os.Getenv("RAY_FEE_PUBKEY")
	apiBaseURL = os.Getenv("API_BASE_URL")

	websocketURL = os.Getenv("WEBSOCKET_URL")
	apiKey = os.Getenv("API_KEY")

	if websocketURL == "" || apiKey == "" || pubkey == "" || apiBaseURL == "" {
		panic(fmt.Sprintf("Environment variables are not set properly: WEBSOCKET_URL=%s, API_KEY=%s, RAY_FEE_PUBKEY=%s, API_BASE_URL=%s", websocketURL, apiKey, pubkey, apiBaseURL))
	}

	stateMgr := NewStateManager()
	apiCli := NewAPIClient(stateMgr, statusCh, tokenCh)
	transMgr := NewTransactionManager(apiCli, stateMgr, statusCh, tokenCh)
	logProc := NewLogProcessor(transMgr, statusCh)
	wsCli := NewWebSocketClient(logCh, statusCh)

	return &App{
		wsClient:       wsCli,
		logProcessor:   logProc,
		transactionMgr: transMgr,
		ApiClient:      apiCli,
		StateManager:   stateMgr,
		StatusUpdates:  statusCh,
		TokenUpdates:   tokenCh,
		LogCh:          logCh,
		Ctx:            ctx,
		Cancel:         cancel,
	}
}

func (app *App) Run() {
	go app.wsClient.Reconnect(app.Ctx)
	// done := make(chan struct{})

	go func() {
		// defer close(done)
		for {
			select {
			case logMsg := <-app.LogCh:
				app.logProcessor.ProcessLog(logMsg)
			case <-app.Ctx.Done():
				return
			}
		}
	}()

	// <-done
}

func (app *App) updateStatus(message string, level LogLevel) {
	app.StatusUpdates <- StatusMessage{Level: level, Message: message}
}

func (app *App) Stop() {
	app.Cancel()
	app.transactionMgr.Wait()
	close(app.StatusUpdates)
	close(app.TokenUpdates)
	close(app.LogCh)
}
