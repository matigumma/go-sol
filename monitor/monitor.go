package monitor

// // type LogLevel int

// const (
// 	INFO LogLevel = iota
// 	WARN
// 	ERR
// 	NONE
// )

// // type StatusMessage struct {
// // 	Level   LogLevel
// // 	Message string
// // }

// // var statusHistory []StatusMessage
// // var websocketURL string
// // var apiKey string

// func updateStatus(message string, level LogLevel, statusUpdates chan<- StatusMessage) {
// 	statusMessage := StatusMessage{Level: level, Message: message}
// 	statusUpdates <- statusMessage

// 	// Agregar el mensaje al historial
// 	statusHistory = append(statusHistory, statusMessage)
// }

// func (m *Monitor) GetStatusHistory() []StatusMessage {
// 	return statusHistory
// }

// // slog.Log(context.TODO(), slog.LevelInfo, fmt.Sprintf("%s", color.New(color.BgHiBlue).SprintFunc()(status)), time.Now().Format("15:04"))
// // updateStatus(status, INFO, statusUpdates)

// // Reproduce el sonido de alerta cuando se detecta un nuevo token
// func (m *Monitor) playAlertSound() {
// 	cmd := exec.Command("say", "New token")
// 	err := cmd.Run()
// 	if err != nil {
// 		updateStatus(fmt.Sprintf("Error executing say command: %v", err), ERR, m.statusUpdates)
// 	}
// }

// var mintState = make(map[string][]types.Report)

// func (m *Monitor) GetMintState() map[string][]types.Report {
// 	return mintState
// }

// type Monitor struct {
// 	tokenUpdates  chan<- []types.TokenInfo
// 	statusUpdates chan<- StatusMessage
// }

// func NewMonitor(tokenUpdates chan<- []types.TokenInfo, statusUpdates chan<- StatusMessage) *Monitor {
// 	websocketURL = os.Getenv("WEBSOCKET_URL")
// 	apiKey = os.Getenv("API_KEY")
// 	return &Monitor{tokenUpdates: tokenUpdates, statusUpdates: statusUpdates}
// }

// func (m *Monitor) Run() {
// 	client, err := m.connectToWebSocket()
// 	if err != nil {
// 		updateStatus(fmt.Sprintf("Failed to connect to WebSocket: %v", err), ERR, m.statusUpdates)
// 	}
// 	defer client.Close()

// 	pubkey := "7YttLkHDoNj9wyDur5pM1ejNaAvT9X4eqaYcHQqtj2G5" // ray_fee_pubkey
// 	updateStatus("Subscribing to logs...", NONE, m.statusUpdates)

// 	for {
// 		if err := m.subscribeToLogs(client, pubkey); err != nil {
// 			updateStatus(fmt.Sprintf("Failed to subscribe to logs: %v", err), ERR, m.statusUpdates)
// 			time.Sleep(5 * time.Second) // Wait before retrying
// 			continue                    // Retry subscription
// 		}
// 		break // Exit loop if subscription is successful
// 	}
// }

// type model struct {
// 	table table.Model
// }

// func newModel(tokens []types.TokenInfo) model {
// 	columns := []table.Column{
// 		{Title: "SYMBOL", Width: 20},
// 		{Title: "ADDRESS", Width: 10},
// 		{Title: "CREATED AT", Width: 20},
// 		{Title: "SCORE", Width: 10},
// 		{Title: "URL", Width: 40},
// 	}

// 	rows := []table.Row{}
// 	seenAddresses := make(map[string]bool)
// 	for _, token := range tokens {
// 		if !seenAddresses[token.Address] {
// 			address := token.Address[:7]
// 			url := fmt.Sprintf("https://rugcheck.xyz/tokens/%s", token.Address)
// 			scoreColor := lipgloss.Color("2") // Green
// 			if token.Score > 2000 {
// 				scoreColor = lipgloss.Color("3") // Yellow
// 			}
// 			if token.Score > 4000 {
// 				scoreColor = lipgloss.Color("1") // Red
// 			}
// 			row := table.Row{token.Symbol, address, token.CreatedAt, lipgloss.NewStyle().Foreground(scoreColor).Render(fmt.Sprintf("%d", token.Score)), url}
// 			rows = append(rows, row)
// 			seenAddresses[token.Address] = true
// 		}
// 	}

