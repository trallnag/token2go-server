# OAuth2 Proxy + Okta

Simple setup using [OAuth2 Proxy](https://github.com/oauth2-proxy/oauth2-proxy)
and Okta to run, develop, and test Token2Go locally with a popular OAuth & OIDC
in front of it.

The working directory for this document is the repository root.

## Okta Web App

Ensure you have registered a web app in Okta. You can more or less follow the
steps described
[here](https://oauth2-proxy.github.io/oauth2-proxy/docs/configuration/oauth_provider#okta---localhost)
in the OAuth2 Proxy documentation.

## Run Token2go

Run Token2go. Either directly or with Air.

```shell
air
go run main.go
```

## Run OAuth2 Proxy

Set OAuth 2.0 and OIDC related environment variables. Replace the values with
your own web application interface from Okta.

```shell
export OAUTH2_PROXY_CLIENT_ID="0oa6gqeo7qX38IBob5d7"
export OAUTH2_PROXY_OIDC_ISSUER_URL="https://dev-06026349.okta.com"
```

Also don't forget the secret. Ensure this doesn't show up in the shell history.

```shell
export OAUTH2_PROXY_CLIENT_SECRET="xxx"
```

Place [oauth2-proxy.cfg](oauth2-proxy.cfg) at `tmp/oauth2-proxy.cfg`.

Run OAuth2 Proxy.

```shell
oauth2-proxy --code-challenge-method=S256 --config=docs/devel/oauth2-proxy-okta/oauth2-proxy.cfg
```

## Usage

Navigate in your browser to
[http://localhost:4180/echo](http://localhost:4180/echo). You should see a bunch
of entries including `X-Auth-Request-Access-Token` which represents the access
token.

## Cleanup

Cleanup if necessary.

```shell
rm tmp/oauth2-proxy.cfg
unset OAUTH2_PROXY_CLIENT_ID
unset OAUTH2_PROXY_OIDC_ISSUER_URL
unset OAUTH2_PROXY_CLIENT_SECRET
```
