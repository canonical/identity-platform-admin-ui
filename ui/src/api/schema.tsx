import { ApiResponse, PaginatedResponse } from "types/api";
import { handleResponse, PAGE_SIZE } from "util/api";
import { Schema } from "types/schema";
import { apiBasePath } from "util/basePaths";

export const fetchSchemas = (
  pageToken: string,
): Promise<PaginatedResponse<Schema[]>> => {
  return new Promise((resolve, reject) => {
    fetch(
      `${apiBasePath}schemas?page_token=${pageToken}&page_size=${PAGE_SIZE}`,
    )
      .then(handleResponse)
      .then((result: PaginatedResponse<Schema[]>) => resolve(result))
      .catch(reject);
  });
};

export const fetchSchema = (schemaId: string): Promise<Schema> => {
  return new Promise((resolve, reject) => {
    fetch(`${apiBasePath}schemas/${schemaId}`)
      .then(handleResponse)
      .then((result: ApiResponse<Schema[]>) => resolve(result.data[0]))
      .catch(reject);
  });
};

export const createSchema = (body: string): Promise<void> => {
  return new Promise((resolve, reject) => {
    fetch(`${apiBasePath}schemas`, {
      method: "POST",
      body: body,
    })
      .then(handleResponse)
      .then(resolve)
      .catch(reject);
  });
};

export const updateSchema = (
  schemaId: string,
  values: string,
): Promise<Schema> => {
  return new Promise((resolve, reject) => {
    fetch(`${apiBasePath}schemas/${schemaId}`, {
      method: "PATCH",
      body: values,
    })
      .then(handleResponse)
      .then((result: ApiResponse<Schema>) => resolve(result.data))
      .catch(reject);
  });
};

export const deleteSchema = (schemaId: string) => {
  return new Promise((resolve, reject) => {
    fetch(`${apiBasePath}schemas/${schemaId}`, {
      method: "DELETE",
    })
      .then(resolve)
      .catch(reject);
  });
};
