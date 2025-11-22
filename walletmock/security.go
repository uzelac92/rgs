package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

const Secret = "testsecret123"

func ValidSignature(player int32, amount float64, reqID, sig string) bool {
	payload := fmt.Sprintf("%d:%f:%s", player, amount, reqID)

	mac := hmac.New(sha256.New, []byte(Secret))
	mac.Write([]byte(payload))
	expected := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(sig), []byte(expected))
}
