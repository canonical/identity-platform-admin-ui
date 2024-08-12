import { getFullPath } from "./getFullPath";

test("handles paths with query and hash", () => {
  window.location.href = "http://example.com/?next=/here#hash";
  expect(getFullPath()).toBe("/?next=/here#hash");
});

test("handles no path", () => {
  window.location.href = "http://example.com/";
  expect(getFullPath()).toBeUndefined();
});