// 	t := table.New(
// 		table.WithColumns(columns),
// 		table.WithRows(rows),
// 		table.WithFocused(true),
// 	)

// 	return model{table: t}
// }

// func (m model) Init() tea.Cmd {
// 	return nil
// }

// func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
// 	switch msg := msg.(type) {
// 	case tea.KeyMsg:
// 		switch msg.String() {
// 		case "ctrl+c", "q":
// 			return m, tea.Quit
// 		}
// 	}
// 	return m, nil
// }

// func (m model) View() string {
// 	return m.table.View()
// }

// func displayTokenTable(tokens []types.TokenInfo) {
// 	p := tea.NewProgram(newModel(tokens))
// 	if err := p.Start(); err != nil {
// 		fmt.Printf("Error running program: %v", err)
// 		os.Exit(1)
// 	}
// }

// func (m *Monitor) CheckMintAddress(mint string) (string, []types.Risk, error) {
// 	url := fmt.Sprintf("https://api.rugcheck.xyz/v1/tokens/%s/report", mint)
// 	var symbol string
// 	var risks []types.Risk

// 	for attempts := 0; attempts < 4; attempts++ {
// 		go func(attempt int) {
// 			resp, err := http.Get(url)
// 			if err != nil {
// 				updateStatus(fmt.Sprintf("Error fetching data: %v", err), ERR, m.statusUpdates)
// 				return
// 			}
// 			defer resp.Body.Close()

// 			if resp.StatusCode == http.StatusOK {
// 				var report types.Report
// 				if err := json.NewDecoder(resp.Body).Decode(&report); err != nil {
// 					return
// 				}

// 				// si es muy mala actualizar el statusbar y salir
// 				if report.Score > 8000 {
// 					updateStatus(fmt.Sprintf("💩 Token Sym:[%s]: '%s' Score[%d]", report.TokenMeta.Symbol, report.TokenMeta.Name, report.Score), NONE, m.statusUpdates)
// 					return
// 				}

// 				risks = report.Risks
// 				symbol = report.TokenMeta.Symbol

// 				// Update the in-memory state with the report
// 				// mintState[mint] = []types.Report{report}
// 				// Agregar el nuevo reporte al historial
// 				mintState[mint] = append(mintState[mint], report)

// 				// Enviar el estado completo de mintState al canal tokenUpdates
// 				var allTokens []types.TokenInfo
// 				for mint, reports := range mintState {
// 					if len(reports) == 0 {
// 						continue
// 					}
// 					// Usar el último reporte (el más reciente)
// 					latestReport := reports[len(reports)-1]
// 					token := types.TokenInfo{
// 						Symbol:    latestReport.TokenMeta.Symbol,
// 						Address:   mint,
// 						CreatedAt: latestReport.DetectedAt.In(time.Local).Format("15:04"),
// 						Score:     int64(latestReport.Score),
// 					}
// 					allTokens = append(allTokens, token)
// 				}

// 				// Ordenar allTokens por la fecha del último reporte
// 				sort.Slice(allTokens, func(i, j int) bool {
// 					return mintState[allTokens[i].Address][len(mintState[allTokens[i].Address])-1].DetectedAt.Before(
// 						mintState[allTokens[j].Address][len(mintState[allTokens[j].Address])-1].DetectedAt,
// 					)
// 				})

// 				if attempt > 0 {
// 					updateStatus(fmt.Sprintf("🌀 Attempt %d: Updating allTokens...", attempt+1), INFO, m.statusUpdates)
// 				}

// 				m.tokenUpdates <- allTokens
// 			}
// 		}(attempts)

