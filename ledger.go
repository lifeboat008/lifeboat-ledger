package ledger

import (
	"context"
	"errors"
	"fmt"

	protocol "github.com/lifeboat008/lifeboat-protocol"
)

var ErrNotFound = errors.New("transaction not found")

type Prepared struct {
	ClaimID         string `json:"claim_id"`
	TransactionHash string `json:"transaction_hash"`
	EnvelopeXDR     string `json:"envelope_xdr"`
	AmountStroops   int64  `json:"amount_stroops"`
	Destination     string `json:"destination"`
	Network         string `json:"network"`
}

type Receipt struct {
	TransactionHash string `json:"transaction_hash"`
	Network         string `json:"network"`
	Confirmed       bool   `json:"confirmed"`
}

type Gateway interface {
	Prepare(context.Context, protocol.Claim) (Prepared, error)
	Submit(context.Context, Prepared) (Receipt, error)
	Lookup(context.Context, string) (Receipt, error)
}

type Service struct {
	gateway Gateway
}

func NewService(gateway Gateway) (*Service, error) {
	if gateway == nil {
		return nil, errors.New("payment gateway is required")
	}
	return &Service{gateway: gateway}, nil
}

func (s *Service) Prepare(ctx context.Context, claim protocol.Claim) (Prepared, error) {
	if claim.State != protocol.ClaimApproved || claim.ID == "" || claim.AmountStroops <= 0 || claim.DestinationAccount == "" {
		return Prepared{}, errors.New("only a complete approved claim can be prepared")
	}
	prepared, err := s.gateway.Prepare(ctx, claim)
	if err != nil {
		return Prepared{}, err
	}
	if prepared.ClaimID != claim.ID || prepared.TransactionHash == "" || prepared.EnvelopeXDR == "" || prepared.Network != "testnet" {
		return Prepared{}, fmt.Errorf("gateway returned an invalid prepared payment")
	}
	return prepared, nil
}

func (s *Service) Submit(ctx context.Context, prepared Prepared) (Receipt, error) {
	if prepared.TransactionHash == "" || prepared.EnvelopeXDR == "" || prepared.Network != "testnet" {
		return Receipt{}, errors.New("invalid prepared testnet payment")
	}
	return s.gateway.Submit(ctx, prepared)
}

func (s *Service) Lookup(ctx context.Context, hash string) (Receipt, error) {
	if hash == "" {
		return Receipt{}, errors.New("transaction hash is required")
	}
	return s.gateway.Lookup(ctx, hash)
}
