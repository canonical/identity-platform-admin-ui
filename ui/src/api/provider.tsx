import { handleRequest } from "util/api";
import { IdentityProvider } from "types/provider";
import { PaginatedResponse, ApiResponse } from "types/api";
import { axiosInstance } from "./axios";

export const fetchProviders = (): Promise<
  PaginatedResponse<IdentityProvider[]>
> =>
  handleRequest(() =>
    axiosInstance.get<PaginatedResponse<IdentityProvider[]>>(`/idps`),
  );

export const fetchProvider = (
  providerId: string,
): Promise<IdentityProvider> => {
  return new Promise((resolve, reject) => {
    handleRequest<IdentityProvider[], ApiResponse<IdentityProvider[]>>(() =>
      axiosInstance.get<ApiResponse<IdentityProvider[]>>(`/idps/${providerId}`),
    )
      .then(({ data }) => resolve(data[0]))
      .catch((error) => reject(error));
  });
};

export const createProvider = (body: string): Promise<unknown> =>
  handleRequest(() => axiosInstance.post("/idps", body));

export const updateProvider = (
  providerId: string,
  values: string,
): Promise<ApiResponse<IdentityProvider>> =>
  handleRequest(() =>
    axiosInstance.patch<ApiResponse<IdentityProvider>>(
      `/idps/${providerId}`,
      values,
    ),
  );

export const deleteProvider = (providerId: string): Promise<unknown> =>
  handleRequest(() => axiosInstance.delete(`/idps/${providerId}`));
