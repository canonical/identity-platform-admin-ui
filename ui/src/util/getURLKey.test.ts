import { urls } from "urls";
import { getURLKey } from "./getURLKey";
test("handles indexes", () => {
  expect(getURLKey(urls.clients.index)).toBe("clients");
});

test("handles named paths", () => {
  expect(getURLKey(urls.roles.add)).toBe("roles.add");
});

test("handles functions", () => {
  expect(getURLKey(urls.roles.edit({ id: "role1" }))).toBe("roles.edit");
});

test("handles paths that don't exist", () => {
  expect(getURLKey("/nothing/here")).toBeNull();
});
