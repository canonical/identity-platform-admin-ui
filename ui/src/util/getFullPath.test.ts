import { getFullPath } from "./getFullPath";

vi.mock("./basePaths", async () => {
  const actual = await vi.importActual("./basePaths");
  return {
    ...actual,
    basePath: "/example/ui/",
  };
});

test("handles paths with query and hash", () => {
  expect(getFullPath("http://example.com/?next=/here#hash")).toBe(
    "/?next=/here#hash",
  );
});

test("removes base", () => {
  expect(
    getFullPath("http://example.com/example/ui/roles/?next=/here#hash", true),
  ).toBe("/roles/?next=/here#hash");
});

test("handles no path", () => {
  expect(getFullPath("http://example.com/")).toBeUndefined();
});
