import { calculateBasePath } from "./basePaths";

describe("calculateBasePath", () => {
  it("resolves with ui path", () => {
    vi.stubGlobal("location", { pathname: "/ui/" });
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
});
