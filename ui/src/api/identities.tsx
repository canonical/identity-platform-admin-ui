import { ApiResponse, PaginatedResponse } from "types/api";
import { handleResponse, PAGE_SIZE } from "util/api";
import { Identity } from "types/identity";

export const fetchIdentities = (
  pageToken: string,
): Promise<PaginatedResponse<Identity[]>> => {
  return new Promise((resolve, reject) => {
    fetch(`/api/v0/identities?page_token=${pageToken}&size=${PAGE_SIZE}`)
      .then(handleResponse)
      .then((result: PaginatedResponse<Identity[]>) => resolve(result))
      .catch(reject);
  });
};

export const fetchIdentity = (identityId: string): Promise<Identity> => {
  return new Promise((resolve, reject) => {
    fetch(`/api/v0/identities/${identityId}`)
      .then(handleResponse)
      .then((result: ApiResponse<Identity[]>) => resolve(result.data[0]))
      .catch(reject);
  });
};

export const createIdentity = (body: string): Promise<void> => {
  return new Promise((resolve, reject) => {
    fetch("/api/v0/identities", {
      method: "POST",
      body: body,
    })
      .then(handleResponse)
      .then(resolve)
      .catch(reject);
  });
};

export const updateIdentity = (
  identityId: string,
  values: string,
): Promise<Identity> => {
  return new Promise((resolve, reject) => {
    fetch(`/api/v0/identities/${identityId}`, {
      method: "PATCH",
      body: values,
    })
      .then(handleResponse)
      .then((result: ApiResponse<Identity>) => resolve(result.data))
      .catch(reject);
  });
};

export const deleteIdentity = (identityId: string) => {
  return new Promise((resolve, reject) => {
    fetch(`/api/v0/identities/${identityId}`, {
      method: "DELETE",
    })
      .then(resolve)
      .catch(reject);
  });
};
