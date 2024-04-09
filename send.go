package tonconnect

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"golang.org/x/sync/errgroup"
)

type sendTransactionRequest struct {
	ID     string   `json:"id"`
	Method string   `json:"method"`
	Params []string `json:"params"`
}

type Transaction struct {
	ValidUntil uint64    `json:"valid_until,omitempty"`
	Network    string    `json:"network,omitempty"`
	From       string    `json:"from,omitempty"`
	Messages   []Message `json:"messages"`
}

type Message struct {
	Address   string `json:"address"`
	Amount    string `json:"amount"`
	Payload   []byte `json:"payload,omitempty"`
	StateInit []byte `json:"stateInit,omitempty"`
}

type sendTransactionResponse struct {
	ID     string `json:"id"`
	Result []byte `json:"result,omitempty"`
	Error  *struct {
		Code    uint64 `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type txOpt = func(*Transaction)

type msgOpt = func(*Message)

func (s *Session) SendTransaction(ctx context.Context, tx Transaction, options ...bridgeMessageOption) ([]byte, error) {
	g, ctx := errgroup.WithContext(ctx)
	msgs := make(chan bridgeMessage)

	id := s.LastRequestID + 1
	g.Go(func() error {
		tr, err := json.Marshal(tx)
		if err != nil {
			return fmt.Errorf("tonconnect: failed to marshal transaction: %w", err)
		}

		req := sendTransactionRequest{
			ID:     strconv.FormatUint(id, 10),
			Method: "sendTransaction",
			Params: []string{string(tr)},
		}

		err = s.sendMessage(ctx, req, "sendTransaction", options...)
		if err == nil {
			s.LastRequestID = id
		}

		return err
	})

	var boc []byte
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
							return fmt.Errorf("tonconnect: user declined the transaction")
						case 400:
							return fmt.Errorf("tonconnect: %q method is not supported", "sendTransaction")
						default:
							return fmt.Errorf("tonconnect: unknown transaction send error")
						}
					}

					res, ok := msg.Message.Result.(string)
					if !ok {
						return fmt.Errorf("tonconnect: transaction result expected to be of type %q", "string")
					}

					boc, err = base64.StdEncoding.DecodeString(res)
					if err != nil {
						return fmt.Errorf("tonconnect: failed to decode transaction result bag of cells")
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

	return boc, err
}

func NewTransaction(options ...txOpt) (*Transaction, error) {
	tx := &Transaction{}
	for _, opt := range options {
		opt(tx)
	}

	return tx, nil
}

func NewMessage(address string, amount string, options ...msgOpt) (*Message, error) {
	msg := &Message{Address: address, Amount: amount}
	for _, opt := range options {
		opt(msg)
	}

	return msg, nil
}

func WithTimeout(timeout time.Duration) txOpt {
	return func(tx *Transaction) {
		tx.ValidUntil = uint64(time.Now().Add(timeout).Unix())
	}
}

func WithMainnet() txOpt {
	return func(tx *Transaction) {
		tx.Network = "-239"
	}
}

func WithTestnet() txOpt {
	return func(tx *Transaction) {
		tx.Network = "-3"
	}
}

func WithFrom(from string) txOpt {
	return func(tx *Transaction) {
		tx.From = from
	}
}

func WithMessage(msg Message) txOpt {
	return func(tx *Transaction) {
		tx.Messages = append(tx.Messages, msg)
	}
}

func WithPayload(payload []byte) msgOpt {
	return func(msg *Message) {
		msg.Payload = payload
	}
}

func WithStateInit(stateInit []byte) msgOpt {
	return func(msg *Message) {
		msg.StateInit = stateInit
	}
}
