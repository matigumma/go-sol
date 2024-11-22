package types

import "time"

type TokenInfo struct {
	Symbol    string
	Address   string
	CreatedAt string
	Score     int64
}

type TokenMeta struct {
	Name            string `json:"name"`
	Symbol          string `json:"symbol"`
	URI             string `json:"uri"`
	Mutable         bool   `json:"mutable"`
	UpdateAuthority string `json:"updateAuthority"`
}

type KnownAccounts map[string]struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type Holder struct {
	Address        string  `json:"address"`
	Amount         int64   `json:"amount"`
	Decimals       int     `json:"decimals"`
	Pct            float64 `json:"pct"`
	UiAmount       float64 `json:"uiAmount"`
	UiAmountString string  `json:"uiAmountString"`
	Owner          string  `json:"owner"`
	Insider        bool    `json:"insider"`
}

type Report struct {
	TokenMeta            TokenMeta     `json:"tokenMeta"`
	Risks                []Risk        `json:"risks"`
	TotalMarketLiquidity float64       `json:"totalMarketLiquidity"`
	TotalLPProviders     int           `json:"totalLPProviders"`
	Rugged               bool          `json:"rugged"`
	KnownAccounts        KnownAccounts `json:"knownAccounts"`
	Verification         string        `json:"verification"`
	Score                int           `json:"score"`
	FreezeAuthority      string        `json:"freezeAuthority"`
	MintAuthority        string        `json:"mintAuthority"`
	TopHolders           []Holder      `json:"topHolders"`
	DetectedAt           time.Time     `json:"detectedAt"`
}

type MintInfo struct {
	Symbol string
	Risks  []Risk
}

type Risk struct {
	Name  string
	Score int64
	Level string
}
