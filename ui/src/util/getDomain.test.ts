import { getDomain } from "./getDomain";

test("handles extracting domain", () => {
  expect(getDomain("http://example.com/?next=/here#hash")).toBe("example.com");
});

test("handles no domain", () => {
  expect(getDomain("/a/path")).toBeUndefined();
});
