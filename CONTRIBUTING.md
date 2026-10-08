# Contributing to lifeboat-ledger

Lifeboat coordinates reviewed rescue work and testnet payouts for consenting repositories. Start with an issue, keep one change per pull request, and describe the behavior and tests.

- Use Go for product code. Run gofmt, go vet ./..., and go test ./... before requesting review.
- Keep this repository within its responsibility described in README.md; changes across repositories should be coordinated through tagged module releases.
- Use synthetic GitHub events and Stellar testnet data only. Never commit signing keys, access tokens, real sponsor or maintainer information, or production transaction payloads.
- Preserve human approval before payouts. Do not infer ownership or payment authority from GitHub activity alone.
- Add meaningful tests for changed behavior and note user-visible changes in CHANGELOG.md.

For security issues, follow SECURITY.md instead of posting exploit details in a public issue.