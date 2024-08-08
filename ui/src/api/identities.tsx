import { ApiResponse, PaginatedResponse } from "types/api";
import { handleRequest, PAGE_SIZE } from "util/api";
import { Identity } from "types/identity";
import { axiosInstance } from "./axios";

export const fetchIdentities = (
  pageToken: string,
): Promise<PaginatedResponse<Identity[]>> =>
  handleRequest(() =>
    axiosInstance.get<PaginatedResponse<Identity[]>>(
      `/identities?page_token=${pageToken}&size=${PAGE_SIZE}`,
    ),
  );

export const fetchIdentity = (
  identityId: string,
): Promise<ApiResponse<Identity>> =>
  handleRequest(() =>
    axiosInstance.get<ApiResponse<Identity>>(`/identities/${identityId}`),
  );

export const createIdentity = (body: string): Promise<unknown> =>
  handleRequest(() => axiosInstance.post("/identities", body));

export const updateIdentity = (
  identityId: string,
  values: string,
): Promise<ApiResponse<Identity>> =>
  handleRequest(() =>
    axiosInstance.patch<ApiResponse<Identity>>(
      `/identities/${identityId}`,
      values,
    ),
  );

export const deleteIdentity = (identityId: string): Promise<unknown> =>
  handleRequest(() => axiosInstance.delete(`/identities/${identityId}`));
