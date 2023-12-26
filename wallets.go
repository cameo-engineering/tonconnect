package tonconnect

import (
	"slices"
)

type Wallet struct {
	Name         string `json:"name"`
	UniversalURL string `json:"universal_url"`
	BridgeURL    string `json:"bridge_url"`
}

var Wallets = map[string]Wallet{
	"telegram-wallet": {
		Name:         "Wallet",
		UniversalURL: "https://t.me/wallet/start?startapp=",
		BridgeURL:    "https://bridge.tonapi.io/bridge",
	},
	"tonkeeper": {
		Name:         "Tonkeeper",
		UniversalURL: "https://app.tonkeeper.com/ton-connect",
		BridgeURL:    "https://bridge.tonapi.io/bridge",
	},
	"mytonwallet": {
		Name:         "MyTonWallet",
		UniversalURL: "https://connect.mytonwallet.org",
		BridgeURL:    "https://tonconnectbridge.mytonwallet.org/bridge",
	},
	"tonhub": {
		Name:         "Tonhub",
		UniversalURL: "https://tonhub.com/ton-connect",
		BridgeURL:    "https://connect.tonhubapi.com/tonconnect",
	},
}

func getBridgeURLs(wallets ...Wallet) []string {
	var bridges []string
	for _, w := range wallets {
		bridges = append(bridges, w.BridgeURL)
	}

	slices.Sort(bridges)
	return slices.Compact(bridges)
}
