package monitor

import (
	"fmt"

	"github.com/gagliardetto/solana-go/rpc/ws"
)

type LogProcessor struct {
	transactionManager *TransactionManager
	statusUpdates      chan<- StatusMessage
}

func NewLogProcessor(tm *TransactionManager, statusUpdates chan<- StatusMessage) *LogProcessor {
	return &LogProcessor{
		transactionManager: tm,
		statusUpdates:      statusUpdates,
	}
}

func (lp *LogProcessor) ProcessLog(msg *ws.LogResult) {
	lp.updateStatus("Processing log message", INFO)
	if msg.Value.Err != nil {
		lp.updateStatus(fmt.Sprintf("Transaction failed: %v", msg.Value.Err), ERR)
		return
	}

	signature := msg.Value.Signature
	lp.updateStatus(fmt.Sprintf("Transaction Signature: %s", signature), INFO)

	lp.transactionManager.HandleTransaction(signature)
}

func (lp *LogProcessor) updateStatus(message string, level LogLevel) {
	lp.transactionManager.stateManager.AddStatusMessage(StatusMessage{Level: level, Message: message})
}
