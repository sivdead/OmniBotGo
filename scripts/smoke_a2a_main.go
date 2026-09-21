//go:build ignore

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/sivdead/OmniBotGo/internal/a2a"
)

func main() {
	base := flag.String("base", "http://127.0.0.1:10000", "A2A base URL")
	text := flag.String("text", "How much is 10 USD to EUR?", "user text")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	c := a2a.NewClient(*base)
	card, err := c.FetchAgentCard(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "agent card: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("agent=%s protocol=%s\n", card.Name, card.ProtocolVersion)

	reply, err := c.SendText(ctx, *text)
	if err != nil {
		fmt.Fprintf(os.Stderr, "send: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("reply:\n%s\n", reply)
}
