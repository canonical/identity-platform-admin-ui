import { ErrorResponse } from "types/api";

export const PAGE_SIZE = 50;

export const handleResponse = async (response: Response) => {
  if (!response.ok) {
    const result = (await response.json()) as ErrorResponse;
    throw Error(result.error ?? result.message);
  }
  return response.json();
};
