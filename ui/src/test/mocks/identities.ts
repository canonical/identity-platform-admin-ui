import { faker } from "@faker-js/faker";

import { Identity } from "types/identity";

export const mockIdentity = (overrides?: Partial<Identity>): Identity => ({
  id: faker.word.sample(),
  schema_id: faker.word.sample(),
  schema_url: faker.word.sample(),
  state: faker.word.sample(),
  state_changed_at: faker.word.sample(),
  ...overrides,
});
