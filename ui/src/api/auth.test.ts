import axios from "axios";
import MockAdapter from "axios-mock-adapter";

import { fetchMe, authURLs } from "./auth";

const mock = new MockAdapter(axios);

beforeEach(() => {
  mock.reset();
});

test("fetches a user", async () => {
  const user = {
    email: "email",
    name: "name",
    nonce: "nonce",
    sid: "sid",
    sub: "sub",
  };
  mock.onGet(authURLs.me).reply(200, user);
  await expect(fetchMe()).resolves.toStrictEqual(user);
});

test("handles a non-authenticated user", async () => {
  mock.onGet(authURLs.me).reply(401, {});
  await expect(fetchMe()).resolves.toBeNull();
});

test("catches errors", async () => {
  const error = "Uh oh!";
  mock.onGet(authURLs.me).reply(500, { error });
  await expect(fetchMe()).rejects.toBe(error);
});
