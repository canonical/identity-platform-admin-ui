import { fetchMe } from "./auth";

beforeEach(() => {
  fetchMock.resetMocks();
});

test("fetches a user", async () => {
  const user = {
    email: "email",
    name: "name",
    nonce: "nonce",
    sid: "sid",
    sub: "sub",
  };
  fetchMock.mockResponse(JSON.stringify(user), { status: 200 });
  await expect(fetchMe()).resolves.toStrictEqual(user);
});

test("handles a non-authenticated user", async () => {
  fetchMock.mockResponseOnce(JSON.stringify({}), { status: 401 });
  await expect(fetchMe()).resolves.toBeNull();
});

test("catches errors", async () => {
  const error = "Uh oh!";
  fetchMock.mockRejectedValue(error);
  await expect(fetchMe()).rejects.toBe(error);
});
