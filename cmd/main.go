package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cameo-engineering/tonconnect"
	"golang.org/x/exp/maps"
)

func main() {
	s, _ := tonconnect.NewSession()

	connreq, _ := tonconnect.NewConnectRequest(
		"https://raw.githubusercontent.com/Ton-Split/tonconnect-manifest/main/tonconnect-manifest.json",
		tonconnect.WithProofRequest("test"),
	)
	unilink, _ := s.GenerateUniversalLink(tonconnect.Wallets["tonkeeper"], *connreq)
	fmt.Println(unilink)

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	_, err := s.Connect(ctx, maps.Values(tonconnect.Wallets)...)
	if err != nil {
		log.Fatal(err)
	}

	msg, _ := tonconnect.NewMessage("UQBE5SVSmrhuL7w6gbPxkxk0RcX8v4wA8ItWaIJZTHjrYXwk", "1000000")
	tx, _ := tonconnect.NewTransaction(tonconnect.WithTimeout(1*time.Minute), tonconnect.WithTestnet(), tonconnect.WithMessage(*msg))

	time.Sleep(5 * time.Second)

	if boc, err := s.SendTransaction(ctx, *tx); err != nil {
		log.Fatal(err)
	} else {
		fmt.Printf("%x\n", boc)
	}

	time.Sleep(10 * time.Second)

	if err := s.Disconnect(ctx); err != nil {
		log.Fatal(err)
	}
}
