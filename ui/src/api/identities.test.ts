import MockAdapter from "axios-mock-adapter";

import { mockIdentity } from "test/mocks/identities";

import { axiosInstance } from "./axios";
import {
  fetchIdentity,
  fetchIdentities,
  createIdentity,
  updateIdentity,
  deleteIdentity,
} from "./identities";

const mock = new MockAdapter(axiosInstance);

beforeEach(() => {
  mock.reset();
});

test("fetchIdentities", async () => {
  const url = "/identities?page_token=token1&size=50";
  const identities = [mockIdentity(), mockIdentity()];
  mock.onGet(url).reply(200, identities);
  const response = await fetchIdentities("token1");
  expect(mock.history.get[0].url).toBe(url);
  expect(response).toStrictEqual(identities);
});

test("fetchIdentity", async () => {
  const url = "/identities/identity1";
  const identity = mockIdentity();
  mock.onGet(url).reply(200, identity);
  const response = await fetchIdentity("identity1");
  expect(mock.history.get[0].url).toBe(url);
  expect(response).toStrictEqual(identity);
});

test("createIdentity", async () => {
  const url = "/identities";
  const identity = JSON.stringify(mockIdentity());
  mock.onPost(url).reply(200);
  await createIdentity(identity);
  expect(mock.history.post[0].url).toBe(url);
  expect(mock.history.post[0].data).toBe(identity);
});

test("updateIdentity", async () => {
  const url = "/identities/identity1";
  const identity = JSON.stringify(mockIdentity());
  mock.onPut(url).reply(200);
  await updateIdentity("identity1", identity);
  expect(mock.history.put[0].url).toBe(url);
  expect(mock.history.put[0].data).toBe(identity);
});

test("deleteIdentity", async () => {
  const url = "/identities/identity1";
  mock.onDelete(url).reply(200);
  await deleteIdentity("identity1");
  expect(mock.history.delete[0].url).toBe(url);
});
