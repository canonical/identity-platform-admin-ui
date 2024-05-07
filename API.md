# API documentation

## Clients API (Hydra OAuth2 clients)

```text
GET /api/v0/clients
GET /api/v0/clients/{id}
POST /api/v0/clients --> [payload](https://www.ory.sh/docs/hydra/reference/api#tag/oAuth2/operation/createOAuth2Client)
PUT /api/v0/clients/{id} --> [payload](https://www.ory.sh/docs/hydra/reference/api#tag/oAuth2/operation/createOAuth2Client)
DELETE /api/v0/clients/{id}
```

## Identities API (Kratos identity)

```text
GET /api/v0/identities
GET /api/v0/identities/{id}
POST /api/v0/identities --> [payload](https://www.ory.sh/docs/kratos/reference/api#tag/identity/operation/createIdentity)
PUT /api/v0/identities/{id} --> [payload](https://www.ory.sh/docs/kratos/reference/api#tag/identity/operation/updateIdentity)
DELETE /api/v0/identities/{id}
```

## IDProviders API

```text
GET /api/v0/idps
GET /api/v0/idps/{id}
DELETE /api/v0/idps/{id}
POST /api/v0/idps --> [payload](https://github.com/canonical/identity-platform-admin-ui/blob/main/pkg/idp/third_party.go)
PATCH /api/v0/idps/{id} --> [payload](https://github.com/canonical/identity-platform-admin-ui/blob/main/pkg/idp/third_party.go)
```

example below

```json
{
  "id": "abcd",
  "provider": "github",
  "label": "gh",
  "client_id": "randomstring",
  "client_secret": "randomstring",
  "issuer_url": "https://canonical.com",
  "auth_url": "https://canonical.com/auth",
  "token_url": "https://canonical.com/token",
  "microsoft_tenant": "",
  "subject_source": "",
  "apple_team_id": "",
  "apple_private_key_id": "",
  "apple_private_key": "",
  "scope": "openid,session",
  "mapper_url": "https://canonical.com/mapper",
  "requested_claims": {
    "userinfo": {
      "given_name": {
        "essential": true
      },
      "nickname": null,
      "email": {
        "essential": true
      },
      "email_verified": {
        "essential": true
      },
      "picture": null,
      "http://example.info/claims/groups": null
    },
    "id_token": {
      "auth_time": {
        "essential": true
      },
      "acr": {
        "values": [
          "urn:mace:incommon:iap:silver"
        ]
      }
    }
  }
}
```

## Identity Schemas API (Kratos Identity Schemas)

```text
GET /api/v0/schemas
GET /api/v0/schemas/{id}
POST /api/v0/schemas - [payload](###IdentitySchemaContainer)
PATCH /api/v0/schemas/{id} - [payload](###IdentitySchemaContainer)
DELETE /api/v0/schemas/{id}
GET /api/v0/schemas/default
PUT /api/v0/schemas/default
```

### IdentitySchemaContainer

```json
{
  "id": "test",
  "schema": {
    "$id": "https://schemas.canonical.com/presets/kratos/test_v0.json",
    "$schema": "http://json-schema.org/draft-07/schema#",
    "title": "Admin Account",
    "type": "object",
    "properties": {
      "traits": {
        "type": "object",
        "properties": {
          "username": {
            "type": "string",
            "title": "Username",
            "ory.sh/kratos": {
              "credentials": {
                "password": {
                  "identifier": true
                }
              }
            }
          }
        }
      }
    },
    "additionalProperties": true
  }
}
```

## Rules API (Oathkeeper Access Rules)

```text
GET /api/v0/rules
GET /api/v0/rules/{id}
POST /api/v0/rules - [payload](https://www.ory.sh/docs/oathkeeper/reference/api#tag/api/operation/getRule)
PUT /api/v0/rules/{id} - [payload](https://www.ory.sh/docs/oathkeeper/reference/api#tag/api/operation/getRule)
DELETE /api/v0/rules/{id}
```
