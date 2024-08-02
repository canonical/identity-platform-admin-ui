import { ApiResponse, PaginatedResponse } from "types/api";
import { handleRequest, PAGE_SIZE } from "util/api";
import { Identity } from "types/identity";
import axios from "axios";

export const fetchIdentities = (
  pageToken: string,
): Promise<PaginatedResponse<Identity[]>> =>
  handleRequest(() =>
    axios.get<PaginatedResponse<Identity[]>>(
      `/identities?page_token=${pageToken}&size=${PAGE_SIZE}`,
    ),
  );

export const fetchIdentity = (
  identityId: string,
): Promise<ApiResponse<Identity>> =>
  handleRequest(() =>
    axios.get<ApiResponse<Identity>>(`/identities/${identityId}`),
  );

export const createIdentity = (body: string): Promise<unknown> =>
  handleRequest(() => axios.post("/identities", body));

export const updateIdentity = (
  identityId: string,
  values: string,
): Promise<ApiResponse<Identity>> =>
  handleRequest(() =>
    axios.patch<ApiResponse<Identity>>(`/identities/${identityId}`, values),
  );

export const deleteIdentity = (identityId: string): Promise<unknown> =>
  handleRequest(() => axios.delete(`/identities/${identityId}`));
