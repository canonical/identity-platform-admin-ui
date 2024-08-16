import { handleNext } from "./handleNext";

vi.mock("./basePaths", async () => {
  const actual = await vi.importActual("./basePaths");
  return {
    ...actual,
    basePath: "/example/ui/",
  };
});

test("handles the 'next' path param", () => {
  window.location.href = "/example/ui/old/?next=/example/ui/new";
  handleNext();
  expect(window.location.pathname).toBe("/example/ui/new");
});

test("handles the 'next' param with domain", () => {
  window.location.href =
    "http://example.com/example/ui/old/?next=http://example.com/example/ui/new";
  handleNext();
  expect(window.location.pathname).toBe("/example/ui/new");
});

test("handles no 'next' param", () => {
  window.location.href = "/example/ui/current/?search=query";
  handleNext();
  expect(window.location.pathname).toBe("/example/ui/current/");
  expect(window.location.search).toBe("?search=query");
});

test("no redirect if the next param matches the current page", () => {
  window.location.href = "/example/ui/current/?next=/example/ui/current";
  handleNext();
  expect(window.location.pathname).toBe("/example/ui/current/");
  expect(window.location.search).toBe("");
});

test("no redirect if the next param has a different domain", () => {
  window.location.href =
    "http://example.com/example/ui/old/?next=http://notexample.com/example/ui/new";
  handleNext();
  expect(window.location.host).toBe("example.com");
  expect(window.location.pathname).toBe("/example/ui/old/");
  expect(window.location.search).toBe("");
});

test("no redirect if the next param has a different base", () => {
  window.location.href = "/example/ui/current/?next=/api/delete";
  handleNext();
  expect(window.location.pathname).toBe("/example/ui/current/");
  expect(window.location.search).toBe("");
});