// 		// Esperar 5 segundos entre cada intento
// 		time.Sleep(5 * time.Second)
// 	}

// 	return symbol, risks, nil
// }

// func (m *Monitor) connectToWebSocket() (*ws.Client, error) {
// 	updateStatus("Connecting to WebSocket... : "+websocketURL[:len(websocketURL)-9], NONE, m.statusUpdates)
// 	client, err := ws.Connect(context.Background(), websocketURL+apiKey)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return client, nil
// }

// func (m *Monitor) subscribeToLogs(client *ws.Client, pubkey string) error {
// 	program := solana.MustPublicKeyFromBase58(pubkey)

// 	sub, err := client.LogsSubscribeMentions(
// 		program,
// 		rpc.CommitmentConfirmed,
// 	)
// 	if err != nil {
// 		return err
// 	}
// 	defer sub.Unsubscribe()

// 	updateStatus("Start monitoring...", WARN, m.statusUpdates)
// 	for {
// 		msg, err := sub.Recv(context.Background())
// 		if err != nil {
// 			return err
// 		}
// 		m.processLogMessage(msg)
// 	}
// }

// func (m *Monitor) processLogMessage(msg *ws.LogResult) {
// 	if msg.Value.Err != nil {
// 		updateStatus(fmt.Sprintf("Transaction failed: %v", msg.Value.Err), NONE, m.statusUpdates)
// 		return
// 	}

// 	signature := msg.Value.Signature
// 	updateStatus(fmt.Sprintf("Transaction Signature: %s", signature), INFO, m.statusUpdates)

// 	rpcClient := rpc.New("https://mainnet.helius-rpc.com/?api-key=" + apiKey) // rpc.MainNetBeta_RPC: https://api.mainnet-beta.solana.com ||
// 	m.getTransactionDetails(rpcClient, signature)
// }

// func (m *Monitor) getTransactionDetails(rpcClient *rpc.Client, signature solana.Signature) {
// 	cero := uint64(0) // :/

// 	updateStatus("GetTransaction EncodingBase58...", NONE, m.statusUpdates)
// 	tx58, err := rpcClient.GetTransaction(
// 		context.TODO(),
// 		signature,
// 		&rpc.GetTransactionOpts{
// 			Encoding:                       solana.EncodingBase58,
// 			Commitment:                     rpc.CommitmentConfirmed,
// 			MaxSupportedTransactionVersion: &cero,
// 		},
// 	)
// 	if err != nil {
// 		updateStatus(fmt.Sprintf("Error GetTransaction EncodingBase58: %v", err), ERR, m.statusUpdates)
// 	}

// 	if tx58 != nil {
// 		// spew.Dump(tx58)
// 		// spew.Dump(tx58.Transaction.GetBinary())
// 		// if txJson.Meta != nil {
// 		for _, balance := range tx58.Meta.PostTokenBalances {
// 			if balance.Owner.String() == "5Q544fKrFoe6tsEbD7S8EmxGTJYAKtTVhAW5Q5pge4j1" && balance.Mint.String() != "So11111111111111111111111111111111111111112" {
// 				updateStatus("========== New Token Found =========="+fmt.Sprintf(" : %s ...", balance.Mint.String()[:7]), INFO, m.statusUpdates)
// 				m.playAlertSound()
// 				// slog.Info(color.New(color.BgHiYellow).SprintFunc()("========== New Token Found =========="))
// 				// slog.Info(color.New(color.BgHiYellow).SprintFunc()(fmt.Sprintf("Mint Address: %s", balance.Mint)))
// 				// slog.Info(color.New(color.BgHiYellow).SprintFunc()("====================================="))

// 				// Add mint address to mintState if it doesn't exist
// 				if _, exists := mintState[balance.Mint.String()]; !exists {
// 					mintState[balance.Mint.String()] = []types.Report{}
// 				}

// 				// Check mint address for additional API information
// 				go m.CheckMintAddress(balance.Mint.String())
// 			}
// 		}
// 	}
// }
