package main

import (
	"context"
	"fmt"
	"time"

	"github.com/cameo-engineering/tonconnect"
)

func main() {
	s, _ := tonconnect.NewSession()

	connreq, _ := tonconnect.NewConnectRequest(
		"https://getgems.io/tcm.json",
		tonconnect.WithProofRequest("test"),
	)
	fmt.Println(s.GenerateDeeplink(*connreq))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	s.Connect(ctx)
}
