package tonconnect

import (
	"encoding/json"

	"github.com/kevinburke/nacl"
)

type bridgeMessage struct {
	BrdigeURL string
	From      nacl.Key
	Message   walletMessage
}

type walletMessage struct {
	ID      json.Number `json:"id,omitempty"`
	Event   string      `json:"event,omitempty"`
	Type    string      `json:"type,omitempty"`
	Result  any         `json:"result,omitempty"`
	Payload payload     `json:"payload,omitempty"`
	Error   *struct {
		Code    uint64 `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type payload struct {
	Code    uint64             `json:"code,omitempty"`
	Message string             `json:"message,omitempty"`
	Device  deviceInfo         `json:"device,omitempty"`
	Items   []connectItemReply `json:"items,omitempty"`
}

type deviceInfo struct {
	Platform           string `json:"platform"`
	AppName            string `json:"appName"`
	AppVersion         string `json:"appVersion"`
	MaxProtocolVersion uint64 `json:"maxProtocolVersion"`
	Features           []any  `json:"features"`
}

type feature struct {
	Name        string `json:"name"`
	MaxMessages uint64 `json:"maxMessages,omitempty"`
}

type connectItemReply struct {
	Name            string `json:"name"`
	Address         string `json:"address,omitempty"`
	Network         int64  `json:"network,string,omitempty"`
	PublicKey       string `json:"publicKey,omitempty"`
	WalletStateInit []byte `json:"walletStateInit,omitempty"`
	Proof           proof  `json:"proof,omitempty"`
	Error           *struct {
		Code    uint64 `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type proof struct {
	Timestamp uint64 `json:"timestamp"`
	Domain    struct {
		LengthBytes uint64 `json:"lengthBytes"`
		Value       string `json:"value"`
	} `json:"domain"`
	Signature []byte `json:"signature"`
	Payload   string `json:"payload"`
}

type signDataResult struct {
	Signature []byte `json:"signature"`
	Timestamp uint64 `json:"timestamp"`
}
