package ledger

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	protocol "github.com/lifeboat008/lifeboat-protocol"
	"github.com/stellar/go-stellar-sdk/keypair"
)

// This opt-in test creates temporary testnet accounts. It never prints their seeds.
func TestLiveTestnetPayment(t *testing.T) {
	if os.Getenv("LIFEBOAT_LIVE_TESTNET") != "1" {
		t.Skip("set LIFEBOAT_LIVE_TESTNET=1 for the live testnet smoke test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	source, err := keypair.Random()
	if err != nil {
		t.Fatal(err)
	}
	destination, err := keypair.Random()
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Timeout: 20 * time.Second}
	for _, account := range []string{source.Address(), destination.Address()} {
		request, err := http.NewRequestWithContext(ctx, "GET", "https://friendbot.stellar.org/?addr="+url.QueryEscape(account), nil)
		if err != nil {
			t.Fatal(err)
		}
		response, err := client.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = io.Copy(io.Discard, response.Body)
		_ = response.Body.Close()
		if response.StatusCode != http.StatusOK {
			t.Fatalf("testnet faucet returned %d", response.StatusCode)
		}
	}
	gateway, err := NewTestnetGateway(source.Seed(), nil)
	if err != nil {
		t.Fatal(err)
	}
	service, err := NewService(gateway)
	if err != nil {
		t.Fatal(err)
	}
	claim := protocol.Claim{ID: fmt.Sprintf("smoke-%d", time.Now().UnixNano()), State: protocol.ClaimApproved, AmountStroops: 1, DestinationAccount: destination.Address()}
	var prepared Prepared
	for attempt := 0; attempt < 10; attempt++ {
		prepared, err = service.Prepare(ctx, claim)
		if err == nil {
			break
		}
		select {
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-time.After(time.Second):
		}
	}
	if err != nil {
		t.Fatal(err)
	}
	receipt, err := service.Submit(ctx, prepared)
	if err != nil {
		t.Fatal(err)
	}
	if !receipt.Confirmed || receipt.TransactionHash != prepared.TransactionHash {
		t.Fatal("testnet payment was not confirmed")
	}
	lookup, err := service.Lookup(ctx, prepared.TransactionHash)
	if err != nil || !lookup.Confirmed {
		t.Fatalf("confirmed payment could not be reconciled: %v", err)
	}
	t.Logf("confirmed Stellar testnet transaction: %s", prepared.TransactionHash)
}
