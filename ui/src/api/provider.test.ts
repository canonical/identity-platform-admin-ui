import MockAdapter from "axios-mock-adapter";
import { faker } from "@faker-js/faker";

import { mockIdentityProvider } from "test/mocks/providers";

import { axiosInstance } from "./axios";
import {
  fetchProvider,
  fetchProviders,
  createProvider,
  updateProvider,
  deleteProvider,
} from "./provider";

const mock = new MockAdapter(axiosInstance);

beforeEach(() => {
  mock.reset();
});

test("fetchProviders", async () => {
  const url = "/idps";
  const providers = [
    mockIdentityProvider({ id: faker.word.sample() }),
    mockIdentityProvider({ id: faker.word.sample() }),
  ];
  mock.onGet(url).reply(200, providers);
  await expect(fetchProviders()).resolves.toStrictEqual(providers);
  expect(mock.history.get[0].url).toBe(url);
});

test("fetchProvider", async () => {
  const url = "/idps/provider1";
  const providers = [mockIdentityProvider({ id: faker.word.sample() })];
  mock.onGet(url).reply(200, { data: providers });
  await expect(fetchProvider("provider1")).resolves.toStrictEqual(providers[0]);
  expect(mock.history.get[0].url).toBe(url);
});

test("fetchProvider handles errors", async () => {
  const error = "Uh oh!";
  mock.onGet("/idps/provider1").reply(500, { error });
  await expect(fetchProvider("provider1")).rejects.toBe(error);
});

test("createProvider", async () => {
  const url = "/idps";
  const provider = JSON.stringify(
    mockIdentityProvider({ id: faker.word.sample() }),
  );
  mock.onPost(url).reply(200);
  await createProvider(provider);
  expect(mock.history.post[0].url).toBe(url);
  expect(mock.history.post[0].data).toBe(provider);
});

test("updateProvider", async () => {
  const url = "/idps/provider1";
  const provider = JSON.stringify(
    mockIdentityProvider({ id: faker.word.sample() }),
  );
  mock.onPatch(url).reply(200);
  await updateProvider("provider1", provider);
  expect(mock.history.patch[0].url).toBe(url);
  expect(mock.history.patch[0].data).toBe(provider);
});

test("deleteProvider", async () => {
  const url = "/idps/provider1";
  mock.onDelete(url).reply(200);
  await deleteProvider("provider1");
  expect(mock.history.delete[0].url).toBe(url);
});
