package ledger

import (
	"context"
	"testing"

	protocol "github.com/lifeboat008/lifeboat-protocol"
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
