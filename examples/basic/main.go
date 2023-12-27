package main

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"log"
	"time"

	"github.com/cameo-engineering/tonconnect"
	"golang.org/x/exp/maps"
)

func main() {
	s, err := tonconnect.NewSession()
	if err != nil {
		log.Fatal(err)
	}

	data := make([]byte, 32)
	_, err = rand.Read(data)
	if err != nil {
		log.Fatal(err)
	}

	connreq, err := tonconnect.NewConnectRequest(
		"https://raw.githubusercontent.com/cameo-engineering/tonconnect/master/tonconnect-manifest.json",
		tonconnect.WithProofRequest(base32.StdEncoding.EncodeToString(data)),
	)
	if err != nil {
		log.Fatal(err)
	}

	deeplink, err := s.GenerateDeeplink(*connreq, tonconnect.WithBackReturnStrategy())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Deeplink: %s\n\n", deeplink)

	wrapped := tonconnect.WrapDeeplink(deeplink)
	fmt.Printf("Wrapped deeplink: %s\n\n", wrapped)

	for _, wallet := range tonconnect.Wallets {
		link, err := s.GenerateUniversalLink(wallet, *connreq)
		fmt.Printf("%s: %s\n\n", wallet.Name, link)
		if err != nil {
			log.Fatal(err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	res, err := s.Connect(ctx, (maps.Values(tonconnect.Wallets))...)
	if err != nil {
		log.Fatal(err)
	}

	var addr string
	network := "mainnet"
	for _, item := range res.Items {
		if item.Name == "ton_addr" {
			addr = item.Address
			if item.Network == -3 {
				network = "testnet"
			}
		}
	}
	fmt.Printf(
		"%s %s for %s is connected to %s with %s address\n\n",
		res.Device.AppName,
		res.Device.AppVersion,
		res.Device.Platform,
		network,
		addr,
	)

	msg, err := tonconnect.NewMessage("0QBZ_35Wy144n2GBM93YpcV4KOKcIjDJk8DdX4kyXEEHcbLZ", "100000000")
	if err != nil {
		log.Fatal(err)
	}
	tx, err := tonconnect.NewTransaction(
		tonconnect.WithTimeout(10*time.Minute),
		tonconnect.WithTestnet(),
		tonconnect.WithMessage(*msg),
	)
	if err != nil {
		log.Fatal(err)
	}
	boc, err := s.SendTransaction(ctx, *tx)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Bag of Cells: %x", boc)
	}

	if err := s.Disconnect(ctx); err != nil {
		log.Fatal(err)
	}
}
