import { renderComponent } from "test/utils";
import Layout from "./Layout";
import { screen } from "@testing-library/dom";
import { LoginLabel } from "components/Login";

beforeEach(() => {
  fetchMock.resetMocks();
});

test("displays the login screen if the user is not authenticated", async () => {
  fetchMock.mockResponseOnce(JSON.stringify({}), { status: 403 });
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
  fetchMock.mockResponse(JSON.stringify(user), { status: 200 });
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
