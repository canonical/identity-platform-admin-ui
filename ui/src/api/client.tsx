import { Client } from "types/client";
import { PaginatedResponse } from "types/api";
import { handleRequest, PAGE_SIZE } from "util/api";
import axios from "axios";

export const fetchClients = (
  pageToken: string,
): Promise<PaginatedResponse<Client[]>> =>
  handleRequest(() =>
    axios.get<PaginatedResponse<Client[]>>(
      `/clients?page_token=${pageToken}&size=${PAGE_SIZE}`,
    ),
  );

export const fetchClient = (clientId: string): Promise<Client> =>
  handleRequest(() => axios.get<Client>(`/clients/${clientId}`));

export const createClient = (values: string): Promise<Client> =>
  handleRequest(() => axios.post<Client>("/clients", values));

export const updateClient = (
  clientId: string,
  values: string,
): Promise<Client> =>
  handleRequest(() => axios.post<Client>(`/clients/${clientId}`, values));

export const deleteClient = (client: string) =>
  handleRequest(() =>
    axios.get<string>(`/clients/${client}`, {
      method: "DELETE",
    }),
  );
