package tonconnect

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/kevinburke/nacl"
	"golang.org/x/sync/errgroup"
)

type bridgeMessage struct {
	BrdigeURL string
	From      nacl.Key
	Message   walletMessage
}

type connectResponse struct {
	Device deviceInfo         `json:"device,omitempty"`
	Items  []connectItemReply `json:"items,omitempty"`
}

type disconnectRequest struct {
	ID     string `json:"id"`
	Method string `json:"method"`
	Params []any  `json:"params"`
}

func (s *Session) Connect(ctx context.Context, wallets ...Wallet) (*connectResponse, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)
	msgs := make(chan bridgeMessage)

	res := &connectResponse{}
	g.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case msg := <-msgs:
				if msg.Message.Event == "connect" {
					cancel()

					msgID, _ := msg.Message.ID.Int64()
					s.LastRequestID = uint64(msgID)
					s.ClientID = msg.From
					s.BridgeURL = msg.BrdigeURL

					var err error
					res.Items, err = getConnectItems(msg.Message.Payload.Items...)
					res.Device = msg.Message.Payload.Device
					return err
				} else if msg.Message.Event == "connect_error" {
					return getConnectError(msg.Message.Payload)
				}
			}
		}
	})

	for _, u := range getBridgeURLs(wallets...) {
		u := u

		g.Go(func() error {
			return s.connectToBridge(ctx, u, msgs)
		})
	}

	err := g.Wait()

	return res, err
}

func (s *Session) Disconnect(ctx context.Context, options ...bridgeMessageOption) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)
	msgs := make(chan bridgeMessage)

	id := s.LastRequestID + 1
	g.Go(func() error {
		req := disconnectRequest{
			ID:     strconv.FormatUint(id, 10),
			Method: "disconnect",
		}

		err := s.sendMessage(ctx, req, "", options...)
		if err == nil {
			s.LastRequestID = id
		}

		return err
	})

	g.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case msg := <-msgs:
				msgID, _ := msg.Message.ID.Int64()

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
						case 400:
							return fmt.Errorf("tonconnect: %q method is not supported", "sendTransaction")
						default:
							return fmt.Errorf("tonconnect: unknown disconnection error")
						}
					}
				}
			}
		}
	})

	g.Go(func() error {
		return s.connectToBridge(ctx, s.BridgeURL, msgs)
	})

	err := g.Wait()

	return err
}

func getConnectError(payload payload) error {
	if payload.Message != "" {
		return fmt.Errorf("tonconnect: %s", payload.Message)
	}

	switch payload.Code {
	case 1:
		return fmt.Errorf("tonconnect: bad request")
	case 2:
		return fmt.Errorf("tonconnect: app manifest not found")
	case 3:
		return fmt.Errorf("tonconnect: app manifest content error")
	case 100:
		return fmt.Errorf("tonconnect: unknown app")
	case 300:
		return fmt.Errorf("tonconnect: user declined the connection")
	default:
		return fmt.Errorf("tonconnect: unknown connection error")
	}
}

func getConnectItems(items ...connectItemReply) ([]connectItemReply, error) {
	var errs []error
	var res []connectItemReply
	for _, item := range items {
		if item.Error != nil {
			if item.Error.Message != "" {
				errs = append(errs, fmt.Errorf("tonconnect: %s", item.Error.Message))
			} else {
				switch item.Error.Code {
				case 400:
					errs = append(errs, fmt.Errorf("tonconnect: %q method is not supported", item.Name))
				default:
					errs = append(errs, fmt.Errorf("tonconnect: %q method unknown error", item.Name))
				}
			}
		} else {
			res = append(res, item)
		}
	}

	return res, errors.Join(errs...)
}
