package ledger

import (
	"context"
	"testing"

	protocol "github.com/lifeboat008/lifeboat-protocol"
	"github.com/stellar/go-stellar-sdk/clients/horizonclient"
	"github.com/stellar/go-stellar-sdk/keypair"
	"github.com/stellar/go-stellar-sdk/network"
	horizon "github.com/stellar/go-stellar-sdk/protocols/horizon"
	"github.com/stellar/go-stellar-sdk/txnbuild"
)

type fakeGateway struct {
	prepares int
}

func (f *fakeGateway) Prepare(_ context.Context, claim protocol.Claim) (Prepared, error) {
	f.prepares++
	return Prepared{ClaimID: claim.ID, TransactionHash: "hash", EnvelopeXDR: "xdr", Network: "testnet"}, nil
}
func (f *fakeGateway) Submit(_ context.Context, prepared Prepared) (Receipt, error) {
	return Receipt{TransactionHash: prepared.TransactionHash, Network: "testnet", Confirmed: true}, nil
}
func (f *fakeGateway) Lookup(_ context.Context, hash string) (Receipt, error) {
	return Receipt{TransactionHash: hash, Network: "testnet", Confirmed: true}, nil
}

func TestRejectsUnapprovedClaim(t *testing.T) {
	fake := &fakeGateway{}
	service, err := NewService(fake)
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.Prepare(context.Background(), protocol.Claim{ID: "claim", State: protocol.ClaimSubmitted, AmountStroops: 10, DestinationAccount: "G..."})
	if err == nil || fake.prepares != 0 {
		t.Fatal("unapproved claim reached payment gateway")
	}
}

func TestStroopFormatting(t *testing.T) {
	if formatStroops(10_000_001) != "1.0000001" || formatStroops(1) != "0.0000001" {
		t.Fatal("incorrect exact amount formatting")
	}
}

type fakeHorizon struct{ account horizon.Account }

func (f *fakeHorizon) AccountDetail(_ horizonclient.AccountRequest) (horizon.Account, error) {
	return f.account, nil
}
func (f *fakeHorizon) SubmitTransactionXDR(_ string) (horizon.Transaction, error) {
	return horizon.Transaction{}, nil
}
func (f *fakeHorizon) TransactionDetail(_ string) (horizon.Transaction, error) {
	return horizon.Transaction{}, nil
}

func TestPrepareSignsExactTestnetPayment(t *testing.T) {
	source, err := keypair.Random()
	if err != nil {
		t.Fatal(err)
	}
	destination, err := keypair.Random()
	if err != nil {
		t.Fatal(err)
	}
	client := &fakeHorizon{account: horizon.Account{AccountID: source.Address(), Sequence: 7}}
	gateway, err := NewTestnetGateway(source.Seed(), client)
	if err != nil {
		t.Fatal(err)
	}
	claim := protocol.Claim{ID: "claim-123", State: protocol.ClaimApproved, AmountStroops: 10_000_001, DestinationAccount: destination.Address()}
	prepared, err := gateway.Prepare(context.Background(), claim)
	if err != nil {
		t.Fatal(err)
	}
	if prepared.Network != "testnet" || prepared.AmountStroops != claim.AmountStroops || prepared.Destination != destination.Address() {
		t.Fatalf("prepared payment mismatch: %+v", prepared)
	}
	parsed, err := txnbuild.TransactionFromXDR(prepared.EnvelopeXDR)
	if err != nil {
		t.Fatal(err)
	}
	hash, err := parsed.HashHex(network.TestNetworkPassphrase)
	if err != nil {
		t.Fatal(err)
	}
	if hash != prepared.TransactionHash {
		t.Fatal("testnet hash does not match signed envelope")
	}
	transaction, ok := parsed.Transaction()
	if !ok || len(transaction.Operations()) != 1 || len(transaction.Signatures()) != 1 {
		t.Fatal("expected one signed payment operation")
	}
	payment, ok := transaction.Operations()[0].(*txnbuild.Payment)
	if !ok || payment.Amount != "1.0000001" || payment.Destination != destination.Address() {
		t.Fatalf("incorrect payment operation: %+v", payment)
	}
}
