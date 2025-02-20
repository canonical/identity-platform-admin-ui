import { screen } from "@testing-library/dom";

import { renderComponent } from "test/utils";

import Pagination from "./Pagination";
import {
  mockPaginatedResponse,
  mockPaginatedResponseMeta,
} from "test/mocks/responses";
import userEvent from "@testing-library/user-event";
import { Label } from "./types";
import { Location } from "react-router";

test("displays nothing if this is the first page and there is no next page", () => {
  const { result } = renderComponent(
    <Pagination response={mockPaginatedResponse(["data1"], { _meta: {} })} />,
    {
      url: "/",
    },
  );
  expect(result.container.firstChild).toBeNull();
});

test("displays nothing if this is the first page and there is no data", () => {
  const { result } = renderComponent(
    <Pagination
      response={mockPaginatedResponse([], {
        _meta: mockPaginatedResponseMeta({ next: "token1" }),
      })}
    />,
    {
      url: "/",
    },
  );
  expect(result.container.firstChild).toBeNull();
});

test("can navigate to the first page", async () => {
  let location: Location | null = null;
  renderComponent(<Pagination />, {
    url: "/?page_token=token1",
    setLocation: (newLocation) => {
      location = newLocation;
    },
  });
  await userEvent.click(screen.getByRole("button", { name: Label.FIRST_PAGE }));
  expect((location as Location | null)?.search).toBe("?page_token=");
});

test("can navigate to the previous page", async () => {
  let location: Location | null = null;
  renderComponent(
    <Pagination
      response={mockPaginatedResponse(["data1"], {
        _meta: mockPaginatedResponseMeta({ prev: "token2" }),
      })}
    />,
    {
      url: "/?page_token=token1",
      setLocation: (newLocation) => {
        location = newLocation;
      },
    },
  );
  await userEvent.click(
    screen.getByRole("button", { name: Label.PREVIOUS_PAGE }),
  );
  expect((location as Location | null)?.search).toBe("?page_token=token2");
});

test("can navigate to the next page", async () => {
  let location: Location | null = null;
  renderComponent(
    <Pagination
      response={mockPaginatedResponse(["data1"], {
        _meta: mockPaginatedResponseMeta({ next: "token2" }),
      })}
    />,
    {
      url: "/",
      setLocation: (newLocation) => {
        location = newLocation;
      },
    },
  );
  await userEvent.click(screen.getByRole("button", { name: Label.NEXT_PAGE }));
  expect((location as Location | null)?.search).toBe("?page_token=token2");
});
