import { urls } from "urls";
import { keyToPath } from "./keyToPath";

test("handles indexes", () => {
  expect(keyToPath("clients")).toBe(urls.clients.index);
});

test("handles named routes", () => {
  expect(keyToPath("roles.add")).toBe(urls.roles.add);
});

test("handles functions", () => {
  expect(keyToPath("roles.edit")).toBe(urls.roles.edit(null));
});

test("handles keys that don't exist", () => {
  expect(keyToPath("roles.nothing")).toBeNull();
});
test("handles malformed keys", () => {
  expect(keyToPath("alert('you have been hacked!');")).toBeNull();
});
