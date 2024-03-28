import { ApiResponse } from "types/api";
import { handleResponse } from "util/api";
import { Schema } from "types/schema";

export const fetchSchemas = (): Promise<Schema[]> => {
  return new Promise((resolve, reject) => {
    fetch("/api/v0/schemas")
      .then(handleResponse)
      .then((result: ApiResponse<Schema[]>) => resolve(result.data))
      .catch(reject);
  });
};

export const fetchSchema = (schemaId: string): Promise<Schema> => {
  return new Promise((resolve, reject) => {
    fetch(`/api/v0/schemas/${schemaId}`)
      .then(handleResponse)
      .then((result: ApiResponse<Schema[]>) => resolve(result.data[0]))
      .catch(reject);
  });
};

export const createSchema = (body: string): Promise<void> => {
  return new Promise((resolve, reject) => {
    fetch("/api/v0/schemas", {
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
    fetch(`/api/v0/schemas/${schemaId}`, {
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
    fetch(`/api/v0/schemas/${schemaId}`, {
      method: "DELETE",
    })
      .then(resolve)
      .catch(reject);
  });
};
