package tonconnect

import (
	"context"
	"fmt"
	"strconv"

	"golang.org/x/sync/errgroup"
)

type signDataRequest struct {
	ID     string     `json:"id"`
	Method string     `json:"method"`
	Params []SignData `json:"params"`
}

type SignData struct {
	SchemaCRC uint32 `json:"schema_crc"`
	Cell      []byte `json:"cell"`
	PublicKey string `json:"publicKey,omitempty"`
}

type signDataResponse struct {
	ID     string         `json:"id"`
	Result signDataResult `json:"result,omitempty"`
	Error  *struct {
		Code    uint64 `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type signDataOpt = func(*SignData)

func (s *Session) SignData(ctx context.Context, data SignData, options ...bridgeMessageOption) (*signDataResult, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)
	msgs := make(chan bridgeMessage)

	id := s.LastRequestID + 1
	g.Go(func() error {
		req := signDataRequest{
			ID:     strconv.FormatUint(id, 10),
			Method: "signData",
			Params: []SignData{data},
		}

		err := s.sendMessage(ctx, req, "signData", options...)
		if err == nil {
			s.LastRequestID = id
		}

		return err
	})

	var res signDataResult
	g.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case msg := <-msgs:
				msgID, err := msg.Message.ID.Int64()
				if err == nil {
					s.LastRequestID = uint64(msgID)
				}
				if int64(id) == msgID {
					if msg.Message.Error != nil {
						if msg.Message.Error.Message != "" {
							return fmt.Errorf("tonconnect: %s", msg.Message.Error.Message)
						}

						switch msg.Message.Error.Code {
						case 1:
							return fmt.Errorf("tonconnect: bad request")
						case 100:
							return fmt.Errorf("tonconnect: unknown app")
						case 300:
							return fmt.Errorf("tonconnect: user declined the signature request")
						case 400:
							return fmt.Errorf("tonconnect: %q method is not supported", "signData")
						default:
							return fmt.Errorf("tonconnect: unknown data sign error")
						}
					}

					cancel()

					var ok bool
					res, ok = msg.Message.Result.(signDataResult)
					if !ok {
						return fmt.Errorf("tonconnect: data sign result expected to be of type %q", "signDataResult")
					}

					return nil
				}
			}
		}
	})

	g.Go(func() error {
		return s.connectToBridge(ctx, s.BridgeURL, msgs)
	})

	err := g.Wait()

	return &res, err
}

func NewSignDataRequest(schemaCRC uint32, cell []byte, options ...signDataOpt) (*SignData, error) {
	data := &SignData{SchemaCRC: schemaCRC, Cell: cell}
	for _, opt := range options {
		opt(data)
	}

	return data, nil
}

func WithPublicKey(pubkey string) signDataOpt {
	return func(data *SignData) {
		data.PublicKey = pubkey
	}
}
