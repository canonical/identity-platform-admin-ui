import { QueryCache, QueryClient } from "@tanstack/react-query";
import { screen, waitFor } from "@testing-library/dom";
import MockAdapter from "axios-mock-adapter";

import { authURLs } from "api/auth";
import { LoginLabel } from "components/Login";
import { renderComponent } from "test/utils";

import App from "./App";
import { axiosInstance } from "./api/axios";

const mock = new MockAdapter(axiosInstance);

beforeEach(() => {
  mock.reset();
  mock.onGet(authURLs.me).reply(200, {
    data: {
      email: "email",
      name: "name",
      nonce: "nonce",
      sid: "sid",
      sub: "sub",
    },
  });
});

test("does not display login for successful responses", async () => {
  mock.onGet("/test").reply(200, {});
  renderComponent(<App />);
  await axiosInstance.get("/test");
  await waitFor(() => {
    expect(
      screen.queryByRole("heading", { name: LoginLabel.TITLE }),
    ).not.toBeInTheDocument();
  });
});

test("does not display login for non-authentication errors", async () => {
  mock.onGet("/test").reply(500, {});
  renderComponent(<App />);
  axiosInstance.get("/test").catch(() => {
    // Don't do anything with this 500 error.
  });
  await waitFor(() => {
    expect(
      screen.queryByRole("heading", { name: LoginLabel.TITLE }),
    ).not.toBeInTheDocument();
  });
});

test("displays login for authentication errors", async () => {
  mock.onGet("/test").reply(401, {});
  renderComponent(<App />);
  axiosInstance.get("/test").catch(() => {
    // Don't do anything with this 401 error.
  });
  expect(
    await screen.findByRole("heading", { name: LoginLabel.TITLE }),
  ).toBeInTheDocument();
});

test("does not invalidate the cache on authentication errors from /auth/me", async () => {
  mock.onGet(authURLs.me).reply(401, {});
  const queryCache = new QueryCache();
  let cacheInvalidated = false;
  queryCache.subscribe((event) => {
    if ("action" in event && event.action.type === "invalidate") {
      cacheInvalidated = true;
    }
  });
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        staleTime: 1,
      },
    },
    queryCache,
  });
  renderComponent(<App />, { queryClient });
  await axiosInstance.get(authURLs.me).catch(() => {
    // Don't do anything with this 401 error.
  });
  expect(cacheInvalidated).toBeFalsy();
});
