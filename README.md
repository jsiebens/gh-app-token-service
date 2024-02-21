# gh-app-token-service

This lightweight service exchanges OIDC token from the trusted GitHub Actions token issuer for a short-lived GitHub App installation access token.

## Configuration

Configuration is done by settings environment variables:

- APP_ID: the App's ID
- APP_PRIVATE_KEY: the App's private key (file location or complete content)

## Deployment

When using in GitHub Actions, this service should be publicly available. E.g. on Google Cloud Run of [fly.io](https://fly.io).

## Alternatives

- [actions/create-github-app-token](https://github.com/actions/create-github-app-token)