import { ApiResponse } from "types/api";
import { handleResponse } from "util/api";
import { IdentityProvider } from "types/provider";

export const fetchProviders = (): Promise<IdentityProvider[]> => {
  return new Promise((resolve, reject) => {
    fetch("/api/v0/idps")
      .then(handleResponse)
      .then((result: ApiResponse<IdentityProvider[]>) => resolve(result.data))
      .catch(reject);
  });
};

export const fetchProvider = (
  providerId: string,
): Promise<IdentityProvider> => {
  return new Promise((resolve, reject) => {
    fetch(`/api/v0/idps/${providerId}`)
      .then(handleResponse)
      .then((result: ApiResponse<IdentityProvider[]>) =>
        resolve(result.data[0]),
      )
      .catch(reject);
  });
};

export const createProvider = (body: string): Promise<void> => {
  return new Promise((resolve, reject) => {
    fetch("/api/v0/idps", {
      method: "POST",
      body: body,
    })
      .then(handleResponse)
      .then(resolve)
      .catch(reject);
  });
};

export const updateProvider = (
  providerId: string,
  values: string,
): Promise<IdentityProvider> => {
  return new Promise((resolve, reject) => {
    fetch(`/api/v0/idps/${providerId}`, {
      method: "PATCH",
      body: values,
    })
      .then(handleResponse)
      .then((result: ApiResponse<IdentityProvider>) => resolve(result.data))
      .catch(reject);
  });
};

export const deleteProvider = (providerId: string) => {
  return new Promise((resolve, reject) => {
    fetch(`/api/v0/idps/${providerId}`, {
      method: "DELETE",
    })
      .then(resolve)
      .catch(reject);
  });
};
