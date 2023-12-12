package tonconnect

import (
	"context"

	"golang.org/x/sync/errgroup"
)

type connectRequest struct {
	ManifestURL string        `json:"manifestUrl"`
	Items       []connectItem `json:"items"`
}

type connectItem struct {
	Name    string `json:"name"`
	Payload string `json:"payload,omitempty"`
}

type connReqOpt = func(*connectRequest)

const (
	tonAddrName  string = "ton_addr"
	tonProofName string = "ton_proof"
)

func NewConnectRequest(manifestURL string, options ...connReqOpt) (*connectRequest, error) {
	connReq := &connectRequest{
		ManifestURL: manifestURL,
	}
	connReq.Items = append(connReq.Items, connectItem{Name: tonAddrName})

	for _, opt := range options {
		opt(connReq)
	}

	return connReq, nil
}

func WithProofRequest(payload string) connReqOpt {
	return func(connReq *connectRequest) {
		connReq.Items = append(connReq.Items, connectItem{Name: tonProofName, Payload: payload})
	}
}

func (s *Session) Connect(ctx context.Context, wallets ...Wallet) error {
	g, ctx := errgroup.WithContext(ctx)
	msgs := make(chan walletMessage)
	for _, w := range wallets {
		w := w
		g.Go(func() error {
			return s.connectToBridge(ctx, w.BridgeURL, msgs)
		})
	}

	return g.Wait()
}

func (s *Session) Disconnect(ctx context.Context) error {
	return nil
}
