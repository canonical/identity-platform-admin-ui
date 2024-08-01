import { PaginatedResponse } from "types/api";
import { handleRequest, PAGE_SIZE } from "util/api";
import { Schema } from "types/schema";
import axios from "axios";

export const fetchSchemas = (
  pageToken: string,
): Promise<PaginatedResponse<Schema[]>> =>
  handleRequest(() =>
    axios.get<PaginatedResponse<Schema[]>>(
      `/schemas?page_token=${pageToken}&page_size=${PAGE_SIZE}`,
    ),
  );

export const fetchSchema = (schemaId: string): Promise<Schema> => {
  return new Promise((resolve, reject) => {
    handleRequest(() => axios.get<Schema[]>(`/schemas/${schemaId}`))
      .then((data) => resolve(data[0]))
      .catch((error) => reject(error));
  });
};

export const createSchema = (body: string): Promise<void> =>
  handleRequest(() => axios.post("/schemas", body));

export const updateSchema = (
  schemaId: string,
  values: string,
): Promise<Schema> =>
  handleRequest(() => axios.patch<Schema>(`/schemas/${schemaId}`, values));

export const deleteSchema = (schemaId: string): Promise<void> =>
  handleRequest(() => axios.delete(`/schemas/${schemaId}`));
