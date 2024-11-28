import { renderWrappedHook } from "test/utils";
import { usePagination } from "./usePagination";
import { Location } from "react-router-dom";
import { act } from "@testing-library/react";

test("gets the token from the query params", () => {
  const { result } = renderWrappedHook(() => usePagination(), {
    url: "/?page_token=token1",
  });
  expect(result.current.pageToken).toBe("token1");
});

test("can set the token", () => {
  let location: Location | null = null;
  const { result } = renderWrappedHook(() => usePagination(), {
    url: "/?search=query",
    setLocation: (newLocation) => {
      location = newLocation;
    },
  });
  act(() => result.current.setPageToken("token1"));
  expect((location as Location | null)?.search).toBe(
    "?search=query&page_token=token1",
  );
});
