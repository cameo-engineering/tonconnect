package tonconnect

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
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

type bridgeMessageOptions struct {
	TTL   string
	Topic string
}

type bridgeMessageOption = func(*bridgeMessageOptions)

func NewSession() (*Session, error) {
	id, pk, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("tonconnect: failed to generate key pair: %w", err)
	}

	s := &Session{ID: id, PrivateKey: pk, LastRequestID: 1}

	return s, nil
}

func (s *Session) connectToBridge(ctx context.Context, bridgeURL string, msgs chan<- bridgeMessage) error {
	u, err := url.Parse(bridgeURL)
	if err != nil {
		return fmt.Errorf("tonconnect: failed to parse bridge URL: %w", err)
	}

	u = u.JoinPath("/events")
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
	unsub := conn.SubscribeEvent("message", func(e sse.Event) {
		var bmsg struct {
			From    string `json:"from"`
			Message []byte `json:"message"`
		}
		if err := json.Unmarshal([]byte(e.Data), &bmsg); err == nil {
			var msg walletMessage
			if clientID, err := s.decrypt(bmsg.From, bmsg.Message, &msg); err == nil {
				msgs <- bridgeMessage{BrdigeURL: bridgeURL, From: clientID, Message: msg}
				id, err := strconv.ParseUint(e.LastEventID, 10, 64)
				if err == nil {
					s.LastEventID = id
				}
			}
		}
	})
	defer unsub()

	if err := conn.Connect(); !(errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)) {
		return fmt.Errorf("tonconnect: failed to connect to bridge: %w", err)
	}

	return nil
}

func (s *Session) sendMessage(ctx context.Context, msg any, topic string, options ...bridgeMessageOption) error {
	opts := &bridgeMessageOptions{TTL: "300"}
	for _, opt := range options {
		opt(opts)
	}

	u, err := url.Parse(s.BridgeURL)
	if err != nil {
		return fmt.Errorf("tonconnect: failed to parse bridge URL: %w", err)
	}

	u = u.JoinPath("/message")
	q := u.Query()
	q.Set("client_id", hex.EncodeToString(s.ID[:]))
	q.Set("to", hex.EncodeToString(s.ClientID[:]))
	if opts.TTL != "" {
		q.Set("ttl", opts.TTL)
	}
	if topic != "" {
		q.Set("topic", topic)
	}
	u.RawQuery = q.Encode()

	data, err := s.encrypt(msg)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewBuffer([]byte(base64.StdEncoding.EncodeToString(data))))
	if err != nil {
		return fmt.Errorf("tonconnect: failed to initialize HTTP request: %w", err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("tonconnect: failed to send message: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		// TODO: parse error message
		return fmt.Errorf("tonconnect: failed to send message")
	}

	return nil
}

func (s *Session) encrypt(msg any) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("tonconnect: failed to marshal message to encrypt: %w", err)
	}

	return box.EasySeal(data, s.ClientID, s.PrivateKey), nil
}

func (s *Session) decrypt(from string, msg []byte, v any) (nacl.Key, error) {
	clientID, err := nacl.Load(from)
	if err != nil {
		return clientID, fmt.Errorf("tonconnect: failed to load client ID: %w", err)
	}

	if s.ClientID != nil && !bytes.Equal(s.ClientID[:], clientID[:]) {
		return clientID, fmt.Errorf("tonconnect: session and bridge message client IDs don't match")
	}

	data, err := box.EasyOpen(msg, clientID, s.PrivateKey)
	if err != nil {
		return clientID, fmt.Errorf("tonconnect: failed to decrypt bridge message: %w", err)
	}

	err = json.Unmarshal(data, v)
	if err != nil {
		return clientID, fmt.Errorf("tonconnect: failed to unmarshal decrypted data: %w", err)
	}

	return clientID, nil
}

func WithTTL(ttl uint64) bridgeMessageOption {
	return func(opts *bridgeMessageOptions) {
		opts.TTL = strconv.FormatUint(ttl, 10)
	}
}
