import { Client } from "types/client";
import { PaginatedResponse, ApiResponse } from "types/api";
import { handleRequest, PAGE_SIZE } from "util/api";
import { axiosInstance } from "./axios";

export const fetchClients = (
  pageToken: string,
): Promise<PaginatedResponse<Client[]>> =>
  handleRequest(() =>
    axiosInstance.get<PaginatedResponse<Client[]>>(
      `/clients?page_token=${pageToken}&size=${PAGE_SIZE}`,
    ),
  );

export const fetchClient = (clientId: string): Promise<ApiResponse<Client>> =>
  handleRequest(() =>
    axiosInstance.get<ApiResponse<Client>>(`/clients/${clientId}`),
  );

export const createClient = (values: string): Promise<ApiResponse<Client>> =>
  handleRequest(() =>
    axiosInstance.post<ApiResponse<Client>>("/clients", values),
  );

export const updateClient = (
  clientId: string,
  values: string,
): Promise<ApiResponse<Client>> =>
  handleRequest(() =>
    axiosInstance.put<ApiResponse<Client>>(`/clients/${clientId}`, values),
  );

export const deleteClient = (client: string) =>
  handleRequest(() =>
    axiosInstance.delete<ApiResponse<string>>(`/clients/${client}`),
  );
