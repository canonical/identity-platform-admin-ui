import { waitFor } from "@testing-library/dom";
import MockAdapter from "axios-mock-adapter";

import { authURLs } from "api/auth";
import { changeURL } from "test/utils";

import { init, Label } from "./root";
import { axiosInstance } from "./api/axios";
import { act } from "@testing-library/react";

const mock = new MockAdapter(axiosInstance);

const cleanRootNode = () =>
  document.querySelectorAll("#app").forEach((root) => {
    document.body.removeChild(root);
  });

const createRootNode = () => {
  const root = document.createElement("div");
  root.setAttribute("id", "app");
  document.body.appendChild(root);
};

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
  createRootNode();
  changeURL("/ui");
});

afterEach(() => {
  cleanRootNode();
});

test("handles no root node", () => {
  cleanRootNode();
  expect(init).toThrowError(Label.ERROR);
});

test("initialises the index", async () => {
  act(() => init());
  await waitFor(() =>
    expect(document.querySelector("#app-layout")).toBeInTheDocument(),
  );
});
