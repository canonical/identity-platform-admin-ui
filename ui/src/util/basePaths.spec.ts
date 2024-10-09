import {
  appendBasePath,
  calculateBasePath,
  appendAPIBasePath,
} from "./basePaths";

vi.mock("./basePaths", async () => {
  window.base = "/example";
  const actual = await vi.importActual("./basePaths");
  return {
    ...actual,
    basePath: "/example/ui/",
    apiBasePath: "/example/api/v0/",
  };
});

describe("calculateBasePath", () => {
  it("resolves with ui path", () => {
    window.base = "/test/";
    const result = calculateBasePath();
    expect(result).toBe("/test/");
  });

  it("resolves with ui path without trailing slash", () => {
    window.base = "/test";
    const result = calculateBasePath();
    expect(result).toBe("/test/");
  });

  it("resolves with root path if the base is not provided", () => {
    if (window.base) {
      delete window.base;
    }
    const result = calculateBasePath();
    expect(result).toBe("/");
  });
});

describe("appendBasePath", () => {
  it("handles paths with a leading slash", () => {
    expect(appendBasePath("/test")).toBe("/example/ui/test");
  });

  it("handles paths without a leading slash", () => {
    expect(appendBasePath("test")).toBe("/example/ui/test");
  });
});

describe("appendAPIBasePath", () => {
  it("handles paths with a leading slash", () => {
    expect(appendAPIBasePath("/test")).toBe("/example/api/v0/test");
  });

  it("handles paths without a leading slash", () => {
    expect(appendAPIBasePath("test")).toBe("/example/api/v0/test");
  });
});
