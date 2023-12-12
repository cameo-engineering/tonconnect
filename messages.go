package tonconnect

type walletMessage struct {
	ID      uint64 `json:"id,omitempty"`
	Event   string `json:"event,omitempty"`
	Result  string `json:"result,omitempty"`
	Payload struct {
		Code    uint64             `json:"code,omitempty"`
		Message string             `json:"message,omitempty"`
		Device  deviceInfo         `json:"device,omitempty"`
		Items   []connectItemReply `json:"items,omitempty"`
	} `json:"payload,omitempty"`
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
	Error           struct {
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
