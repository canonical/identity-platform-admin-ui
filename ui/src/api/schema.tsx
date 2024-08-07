import { PaginatedResponse, ApiResponse } from "types/api";
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
    handleRequest<Schema[], ApiResponse<Schema[]>>(() =>
      axios.get<ApiResponse<Schema[]>>(`/schemas/${schemaId}`),
    )
      .then(({ data }) => resolve(data[0]))
      .catch((error) => reject(error));
  });
};

export const createSchema = (body: string): Promise<unknown> =>
  handleRequest(() => axios.post("/schemas", body));

export const updateSchema = (
  schemaId: string,
  values: string,
): Promise<ApiResponse<Schema>> =>
  handleRequest(() =>
    axios.patch<ApiResponse<Schema>>(`/schemas/${schemaId}`, values),
  );

export const deleteSchema = (schemaId: string): Promise<unknown> =>
  handleRequest(() => axios.delete(`/schemas/${schemaId}`));
