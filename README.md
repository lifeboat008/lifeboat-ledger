<p align="center"><img src="assets/logo.svg" alt="Lifeboat logo" width="112"></p>

# lifeboat-ledger

[![Go CI](https://github.com/lifeboat008/lifeboat-ledger/actions/workflows/ci.yml/badge.svg)](https://github.com/lifeboat008/lifeboat-ledger/actions/workflows/ci.yml)

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

Run `go test ./...` and `go vet ./...` with Go 1.26.3 or newer. The tagged `lifeboat-protocol` dependency is public; no module token is needed. Do not use this testnet pilot for real funds.

The [product requirements](https://github.com/lifeboat008/lifeboat-api/blob/main/product/docs/PRD.md), [architecture](https://github.com/lifeboat008/lifeboat-api/blob/main/product/docs/ARCHITECTURE.md), and [Wave plan](https://github.com/lifeboat008/lifeboat-api/blob/main/product/docs/WAVE.md) are versioned in `lifeboat-api`.

## Contributing and security

See [CONTRIBUTING.md](CONTRIBUTING.md) for pull requests, [SECURITY.md](SECURITY.md) for private vulnerability reports, and [LICENSE](LICENSE) for MIT terms.

Maintainers: [lifeboat008](https://github.com/lifeboat008). Discuss public work in [issues](https://github.com/lifeboat008/lifeboat-ledger/issues); report vulnerabilities privately as described in SECURITY.md. This pilot has not had a formal security audit.
