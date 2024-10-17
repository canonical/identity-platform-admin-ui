// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

import {
  appendBasePath,
  calculateBasePath,
  appendAPIBasePath,
} from "./basePaths";

vi.mock("./basePaths", async () => {
  const basePath = "/example/ui";
  // Can't use the setBase function here as this mock gets hoisted to the top of
  // the file so gets executed before setBase exists.
  const base = document.createElement("base");
  base.setAttribute("href", basePath);
  document.body.appendChild(base);
  const actual = await vi.importActual("./basePaths");
  return {
    ...actual,
    basePath,
  };
});

const cleanBase = () =>
  document.querySelectorAll("base").forEach((base) => {
    document.body.removeChild(base);
  });

const setBase = (href: string) => {
  const base = document.createElement("base");
  base.setAttribute("href", href);
  document.body.appendChild(base);
};

// Clean up the base created by the mock above.
cleanBase();

describe("calculateBasePath", () => {
  afterEach(() => {
    cleanBase();
  });

  it("resolves with ui path", () => {
    setBase("/test/");
    const result = calculateBasePath();
    expect(result).toBe("/test/");
  });

  it("resolves with ui path without trailing slash", () => {
    setBase("/test");
    const result = calculateBasePath();
    expect(result).toBe("/test/");
  });

  it("handles full URL", () => {
    setBase("http://example.com/test/");
    const result = calculateBasePath();
    expect(result).toBe("/test/");
  });

  it("resolves if the base is not provided", () => {
    const result = calculateBasePath();
    expect(result).toBe("/ui/");
  });
});

describe("appendBasePath", () => {
  it("handles paths with a leading slash", () => {
    setBase("/test/");
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
