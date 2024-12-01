import { ApiResponse, ErrorResponse } from "types/api";
import { AxiosResponse } from "axios";
import { AxiosError } from "axios";

export const PAGE_SIZE = 50;

export const handleRequest = <T, R extends ApiResponse<T>>(
  request: () => Promise<AxiosResponse<R>>,
): Promise<R> => {
  return new Promise((resolve, reject) => {
    request()
      .then((response) => resolve(response.data))
      .catch(({ response }: AxiosError<ErrorResponse>) =>
        reject(response?.data?.error ?? response?.data?.message),
      );
  });
};
