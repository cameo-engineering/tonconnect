package tonconnect

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

type linkOptions struct {
	ReturnStrategy string
}

type linkOption = func(*linkOptions)

const (
	versionKey     string = "v"
	versionVal     string = "2"
	idKey          string = "id"
	connReqKey     string = "r"
	retStrategyKey string = "ret"

	backRetStrategyID string = "back"
	noneRetStrategyID string = "none"

	urlScheme string = "tc"
)

func (s *Session) GenerateUniversalLink(wallet Wallet, connreq connectRequest, options ...linkOption) (string, error) {
	opts := &linkOptions{ReturnStrategy: backRetStrategyID}
	for _, opt := range options {
		opt(opts)
	}

	u, err := url.Parse(wallet.UniversalURL)
	if err != nil {
		return "", fmt.Errorf("tonconnect: failed to parse %q wallet universal URL: %w", wallet.Name, err)
	}

	q := u.Query()
	q.Set(versionKey, versionVal)
	q.Set(idKey, hex.EncodeToString(s.ID[:]))

	data, err := json.Marshal(connreq)
	if err != nil {
		return "", fmt.Errorf("tonconnect: failed to marshal connection request: %w", err)
	}
	q.Set(connReqKey, string(data))

	q.Set(retStrategyKey, opts.ReturnStrategy)
	u.RawQuery = q.Encode()

	link := u.String()
	// HACK:
	if u.Scheme == urlScheme {
		link = strings.Replace(link, ":?", "://?", 1)
	}

	return link, nil
}

func (s *Session) GenerateDeeplink(connreq connectRequest, options ...linkOption) (string, error) {
	w := Wallet{UniversalURL: `tc://`}
	return s.GenerateUniversalLink(w, connreq, options...)
}

func WithBackReturnStrategy() linkOption {
	return func(opts *linkOptions) {
		opts.ReturnStrategy = backRetStrategyID
	}
}

func WithNoneReturnStrategy() linkOption {
	return func(opts *linkOptions) {
		opts.ReturnStrategy = noneRetStrategyID
	}
}

func WithURLReturnStrategy(url string) linkOption {
	return func(opts *linkOptions) {
		opts.ReturnStrategy = url
	}
}
