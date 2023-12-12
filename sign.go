package tonconnect

import "context"

type signDataRequest struct {
	ID     string `json:"id"`
	Method string `json:"method"`
	Params []struct {
	} `json:"params"`
}

func (s *Session) SignData(ctx context.Context) {

}
