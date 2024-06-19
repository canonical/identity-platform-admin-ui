# Admin Service OAuth2 integration with Hydra

### Environment variables needed

```shell
AUTHENTICATION_ENABLED=true (defaults to false)
OIDC_ISSUER=http://localhost:4444
OAUTH2_CLIENT_ID="hydra client id for the Admin UI backend"
OAUTH2_CLIENT_SECRET=client secret
OAUTH2_REDIRECT_URI=http://localhost:${PORT}/api/v0/auth/callback
OAUTH2_CODEGRANT_SCOPES=openid,offline_access (defaults to these two, technically two more would be needed: "profile,email")
OAUTH2_AUTH_COOKIES_ENCRYPTION_KEY="WrfOcYmVBwyduEbKYTUhO4X7XVaOQ1wF" (required, this needs to be exactly 32 bytes in lenght secret key)
OAUTH2_REFRESHTOKEN_COOKIE_TTL_SECONDS=expiration in seconds for the refresh token cookie, should match the hydra refresh token ttl, (defaults to 21600)
ACCESS_TOKEN_VERIFICATION_STRATEGY="jwks|userinfo", (defaults to jwks)
```

## Hydra config

Here's the Hydra config:

```yaml
serve:
  cookies:
    same_site_mode: Lax
  admin:
    cors:
      enabled: true
      allowed_origins:
        - "*"
  public:
    cors:
      enabled: true
      allowed_origins:
        - "*"

log:
  leak_sensitive_values: true
  level: debug

strategies:
  access_token: jwt

ttl:
  access_token: 10m
  id_token: 10m
  refresh_token: 12h

oauth2:
  expose_internal_errors: true
  device_authorization:
    # configure how often a non-interactive device should poll the device token endpoint, default 5s
    token_polling_interval: 5s

urls:
  self:
    issuer: http://localhost:4444
    public: http://localhost:4444
  consent: http://localhost:4455/ui/consent
  login: http://localhost:4455/ui/login
  error: http://localhost:4455/ui/oidc_error
  device_verification: http://localhost:4455/ui/device_code
  post_device_done: http://localhost:4455/ui/device_complete

secrets:
  system:
    - youReallyNeedToChangeThis

```

It is important that hydra has the same host for ISSUER and PUBLIC since
it will affect access token and ID token verification.

## Hydra OAuth2 client assumptions

The Hydra client for Admin UI needs to be set with the proper:

- audience: the client ID value
- redirect URIs: `"http://localhost:$PORT/api/v0/login"`
- scopes: `"openid,offline_access,profile,email"`
- response types: `"token,code,id_token"`
- grant types: `"authorization_code,refresh_token"`

If you use the Login
UI [docker compose](https://github.com/canonical/identity-platform-login-ui/blob/main/docker-compose.yml) to deploy the
solution with Github integration, you can first create an OAuth2
client.
Then update it with the correct values, as in the following script:

```shell
# Update client id for allowing audience in jwt access token
docker compose exec -it \
    hydra hydra update client $CLIENT_ID \
    --skip-tls-verify --name client-name --secret client-secret \
    --skip-consent --grant-type authorization_code,refresh_token \
    --response-type token,code,id_token  \
    --scope openid,offline_access,profile,email \
    --redirect-uri http://localhost:8888/api/v0/login \
    --endpoint http://localhost:4445 \
    --audience $CLIENT_ID
```
