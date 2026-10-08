<p align="center"><img src="assets/logo.svg" alt="Lifeboat logo" width="112"></p>

# lifeboat-ledger

Stellar payment boundary for Lifeboat, written in Go.

## Owns

- Payment request validation against approved protocol state.
- Stellar transaction construction, signing boundary, submission, and reconciliation.
- Transaction lookup and verifiable transaction receipts.

## Does not own

Claim review, sponsor policy, HTTP endpoints, or GitHub evidence collection.

## Testnet payment flow

`Service.Prepare` accepts only an approved claim and asks `TestnetGateway` to build and sign a native XLM payment using Stellar testnet. The prepared value contains the signed envelope XDR and transaction hash. Persist that value before calling `Service.Submit`. On retry, call `Service.Lookup` by the stored hash first; if it is absent, submit the **same** signed envelope. Never prepare another transaction merely because submission returned an error.

`TestnetGateway` loads the source account sequence from testnet Horizon. Supply its signing key at runtime from a secret manager or environment variable. This module does not store the key, maintain budgets, or decide whether a claim deserves payment. The `lifeboat-api` repository handles those duties and records the receipt.

Run `go test ./...` with access to the private, tagged `lifeboat-protocol` dependency. Go 1.26 is required. Do not use this testnet pilot for real funds.

Product PRD and architecture live in the parent `lifeboat/docs` folder in the local workspace.

## Contributing and security

See [CONTRIBUTING.md](CONTRIBUTING.md) for pull requests, [SECURITY.md](SECURITY.md) for private vulnerability reports, and [LICENSE](LICENSE) for MIT terms.
