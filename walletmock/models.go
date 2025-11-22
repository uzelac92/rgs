package main

type WalletRequest struct {
	PlayerID  int32   `json:"player_id"`
	Amount    float64 `json:"amount"`
	RequestID string  `json:"request_id"`
	Signature string  `json:"signature"`
}

type WalletResponse struct {
	Success bool    `json:"success"`
	Balance float64 `json:"balance"`
}
