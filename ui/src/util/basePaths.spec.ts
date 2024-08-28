import {
  appendBasePath,
  calculateBasePath,
  appendAPIBasePath,
} from "./basePaths";

vi.mock("./basePaths", async () => {
  vi.stubGlobal("location", { pathname: "/example/ui/" });
  const actual = await vi.importActual("./basePaths");
  return {
    ...actual,
    basePath: "/example/ui/",
    apiBasePath: "/example/ui/../api/v0/",
  };
});

describe("calculateBasePath", () => {
  it("resolves with ui path", () => {
    vi.stubGlobal("location", { pathname: "/ui/" });
    const result = calculateBasePath();
    expect(result).toBe("/ui/");
  });

  it("resolves with ui path without trailing slash", () => {
    vi.stubGlobal("location", { pathname: "/ui" });
    const result = calculateBasePath();
    expect(result).toBe("/ui/");
  });

  it("resolves with ui path and discards detail page location", () => {
    vi.stubGlobal("location", { pathname: "/ui/foo/bar" });
    const result = calculateBasePath();
    expect(result).toBe("/ui/");
  });

  it("resolves with prefixed ui path", () => {
    vi.stubGlobal("location", { pathname: "/prefix/ui/" });
    const result = calculateBasePath();
    expect(result).toBe("/prefix/ui/");
  });

  it("resolves with prefixed ui path on a detail page", () => {
    vi.stubGlobal("location", { pathname: "/prefix/ui/foo/bar/baz" });
    const result = calculateBasePath();
    expect(result).toBe("/prefix/ui/");
  });

  it("resolves with root path if /ui/ is not part of the pathname", () => {
    vi.stubGlobal("location", { pathname: "/foo/bar/baz" });
    const result = calculateBasePath();
    expect(result).toBe("/");
  });

  it("resolves with root path for partial ui substrings", () => {
    vi.stubGlobal("location", { pathname: "/prefix/uipartial" });
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
    expect(appendAPIBasePath("/test")).toBe("/example/ui/../api/v0/test");
  });

  it("handles paths without a leading slash", () => {
    expect(appendAPIBasePath("test")).toBe("/example/ui/../api/v0/test");
  });
});
