export interface Identity {
  id: string;
  schema_id: string;
  schema_url: string;
  state: string;
  state_changed_at: string;
  traits?: {
    email?: string;
    username?: string;
  };
  metadata_public?: string;
  metadata_admin?: string;
  created_at?: string;
  updated_at?: string;
  organization_id?: string;
}
