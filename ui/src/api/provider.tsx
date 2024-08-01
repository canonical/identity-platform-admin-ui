import { handleRequest } from "util/api";
import { IdentityProvider } from "types/provider";
import axios from "axios";

export const fetchProviders = (): Promise<IdentityProvider[]> =>
  handleRequest(() => axios.get<IdentityProvider[]>(`/idps`));

export const fetchProvider = (
  providerId: string,
): Promise<IdentityProvider> => {
  return new Promise((resolve, reject) => {
    handleRequest(() => axios.get<IdentityProvider[]>(`/idps/${providerId}`))
      .then((data) => resolve(data[0]))
      .catch((error) => reject(error));
  });
};

export const createProvider = (body: string): Promise<void> =>
  handleRequest(() => axios.post("/idps", body));

export const updateProvider = (
  providerId: string,
  values: string,
): Promise<IdentityProvider> =>
  handleRequest(() =>
    axios.patch<IdentityProvider>(`/idps/${providerId}`, values),
  );

export const deleteProvider = (providerId: string): Promise<void> =>
  handleRequest(() => axios.delete(`/idps/${providerId}`));
