export interface IdentityProvider {
  apple_private_key?: string;
  apple_private_key_id?: string;
  apple_team_id?: string;
  auth_url?: string;
  client_id?: string;
  client_secret?: string;
  id?: string;
  issuer_url?: string;
  label?: string;
  mapper_url?: string;
  microsoft_tenant?: string;
  provider?: string;
  requested_claims?: string;
  scope?: string[];
  subject_source?: string;
  token_url?: string;
}
