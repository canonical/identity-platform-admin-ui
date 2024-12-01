import MockAdapter from "axios-mock-adapter";

import { mockClient } from "test/mocks/clients";

import { axiosInstance } from "./axios";
import {
  fetchClient,
  fetchClients,
  createClient,
  updateClient,
  deleteClient,
} from "./client";

const mock = new MockAdapter(axiosInstance);

beforeEach(() => {
  mock.reset();
});

test("fetchClients", async () => {
  const url = "/clients?page_token=token1&size=50";
  const clients = [mockClient(), mockClient()];
  mock.onGet(url).reply(200, clients);
  await expect(fetchClients("token1")).resolves.toStrictEqual(clients);
  expect(mock.history.get[0].url).toBe(url);
});

test("fetchClient", async () => {
  const url = "/clients/client1";
  const client = mockClient();
  mock.onGet(url).reply(200, client);
  await expect(fetchClient("client1")).resolves.toStrictEqual(client);
  expect(mock.history.get[0].url).toBe(url);
});

test("createClient", async () => {
  const url = "/clients";
  const client = JSON.stringify(mockClient());
  mock.onPost(url).reply(200);
  await createClient(client);
  expect(mock.history.post[0].url).toBe(url);
  expect(mock.history.post[0].data).toBe(client);
});

test("updateClient", async () => {
  const url = "/clients/client1";
  const client = JSON.stringify(mockClient());
  mock.onPut(url).reply(200);
  await updateClient("client1", client);
  expect(mock.history.put[0].url).toBe(url);
  expect(mock.history.put[0].data).toBe(client);
});

test("deleteClient", async () => {
  const url = "/clients/client1";
  mock.onDelete(url).reply(200);
  await deleteClient("client1");
  expect(mock.history.delete[0].url).toBe(url);
});
