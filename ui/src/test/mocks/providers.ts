import { IdentityProvider } from "types/provider";

export const mockIdentityProvider = (
  overrides?: Partial<IdentityProvider>,
): IdentityProvider => ({
  ...overrides,
});
