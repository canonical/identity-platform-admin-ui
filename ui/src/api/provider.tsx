import { handleRequest } from "util/api";
import { IdentityProvider } from "types/provider";
import axios from "axios";
import { PaginatedResponse, ApiResponse } from "types/api";

export const fetchProviders = (): Promise<
  PaginatedResponse<IdentityProvider[]>
> =>
  handleRequest(() =>
    axios.get<PaginatedResponse<IdentityProvider[]>>(`/idps`),
  );

export const fetchProvider = (
  providerId: string,
): Promise<IdentityProvider> => {
  return new Promise((resolve, reject) => {
    handleRequest<IdentityProvider[], ApiResponse<IdentityProvider[]>>(() =>
      axios.get<ApiResponse<IdentityProvider[]>>(`/idps/${providerId}`),
    )
      .then(({ data }) => resolve(data[0]))
      .catch((error) => reject(error));
  });
};

export const createProvider = (body: string): Promise<unknown> =>
  handleRequest(() => axios.post("/idps", body));

export const updateProvider = (
  providerId: string,
  values: string,
): Promise<ApiResponse<IdentityProvider>> =>
  handleRequest(() =>
    axios.patch<ApiResponse<IdentityProvider>>(`/idps/${providerId}`, values),
  );

export const deleteProvider = (providerId: string): Promise<unknown> =>
  handleRequest(() => axios.delete(`/idps/${providerId}`));
