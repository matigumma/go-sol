package monitor

import (
	"context"
	"fmt"
	"gosol/types"
	"net/http"
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
	logCh          chan *ws.LogResult
	TokenUpdates   chan []types.TokenInfo
	Ctx            context.Context
	Cancel         context.CancelFunc
}

func (app *App) StartProfilingServer() {
	go func() {
		fmt.Println("Starting pprof server on :6060")
		if err := http.ListenAndServe(":6060", nil); err != nil {
			fmt.Printf("Error starting pprof server: %v\n", err)
		}
	}()
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
		logCh:          logCh,
		Ctx:            ctx,
		Cancel:         cancel,
	}
}

func (app *App) Run() {
	// app.StartProfilingServer()
	go app.wsClient.Reconnect(app.Ctx)

	// Iniciar una goroutine para procesar los logs
	go func() {
		for {
			select {
			case <-app.Ctx.Done():
				return
			case logMsg, ok := <-app.logCh:
				if !ok {
					app.StateManager.AddStatusMessage(StatusMessage{Level: ERR, Message: "FAILED TO GET LOGS"})
					return // Exit if the channel is closed
				}
				fmt.Println("Received log message in app")
				app.logProcessor.ProcessLog(logMsg)
			}
		}
	}()
}

func (app *App) Stop() {
	app.Cancel()
	app.transactionMgr.Wait()
	close(app.StatusUpdates)
	close(app.TokenUpdates)
	close(app.logCh)
}
