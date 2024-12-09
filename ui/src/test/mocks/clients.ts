import { faker } from "@faker-js/faker";

import { Client } from "types/client";

export const mockClient = (overrides?: Partial<Client>): Client => ({
  allowed_cors_origins: [],
  audience: [],
  client_id: faker.word.sample(),
  client_name: faker.word.sample(),
  client_secret: faker.word.sample(),
  client_secret_expires_at: faker.number.int(),
  client_uri: faker.word.sample(),
  contacts: [],
  created_at: faker.date.anytime().toISOString(),
  grant_types: [],
  logo_uri: faker.word.sample(),
  owner: faker.word.sample(),
  policy_uri: faker.word.sample(),
  redirect_uris: [],
  request_object_signing_alg: faker.word.sample(),
  response_types: [],
  scope: faker.word.sample(),
  subject_type: faker.word.sample(),
  token_endpoint_auth_method: faker.word.sample(),
  tos_uri: faker.word.sample(),
  updated_at: faker.date.anytime().toISOString(),
  userinfo_signed_response_alg: faker.word.sample(),
  ...overrides,
});
