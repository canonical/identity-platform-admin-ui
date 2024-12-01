import MockAdapter from "axios-mock-adapter";

import { mockSchema } from "test/mocks/schemas";

import { axiosInstance } from "./axios";
import {
  fetchSchema,
  fetchSchemas,
  createSchema,
  updateSchema,
  deleteSchema,
} from "./schema";

const mock = new MockAdapter(axiosInstance);

beforeEach(() => {
  mock.reset();
});

test("fetchSchemas", async () => {
  const url = "/schemas?page_token=token1&page_size=50";
  const schemas = [mockSchema(), mockSchema()];
  mock.onGet(url).reply(200, schemas);
  await expect(fetchSchemas("token1")).resolves.toStrictEqual(schemas);
  expect(mock.history.get[0].url).toBe(url);
});

test("fetchSchema", async () => {
  const url = "/schemas/schema1";
  const schemas = [mockSchema()];
  mock.onGet(url).reply(200, { data: schemas });
  await expect(fetchSchema("schema1")).resolves.toStrictEqual(schemas[0]);
  expect(mock.history.get[0].url).toBe(url);
});

test("fetchSchema handles errors", async () => {
  const error = "Uh oh!";
  mock.onGet("/schemas/schema1").reply(500, { error });
  await expect(fetchSchema("schema1")).rejects.toBe(error);
});

test("createSchema", async () => {
  const url = "/schemas";
  const schema = JSON.stringify(mockSchema());
  mock.onPost(url).reply(200);
  await createSchema(schema);
  expect(mock.history.post[0].url).toBe(url);
  expect(mock.history.post[0].data).toBe(schema);
});

test("updateSchema", async () => {
  const url = "/schemas/schema1";
  const schema = JSON.stringify(mockSchema());
  mock.onPatch(url).reply(200);
  await updateSchema("schema1", schema);
  expect(mock.history.patch[0].url).toBe(url);
  expect(mock.history.patch[0].data).toBe(schema);
});

test("deleteSchema", async () => {
  const url = "/schemas/schema1";
  mock.onDelete(url).reply(200);
  await deleteSchema("schema1");
  expect(mock.history.delete[0].url).toBe(url);
});
