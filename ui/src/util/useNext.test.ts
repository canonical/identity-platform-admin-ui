import { renderWrappedHook } from "test/utils";
import { useNext } from "./useNext";
import { Location } from "react-router";

vi.mock("./basePaths", async () => {
  const actual = await vi.importActual("./basePaths");
  return {
    ...actual,
    basePath: "/example/ui/",
  };
});

test("handles the 'next' path param", () => {
  let location: Location | null = null;
  renderWrappedHook(() => useNext(), {
    url: "/example/ui/?next=clients",
    setLocation: (newLocation) => {
      location = newLocation;
    },
  });
  expect((location as Location | null)?.pathname).toBe("/client");
});

test("handles no 'next' param", () => {
  let location: Location | null = null;
  renderWrappedHook(() => useNext(), {
    url: "/example/ui/current/?search=query",
    setLocation: (newLocation) => {
      location = newLocation;
    },
  });
  expect((location as Location | null)?.pathname).toBe("/example/ui/current/");
  expect((location as Location | null)?.search).toBe("?search=query");
});

test("no redirect if the next param matches the current page", () => {
  let location: Location | null = null;
  renderWrappedHook(() => useNext(), {
    url: "/example/ui/client/?next=clients",
    setLocation: (newLocation) => {
      location = newLocation;
    },
  });
  expect((location as Location | null)?.pathname).toBe("/client");
  expect((location as Location | null)?.search).toBe("");
});
