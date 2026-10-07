package ledger

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"time"

	protocol "github.com/lifeboat008/lifeboat-protocol"
	"github.com/stellar/go-stellar-sdk/clients/horizonclient"
	"github.com/stellar/go-stellar-sdk/keypair"
	"github.com/stellar/go-stellar-sdk/network"
	horizon "github.com/stellar/go-stellar-sdk/protocols/horizon"
	"github.com/stellar/go-stellar-sdk/txnbuild"
)

type horizonNetwork interface {
	AccountDetail(horizonclient.AccountRequest) (horizon.Account, error)
	SubmitTransactionXDR(string) (horizon.Transaction, error)
	TransactionDetail(string) (horizon.Transaction, error)
}

type TestnetGateway struct {
	signer *keypair.Full
	client horizonNetwork
}

func NewTestnetGateway(secret string, client horizonNetwork) (*TestnetGateway, error) {
	signer, err := keypair.ParseFull(secret)
	if err != nil {
		return nil, errors.New("invalid testnet signing key")
	}
	if client == nil {
		client = &horizonclient.Client{
			HorizonURL: "https://horizon-testnet.stellar.org/",
			HTTP:       &http.Client{Timeout: 20 * time.Second},
			AppName:    "lifeboat",
		}
	}
	return &TestnetGateway{signer: signer, client: client}, nil
}

func (g *TestnetGateway) Prepare(ctx context.Context, claim protocol.Claim) (Prepared, error) {
	if err := ctx.Err(); err != nil {
		return Prepared{}, err
	}
	if claim.State != protocol.ClaimApproved || claim.AmountStroops <= 0 {
		return Prepared{}, errors.New("claim is not approved")
	}
	account, err := g.client.AccountDetail(horizonclient.AccountRequest{AccountID: g.signer.Address()})
	if err != nil {
		return Prepared{}, errors.New("cannot load testnet payment account")
	}
	amount := formatStroops(claim.AmountStroops)
	op := &txnbuild.Payment{Destination: claim.DestinationAccount, Amount: amount, Asset: txnbuild.NativeAsset{}}
	if err := op.Validate(); err != nil {
		return Prepared{}, errors.New("invalid destination or payment amount")
	}
	identifier := sha256.Sum256([]byte(claim.ID))
	memo := "lb:" + hex.EncodeToString(identifier[:12])
	transaction, err := txnbuild.NewTransaction(txnbuild.TransactionParams{
		SourceAccount: &account, IncrementSequenceNum: true,
		Operations: []txnbuild.Operation{op}, BaseFee: txnbuild.MinBaseFee,
		Memo:          txnbuild.MemoText(memo),
		Preconditions: txnbuild.Preconditions{TimeBounds: txnbuild.NewTimeout(300)},
	})
	if err != nil {
		return Prepared{}, errors.New("cannot build testnet payment")
	}
	signed, err := transaction.Sign(network.TestNetworkPassphrase, g.signer)
	if err != nil {
		return Prepared{}, errors.New("cannot sign testnet payment")
	}
	hash, err := signed.HashHex(network.TestNetworkPassphrase)
	if err != nil {
		return Prepared{}, errors.New("cannot hash testnet payment")
	}
	xdr, err := signed.Base64()
	if err != nil {
		return Prepared{}, errors.New("cannot encode testnet payment")
	}
	return Prepared{ClaimID: claim.ID, TransactionHash: hash, EnvelopeXDR: xdr, AmountStroops: claim.AmountStroops, Destination: claim.DestinationAccount, Network: "testnet"}, nil
}

func (g *TestnetGateway) Submit(ctx context.Context, prepared Prepared) (Receipt, error) {
	if err := ctx.Err(); err != nil {
		return Receipt{}, err
	}
	result, err := g.client.SubmitTransactionXDR(prepared.EnvelopeXDR)
	if err != nil {
		return Receipt{}, errors.New("testnet submission uncertain; reconcile the stored transaction hash")
	}
	if result.Hash != prepared.TransactionHash {
		return Receipt{}, fmt.Errorf("testnet returned a different transaction hash")
	}
	return Receipt{TransactionHash: result.Hash, Network: "testnet", Confirmed: result.Successful}, nil
}

func (g *TestnetGateway) Lookup(ctx context.Context, hash string) (Receipt, error) {
	if err := ctx.Err(); err != nil {
		return Receipt{}, err
	}
	result, err := g.client.TransactionDetail(hash)
	if err != nil {
		if horizonErr := horizonclient.GetError(err); horizonErr != nil && horizonErr.Response != nil && horizonErr.Response.StatusCode == http.StatusNotFound {
			return Receipt{}, ErrNotFound
		}
		return Receipt{}, errors.New("cannot reconcile testnet transaction")
	}
	return Receipt{TransactionHash: result.Hash, Network: "testnet", Confirmed: result.Successful}, nil
}

func formatStroops(stroops int64) string {
	return fmt.Sprintf("%d.%07d", stroops/10_000_000, stroops%10_000_000)
}
