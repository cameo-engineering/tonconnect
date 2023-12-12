package tonconnect

import "context"

type sendTransactionRequest struct {
	ID     string `json:"id"`
	Method string `json:"method"`
	Params []struct {
	} `json:"params"`
}

func (s *Session) SendTransaction(ctx context.Context) {

}
