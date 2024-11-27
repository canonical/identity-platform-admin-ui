import { faker } from "@faker-js/faker";

import { Schema } from "types/schema";

export const mockSchema = (overrides?: Partial<Schema>): Schema => ({
  id: faker.word.sample(),
  schema: {},
  ...overrides,
});
