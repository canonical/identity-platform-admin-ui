export interface Client {
  allowed_cors_origins: string[];
  audience: string[];
  authorization_code_grant_access_token_lifespan?: string;
  authorization_code_grant_id_token_lifespan?: string;
  authorization_code_grant_refresh_token_lifespan?: string;
  client_credentials_grant_access_token_lifespan?: string;
  client_id: string;
  client_name: string;
  client_secret: string;
  client_secret_expires_at: number;
  client_uri: string;
  contacts: string[];
  created_at: string;
  grant_types: string[];
  implicit_grant_access_token_lifespan?: string;
  implicit_grant_id_token_lifespan?: string;
  jwt_bearer_grant_access_token_lifespan?: string;
  logo_uri: string;
  owner: string;
  policy_uri: string;
  redirect_uris: string[];
  refresh_token_grant_access_token_lifespan?: string;
  refresh_token_grant_id_token_lifespan?: string;
  refresh_token_grant_refresh_token_lifespan?: string;
  request_object_signing_alg: string;
  response_types: string[];
  scope: string;
  subject_type: string;
  token_endpoint_auth_method: string;
  tos_uri: string;
  updated_at: string;
  userinfo_signed_response_alg: string;
}
