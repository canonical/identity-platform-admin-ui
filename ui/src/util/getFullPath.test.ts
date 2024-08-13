import { getFullPath } from "./getFullPath";

test("handles paths with query and hash", () => {
  expect(getFullPath("http://example.com/?next=/here#hash")).toBe(
    "/?next=/here#hash",
  );
});

test("handles no path", () => {
  expect(getFullPath("http://example.com/")).toBeUndefined();
});
