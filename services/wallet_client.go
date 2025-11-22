package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"rgs/observability"
	"time"

	"go.uber.org/zap"
)

type WalletClient struct {
	baseURL string
	secret  string
	client  *http.Client
}

type walletRequest struct {
	PlayerID  int32   `json:"player_id"`
	Amount    float64 `json:"amount"`
	RequestID string  `json:"request_id"`
	Signature string  `json:"signature"`
}

type walletResponse struct {
	Success bool    `json:"success"`
	Balance float64 `json:"balance"`
}

func NewWalletClient(baseURL, secret string) *WalletClient {
	return &WalletClient{
		baseURL: baseURL,
		secret:  secret,
		client: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

func (w *WalletClient) sign(playerID int32, amount float64, requestID string) string {
	payload := fmt.Sprintf("%d:%f:%s", playerID, amount, requestID)

	mac := hmac.New(sha256.New, []byte(w.secret))
	mac.Write([]byte(payload))

	return hex.EncodeToString(mac.Sum(nil))
}

func (w *WalletClient) call(ctx context.Context, path string, playerID int32, amount float64, requestID string) (bool, error) {
	reqBody := walletRequest{
		PlayerID:  playerID,
		Amount:    amount,
		RequestID: requestID,
		Signature: w.sign(playerID, amount, requestID),
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return false, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", w.baseURL+path, bytes.NewBuffer(data))
	if err != nil {
		return false, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(httpReq)
	if err != nil {
		observability.Logger.Error("wallet http request failed", zap.Error(err))
		return false, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			observability.Logger.Error("failed to close wallet call", zap.Error(err))
		}
	}(resp.Body)

	var res walletResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		observability.Logger.Error("failed to decode wallet response", zap.Error(err))
		return false, err
	}

	return res.Success, nil
}

func (w *WalletClient) Debit(ctx context.Context, playerID int32, amount float64, requestID string) (bool, error) {
	return w.call(ctx, "/wallet/debit", playerID, amount, requestID)
}

func (w *WalletClient) Credit(ctx context.Context, playerID int32, amount float64, requestID string) (bool, error) {
	return w.call(ctx, "/wallet/credit", playerID, amount, requestID)
}
