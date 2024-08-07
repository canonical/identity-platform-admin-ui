import { screen } from "@testing-library/dom";
import axios from "axios";
import MockAdapter from "axios-mock-adapter";

import { authURLs } from "api/auth";
import { LoginLabel } from "components/Login";
import { renderComponent } from "test/utils";

import Layout from "./Layout";

const mock = new MockAdapter(axios);

beforeEach(() => {
  mock.reset();
});

test("displays the login screen if the user is not authenticated", async () => {
  mock.onGet(authURLs.me).reply(403, {});
  renderComponent(<Layout />);
  expect(
    await screen.findByRole("heading", { name: LoginLabel.TITLE }),
  ).toBeInTheDocument();
});

test("displays the layout and content if the user is authenticated", async () => {
  const user = {
    email: "email",
    name: "name",
    nonce: "nonce",
    sid: "sid",
    sub: "sub",
  };
  mock.onGet(authURLs.me).reply(200, { data: user });
  renderComponent(<Layout />, {
    path: "/",
    url: "/",
    routeChildren: [
      {
        element: <h1>Content</h1>,
        path: "/",
      },
    ],
  });
  expect(await screen.findByRole("heading", { name: "Content" }));
  expect(document.querySelector("#app-layout")).toBeInTheDocument();
});
