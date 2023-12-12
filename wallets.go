package tonconnect

type Wallet struct {
	Name         string `json:"name"`
	UniversalURL string `json:"universal_url"`
	BridgeURL    string `json:"bridge_url"`
}
