package tonconnect

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

type ConnectRequest struct {
	ManifestURL string        `json:"manifestUrl"`
	Items       []ConnectItem `json:"items"`
}

type ConnectItem struct {
	Name    string `json:"name"`
	Payload string `json:"payload,omitempty"`
}

type connReqOpt = func(*ConnectRequest)

type linkOptions struct {
	ReturnStrategy string
}

type linkOption = func(*linkOptions)

const (
	wrapURL string = "https://ton-connect.github.io/open-tc"
)

func NewConnectRequest(manifestURL string, options ...connReqOpt) (*ConnectRequest, error) {
	connReq := &ConnectRequest{
		ManifestURL: manifestURL,
	}
	connReq.Items = append(connReq.Items, ConnectItem{Name: "ton_addr"})

	for _, opt := range options {
		opt(connReq)
	}

	return connReq, nil
}

func WithProofRequest(payload string) connReqOpt {
	return func(connReq *ConnectRequest) {
		connReq.Items = append(connReq.Items, ConnectItem{Name: "ton_proof", Payload: payload})
	}
}

func (s *Session) GenerateUniversalLink(wallet Wallet, connreq ConnectRequest, options ...linkOption) (string, error) {
	opts := &linkOptions{ReturnStrategy: "back"}
	for _, opt := range options {
		opt(opts)
	}

	u, err := url.Parse(wallet.UniversalURL)
	if err != nil {
		return "", fmt.Errorf("tonconnect: failed to parse %q wallet universal URL: %w", wallet.Name, err)
	}

	q := u.Query()
	q.Set("v", "2")
	q.Set("id", hex.EncodeToString(s.ID[:]))

	data, err := json.Marshal(connreq)
	if err != nil {
		return "", fmt.Errorf("tonconnect: failed to marshal connection request: %w", err)
	}
	q.Set("r", string(data))

	q.Set("ret", opts.ReturnStrategy)

	rawQuery := q.Encode()
	if isTelegramURL(u) {
		clear(q)
		q.Set("startapp", "tonconnect-"+encodeTelegramURLParams(rawQuery))
		rawQuery = q.Encode()
	}
	u.RawQuery = rawQuery

	link := u.String()
	// HACK:
	if u.Scheme == "tc" {
		link = strings.Replace(link, ":?", "://?", 1)
	}

	return link, nil
}

func (s *Session) GenerateDeeplink(connreq ConnectRequest, options ...linkOption) (string, error) {
	w := Wallet{UniversalURL: `tc://`}

	return s.GenerateUniversalLink(w, connreq, options...)
}

func WrapDeeplink(link string) string {
	link = url.QueryEscape(link)
	return fmt.Sprintf("%s?connect=%s", wrapURL, link)
}

func WithBackReturnStrategy() linkOption {
	return func(opts *linkOptions) {
		opts.ReturnStrategy = "back"
	}
}

func WithNoneReturnStrategy() linkOption {
	return func(opts *linkOptions) {
		opts.ReturnStrategy = "none"
	}
}

func WithURLReturnStrategy(url string) linkOption {
	return func(opts *linkOptions) {
		opts.ReturnStrategy = url
	}
}

func isTelegramURL(u *url.URL) bool {
	return u.Scheme == "tg" || u.Hostname() == "t.me"
}

func encodeTelegramURLParams(params string) string {
	params = strings.ReplaceAll(params, ".", "%2E")
	params = strings.ReplaceAll(params, "-", "%2D")
	params = strings.ReplaceAll(params, "_", "%5F")
	params = strings.ReplaceAll(params, "&", "-")
	params = strings.ReplaceAll(params, "=", "__")
	params = strings.ReplaceAll(params, "%", "--")

	return params
}
