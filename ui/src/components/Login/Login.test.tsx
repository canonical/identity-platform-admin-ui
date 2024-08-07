import { renderComponent } from "test/utils";
import Login from "./Login";
import { screen, within } from "@testing-library/dom";

test("displays the loading state", () => {
  renderComponent(<Login isLoading />);
  expect(
    within(screen.getByRole("alert")).getByText("Loading"),
  ).toBeInTheDocument();
});

test("displays errors", () => {
  const error = "Uh oh!";
  renderComponent(<Login error={error} />);
  expect(within(screen.getByRole("code")).getByText(error)).toBeInTheDocument();
});

test("displays the login button", () => {
  renderComponent(<Login />);
  expect(
    screen.getByRole("link", { name: "Sign in to Identity platform" }),
  ).toBeInTheDocument();
});
