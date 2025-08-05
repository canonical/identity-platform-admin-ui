import { renderWrappedHook } from "test/utils";
import { useNext } from "./useNext";
import { Location } from "react-router";
import { waitFor } from "@testing-library/dom";

vi.mock("./basePaths", async () => {
  const actual = await vi.importActual("./basePaths");
  return {
    ...actual,
    basePath: "/example/ui/",
  };
});

test("handles the 'next' path param", async () => {
  let location: Location | null = null;
  renderWrappedHook(() => useNext(), {
    url: "/example/ui/?next=clients",
    setLocation: (newLocation) => {
      location = newLocation;
    },
  });
  await waitFor(() => {
    expect(location?.pathname).toBe("/client");
  });
});

test("handles no 'next' param", async () => {
  let location: Location | null = null;
  renderWrappedHook(() => useNext(), {
    url: "/example/ui/current/?search=query",
    setLocation: (newLocation) => {
      location = newLocation;
    },
  });
  await waitFor(() => {
    expect(location?.pathname).toBe("/example/ui/current/");
    expect(location?.search).toBe("?search=query");
  });
});

test("no redirect if the next param matches the current page", async () => {
  let location: Location | null = null;
  renderWrappedHook(() => useNext(), {
    url: "/example/ui/client/?next=clients",
    setLocation: (newLocation) => {
      location = newLocation;
    },
  });
  await waitFor(() => {
    expect(location?.pathname).toBe("/client");
    expect(location?.search).toBe("");
  });
});
