package tonconnect

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/kevinburke/nacl"
	"github.com/kevinburke/nacl/box"
	"github.com/tmaxmax/go-sse"
)

type Session struct {
	ID            nacl.Key `json:"id"`
	PrivateKey    nacl.Key `json:"private_key"`
	ClientID      nacl.Key `json:"client_id,omitempty"`
	BridgeURL     string   `json:"brdige_url,omitempty"`
	LastEventID   uint64   `json:"last_event_id,string,omitempty"`
	LastRequestID uint64   `json:"last_request_id,string,omitempty"`
}

type bridgeMessage struct {
	From    string `json:"from"`
	Message []byte `json:"message"`
}

func NewSession() (*Session, error) {
	id, pk, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("tonconnect: failed to generate key pair: %w", err)
	}

	s := &Session{ID: id, PrivateKey: pk, LastRequestID: 1}

	return s, nil
}

func (s *Session) connectToBridge(ctx context.Context, bridgeURL string, msgs chan<- walletMessage) error {
	u, err := url.Parse(bridgeURL)
	if err != nil {
		return fmt.Errorf("tonconnect: failed to parse bridge URL: %w", err)
	}

	u.Path += "/events"
	q := u.Query()
	q.Set("client_id", hex.EncodeToString(s.ID[:]))
	if s.LastEventID > 0 {
		q.Set("last_event_id", strconv.FormatUint(uint64(s.LastEventID), 10))
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), http.NoBody)
	if err != nil {
		return fmt.Errorf("tonconnect: failed to initialize HTTP request: %w", err)
	}

	conn := sse.NewConnection(req)
	conn.SubscribeEvent("message", func(e sse.Event) {
		var bm bridgeMessage
		if err := json.Unmarshal([]byte(e.Data), &bm); err == nil {
			var msg walletMessage
			if err := s.decrypt(&bm, &msg); err == nil {
				msgs <- msg
				id, err := strconv.ParseUint(e.LastEventID, 10, 64)
				if err == nil {
					s.LastEventID = id
				}
			}
		}
	})

	if err := conn.Connect(); err != nil {
		return err
	}

	return nil
}

func (s *Session) sendMessage(bridgeURL string, msg []byte) error {
	u, _ := url.Parse(bridgeURL)
	u.Path = "message"
	q := u.Query()
	q.Set("client_id", hex.EncodeToString(s.ID[:]))
	q.Set("to", hex.EncodeToString(s.ClientID[:]))
	u.RawQuery = q.Encode()

	return nil
}

func (s *Session) encrypt(msg any) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("tonconnect: failed to marshal message: %w", err)
	}

	return box.EasySeal(data, s.ClientID, s.PrivateKey), nil
}

func (s *Session) decrypt(msg *bridgeMessage, v any) error {
	clientID, err := nacl.Load(msg.From)
	if err != nil {
		return fmt.Errorf("tonconnect: failed to load client ID: %w", err)
	}

	if s.ClientID != nil && s.ClientID != clientID {
		return fmt.Errorf("tonconnect: invalid session")
	}

	data, err := box.EasyOpen(msg.Message, clientID, s.PrivateKey)
	if err != nil {
		return fmt.Errorf("tonconnect: failed to decrypt bridge message: %w", err)
	}

	err = json.Unmarshal(data, v)
	if err != nil {
		return fmt.Errorf("tonconnect: failed to unmarshal data: %w", err)
	}

	return nil
}
