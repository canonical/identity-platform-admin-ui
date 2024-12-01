import { faker } from "@faker-js/faker";

import { ApiResponse, PaginatedResponse, ErrorResponse } from "types/api";

export const mockApiResponse = <T>(
  data: T,
  overrides?: Partial<ApiResponse<T>>,
): ApiResponse<T> => ({
  data,
  message: faker.word.sample(),
  status: faker.number.int(),
  ...overrides,
});

export const mockPaginatedResponseMeta = (
  overrides?: Partial<PaginatedResponse<unknown>["_meta"]>,
): PaginatedResponse<unknown>["_meta"] => ({
  ...overrides,
});

export const mockPaginatedResponse = <T>(
  data: T,
  overrides?: Partial<PaginatedResponse<T>>,
): PaginatedResponse<T> => ({
  _meta: mockPaginatedResponseMeta(),
  ...mockApiResponse<T>(data),
  ...overrides,
});

export const mockErrorResponse = (
  overrides?: Partial<ErrorResponse>,
): ErrorResponse => ({
  ...overrides,
});
