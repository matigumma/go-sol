package monitor

import (
	"encoding/json"
	"fmt"
	"gosol/types"
	"net/http"
)

type APIClient struct {
	baseURL       string
	stateManager  *StateManager
	statusUpdates chan<- StatusMessage
	tokenUpdates  chan<- []types.TokenInfo
	requestThrottle chan struct{}
}

func NewAPIClient(baseURL string, stateManager *StateManager, statusUpdates chan<- StatusMessage, tokenUpdates chan<- []types.TokenInfo) *APIClient {
	return &APIClient{
		baseURL:         baseURL,
		stateManager:    stateManager,
		statusUpdates:   statusUpdates,
		tokenUpdates:    tokenUpdates,
		requestThrottle: make(chan struct{}, 10), // Limitar a 10 solicitudes concurrentes
	}

func NewAPIClient(baseURL string, stateManager *StateManager, statusUpdates chan<- StatusMessage, tokenUpdates chan<- []types.TokenInfo) *APIClient {
	return &APIClient{
		baseURL:       baseURL,
		stateManager:  stateManager,
		statusUpdates: statusUpdates,
		tokenUpdates:  tokenUpdates,
	}
}

func (api *APIClient) FetchAndProcessReport(mint string) {
	go func() {
		api.requestThrottle <- struct{}{} // Adquirir un "permiso" para hacer la solicitud
		defer func() { <-api.requestThrottle }() // Liberar el "permiso" al finalizar

		report, err := api.fetchTokenReport(mint)
		if err != nil {
			api.stateManager.AddStatusMessage(StatusMessage{Level: ERR, Message: fmt.Sprintf("Error fetching report for %s: %v", mint, err)})
			return
		}

		if report.Score > 8000 {
			api.handleHighRiskToken(report)
			return
		}

		api.stateManager.UpdateMintState(mint, report)
		api.stateManager.SendTokenUpdates(api.tokenUpdates)
	}()
}

func (api *APIClient) fetchTokenReport(mint string) (types.Report, error) {
	var report types.Report
	var err error
	for attempts := 0; attempts < 3; attempts++ {
		report, err = api.tryFetchTokenReport(mint)
		if err == nil {
			return report, nil
		}
		time.Sleep(time.Duration(attempts+1) * time.Second) // Esperar más tiempo en cada intento
	}
	return report, err
}

func (api *APIClient) tryFetchTokenReport(mint string) (types.Report, error) {
	url := fmt.Sprintf("%s/v1/tokens/%s/report", api.baseURL, mint)
	resp, err := http.Get(url)
	if err != nil {
		return types.Report{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return types.Report{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var report types.Report
	if err := json.NewDecoder(resp.Body).Decode(&report); err != nil {
		return types.Report{}, err
	}

	return report, nil
}

func (api *APIClient) handleHighRiskToken(report types.Report) {
	api.stateManager.AddStatusMessage(StatusMessage{Level: NONE, Message: fmt.Sprintf("💩 Token Sym:[%s]: '%s' Score[%d]", report.TokenMeta.Symbol, report.TokenMeta.Name, report.Score)})
}

func (api *APIClient) RequestReportOnDemand(mint string) {
	api.FetchAndProcessReport(mint)
}
