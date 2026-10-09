<p align="center"><img src="assets/logo.svg" alt="Lifeboat logo" width="112"></p>

<h1 align="center">lifeboat-ledger</h1>

<p align="center">The Stellar payment boundary for Lifeboat, written in Go.</p>

<p align="center">
  <a href="https://github.com/lifeboat008/lifeboat-ledger/actions/workflows/ci.yml"><img src="https://github.com/lifeboat008/lifeboat-ledger/actions/workflows/ci.yml/badge.svg" alt="Go CI"></a>
  <img src="https://img.shields.io/badge/license-MIT-blue" alt="MIT license">
  <img src="https://img.shields.io/badge/go-1.26.3%2B-00ADD8?logo=go&logoColor=white" alt="Go 1.26.3+">
  <img src="https://img.shields.io/badge/Stellar-testnet%20only-7D00FF" alt="Stellar testnet only">
  <img src="https://img.shields.io/badge/payments-idempotent-success" alt="Idempotent payments">
  <img src="https://img.shields.io/badge/status-pilot-orange" alt="Pilot status">
  <a href="https://cjay-1.gitbook.io/lifeboat-docs/"><img src="https://img.shields.io/badge/docs-GitBook-3884FF" alt="Documentation"></a>
</p>

## Contents

- [What is Lifeboat](#what-is-lifeboat)
- [What this repository does](#what-this-repository-does)
- [Testnet payment flow](#testnet-payment-flow)
- [Quick start](#quick-start)
- [Use it in Go](#use-it-in-go)
- [Payment details](#payment-details)
- [Rules for callers](#rules-for-callers)
- [How the four repositories fit together](#how-the-four-repositories-fit-together)
- [Project status](#project-status)
- [Open work](#open-work)
- [Documentation](#documentation)
- [Contributing and security](#contributing-and-security)
- [Maintainers](#maintainers)
- [Contributors](#contributors)
- [License](#license)

## What is Lifeboat

Lifeboat helps companies keep the open-source projects they depend on healthy. A sponsor commits a budget to a defined maintenance plan. A maintainer submits evidence of finished work as a claim. A named human reviewer approves or rejects it. Only an approved claim is paid, through a Stellar payment whose transaction hash anyone can verify.

GitHub activity is evidence, never approval. Lifeboat never pays because a project looks inactive.

## What this repository does

`lifeboat-ledger` turns an **approved** claim into one Stellar testnet payment and a verifiable receipt, even across retries.

| Owns | Does not own |
| --- | --- |
| Payment request validation against approved protocol state | Claim review or sponsor policy |
| Transaction construction, signing boundary, submission, reconciliation | Budgets and HTTP endpoints |
| Transaction lookup and receipts | GitHub evidence collection |

It does not store the signing key, hold budgets, or decide whether a claim deserves payment. `lifeboat-api` handles those duties and records the receipt.

## Testnet payment flow

```text
Prepare ──► persist hash + envelope ──► Lookup by hash ──► (not found?) Submit same envelope ──► Receipt
```

1. `Service.Prepare` accepts only an approved claim and asks `TestnetGateway` to build and sign a native XLM payment.
2. **Persist** the returned hash and signed envelope before going further.
3. On every attempt call `Service.Lookup` with the stored hash first.
4. Only if the hash is unknown, call `Service.Submit` with the **same** signed envelope.
5. Never prepare a second transaction because a submission returned an error.

## Quick start

With Go 1.26.3 or newer:

```bash
git clone https://github.com/lifeboat008/lifeboat-ledger.git
cd lifeboat-ledger
go test ./...
go vet ./...
```

The tagged `lifeboat-protocol` dependency is public, so no module token is needed. These tests use a fake gateway and need no network.

An opt-in test sends a real testnet payment through temporary Friendbot-funded accounts:

```bash
LIFEBOAT_LIVE_TESTNET=1 go test ./... -v
```

## Use it in Go

```go
gateway, err := ledger.NewTestnetGateway(signingKey, nil) // nil uses testnet Horizon
if err != nil { /* invalid key */ }
svc, _ := ledger.NewService(gateway)

prepared, err := svc.Prepare(ctx, claim) // claim.State must be "approved"
// persist prepared.TransactionHash and prepared.EnvelopeXDR here

receipt, err := svc.Lookup(ctx, prepared.TransactionHash)
if errors.Is(err, ledger.ErrNotFound) {
    receipt, err = svc.Submit(ctx, prepared)
}
```

Supply the signing key at runtime from a secret manager or environment variable. Never commit it, log it, or put it in a fixture.

`Gateway` is an interface, so you can fake the network in your own tests.

## Payment details

| Field | Value |
| --- | --- |
| Network | Stellar testnet (Horizon) |
| Asset | Native XLM |
| Operation | One `Payment` to the claim's destination account |
| Amount | The claim amount in stroops, formatted to 7 decimals |
| Memo | `lb:` plus 24 hex characters derived from the claim ID |
| Validity | 300-second time bound |

## Rules for callers

- Persist before you submit. A stored hash is how you recover after a crash.
- Look up before you resubmit.
- A `Receipt` with `Confirmed: false` is not a payment.
- Do not use this testnet pilot for real funds.

## How the four repositories fit together

```text
lifeboat-protocol ──► lifeboat-ledger ──► lifeboat-api ◄── lifeboat-github
```

| Repository | Responsibility |
| --- | --- |
| [lifeboat-protocol](https://github.com/lifeboat008/lifeboat-protocol) | Domain types, validation, budget arithmetic |
| [lifeboat-ledger](https://github.com/lifeboat008/lifeboat-ledger) | Stellar testnet payment construction, submission, reconciliation (this repository) |
| [lifeboat-api](https://github.com/lifeboat008/lifeboat-api) | HTTP API, persistence, authorization, audit history |
| [lifeboat-github](https://github.com/lifeboat008/lifeboat-github) | GitHub webhook verification and evidence submission |

## Project status

Lifeboat is a **Stellar testnet pilot** with no formal security audit. A confirmed testnet payment from this library is recorded in the [verification page](https://cjay-1.gitbook.io/lifeboat-docs/project-status/verification). Mainnet needs custody, asset, jurisdiction, and legal decisions first.

Known limitation, inferred from the code and not yet tested: a prepared payment expires after 300 seconds, and there is no re-prepare path yet. See the open work below.

## Open work

- [Background reconciliation worker for uncertain payments](https://github.com/lifeboat008/lifeboat-ledger/issues/1) (complexity: high)
- [Recover a prepared payment whose time bound has expired](https://github.com/lifeboat008/lifeboat-api/issues/4) (in `lifeboat-api`)

The product requirements, architecture, and Wave plan are versioned in `lifeboat-api` ([PRD](https://github.com/lifeboat008/lifeboat-api/blob/main/product/docs/PRD.md), [architecture](https://github.com/lifeboat008/lifeboat-api/blob/main/product/docs/ARCHITECTURE.md), [Wave plan](https://github.com/lifeboat008/lifeboat-api/blob/main/product/docs/WAVE.md)).

## Documentation

Full documentation, including [Payments on Stellar](https://cjay-1.gitbook.io/lifeboat-docs/how-it-works/payments-on-stellar), the API reference, role guides, security notes, and operations runbooks, is at **[cjay-1.gitbook.io/lifeboat-docs](https://cjay-1.gitbook.io/lifeboat-docs/)**.

## Contributing and security

- Read [CONTRIBUTING.md](CONTRIBUTING.md) before opening a pull request. Run `gofmt`, `go vet ./...`, and `go test ./...`, and add tests for changed behavior.
- Use testnet and synthetic data only. Never commit signing keys or tokens.
- Report vulnerabilities privately as described in [SECURITY.md](SECURITY.md). Do not post exploit details in a public issue.

## Maintainers

| Maintainer | Contact |
| --- | --- |
| [lifeboat008](https://github.com/lifeboat008) | [Open an issue](https://github.com/lifeboat008/lifeboat-ledger/issues) for public work; use SECURITY.md for private reports |

## Contributors

<a href="https://github.com/lifeboat008/lifeboat-ledger/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=lifeboat008/lifeboat-ledger" alt="Contributors">
</a>

## License

[MIT](LICENSE). This pilot has not had a formal security audit.
