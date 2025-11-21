package game

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

type ProvablyFairResult struct {
	ServerSeed string `json:"server_seed"`
	ClientSeed string `json:"client_seed"`
	Hash       string `json:"hash"`
	Outcome    int32  `json:"outcome"`
}

func GenerateOutcome(serverSeed, clientSeed string) ProvablyFairResult {
	h := hmac.New(sha256.New, []byte(serverSeed))
	h.Write([]byte(clientSeed))
	hash := h.Sum(nil)

	hashHex := hex.EncodeToString(hash)

	firstByte := int(hash[0])
	outcome := int32((firstByte % 6) + 1)

	return ProvablyFairResult{
		ServerSeed: serverSeed,
		ClientSeed: clientSeed,
		Hash:       hashHex,
		Outcome:    outcome,
	}
}
