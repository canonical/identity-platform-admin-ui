import { axiosInstance } from "api/axios";
import MockAdapter from "axios-mock-adapter";
import { ApiResponse } from "types/api";

import { handleRequest } from "./api";

const mock = new MockAdapter(axiosInstance);

type Response = { test: string };

const handler = (): Promise<ApiResponse<Response>> =>
  handleRequest(() => axiosInstance.get<ApiResponse<Response>>("/test"));

beforeEach(() => {
  mock.reset();
});

test("resolves data", async () => {
  mock.onGet("/test").reply(200, { test: "Test" });
  await expect(handler()).resolves.toMatchObject({ test: "Test" });
});

test("handles errors", async () => {
  const error = "Uh oh!";
  mock.onGet("/test").reply(500, { error });
  await expect(handler()).rejects.toBe(error);
});

test("handles error messages", async () => {
  const message = "Uh oh!";
  mock.onGet("/test").reply(500, { message });
  await expect(handler()).rejects.toBe(message);
});
