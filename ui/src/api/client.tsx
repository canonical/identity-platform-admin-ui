import { Client } from "types/client";
import { ApiResponse, PaginatedResponse } from "types/api";
import { handleResponse, PAGE_SIZE } from "util/api";

export const fetchClients = (
  pageToken: string,
): Promise<PaginatedResponse<Client[]>> => {
  return new Promise((resolve, reject) => {
    fetch(
      `${import.meta.env.VITE_API_URL}/clients?page_token=${pageToken}&size=${PAGE_SIZE}`,
    )
      .then(handleResponse)
      .then((result: PaginatedResponse<Client[]>) => resolve(result))
      .catch(reject);
  });
};

export const fetchClient = (clientId: string): Promise<Client> => {
  return new Promise((resolve, reject) => {
    fetch(`${import.meta.env.VITE_API_URL}/clients/${clientId}`)
      .then(handleResponse)
      .then((result: ApiResponse<Client>) => resolve(result.data))
      .catch(reject);
  });
};

export const createClient = (values: string): Promise<Client> => {
  return new Promise((resolve, reject) => {
    fetch(`${import.meta.env.VITE_API_URL}/clients`, {
      method: "POST",
      body: values,
    })
      .then(handleResponse)
      .then((result: ApiResponse<Client>) => resolve(result.data))
      .catch(reject);
  });
};

export const updateClient = (
  clientId: string,
  values: string,
): Promise<Client> => {
  return new Promise((resolve, reject) => {
    fetch(`${import.meta.env.VITE_API_URL}/clients/${clientId}`, {
      method: "PUT",
      body: values,
    })
      .then(handleResponse)
      .then((result: ApiResponse<Client>) => resolve(result.data))
      .catch(reject);
  });
};

export const deleteClient = (client: string) => {
  return new Promise((resolve, reject) => {
    fetch(`${import.meta.env.VITE_API_URL}/clients/${client}`, {
      method: "DELETE",
    })
      .then(resolve)
      .catch(reject);
  });
};
