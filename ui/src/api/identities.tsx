import { PaginatedResponse } from "types/api";
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

export const fetchIdentity = (identityId: string): Promise<Identity> =>
  handleRequest(() => axios.get<Identity>(`/identities/${identityId}`));

export const createIdentity = (body: string): Promise<void> =>
  handleRequest(() => axios.post("/identities", body));

export const updateIdentity = (
  identityId: string,
  values: string,
): Promise<Identity> =>
  handleRequest(() =>
    axios.patch<Identity>(`/identities/${identityId}`, values),
  );

export const deleteIdentity = (identityId: string): Promise<void> =>
  handleRequest(() => axios.delete(`/identities/${identityId}`));
