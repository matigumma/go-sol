package monitor

import (
	"encoding/json"
	"fmt"
	"matu/gosol/types"
	"net/http"
	"time"
)

type APIClient struct {
	stateManager  *StateManager
	statusUpdates chan<- StatusMessage
	tokenUpdates  chan<- []types.TokenInfo
	// requestThrottle chan struct{}
}

func NewAPIClient(stateManager *StateManager, statusUpdates chan<- StatusMessage, tokenUpdates chan<- []types.TokenInfo) *APIClient {
	return &APIClient{
		stateManager:  stateManager,
		statusUpdates: statusUpdates,
		tokenUpdates:  tokenUpdates,
		// requestThrottle: make(chan struct{}, 10), // Limitar a 10 solicitudes concurrentes
	}
}

func (api *APIClient) FetchAndProcessReport(mint string) bool {
	api.statusUpdates <- StatusMessage{Level: INFO, Message: fmt.Sprintf("fetching at %v: %v", time.Now().Format("2006-01-02 15:04:05"), mint)}

	report, err := api.fetchTokenReport(mint)
	if err != nil {
		api.statusUpdates <- StatusMessage{Level: ERR, Message: fmt.Sprintf("Error fetching report for %s: %v", mint, err)}
		return true
	}

	if report.Score > 8000 {
		api.handleHighRiskToken(report)
		return true
	}

	PushToDiscord(report, api.statusUpdates)
	api.stateManager.UpdateMintState(mint, report)
	api.stateManager.SendTokenUpdates(api.tokenUpdates)

	return true
}

func (api *APIClient) fetchTokenReport(mint string) (types.Report, error) {
	var report types.Report
	var err error
	report, err = api.tryFetchTokenReport(mint)
	if err == nil {
		return report, nil
	}
	return report, err
}

func (api *APIClient) tryFetchTokenReport(mint string) (types.Report, error) {
	url := fmt.Sprintf("%s/v1/tokens/%s/report", apiBaseURL, mint)
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
	api.statusUpdates <- StatusMessage{Level: NONE, Message: fmt.Sprintf("💩 Token Sym:[%s]: '%s' Score[%d]", report.TokenMeta.Symbol, report.TokenMeta.Name, report.Score)}
}

func (api *APIClient) RequestReportOnDemand(mint string) {
	api.FetchAndProcessReport(mint)
}
