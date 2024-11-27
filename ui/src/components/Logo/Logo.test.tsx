import { screen } from "@testing-library/dom";

import Logo from "./Logo";
import { SITE_NAME } from "consts";
import { renderComponent } from "test/utils";

test("links to the root", () => {
  renderComponent(<Logo />);
  expect(screen.getByRole("link")).toHaveAttribute("href", "/");
});

test("displays the site title", () => {
  renderComponent(<Logo />);
  expect(screen.getByText(SITE_NAME)).toBeInTheDocument();
});
