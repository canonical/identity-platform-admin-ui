import { renderComponent } from "test/utils";
import { screen, within } from "@testing-library/dom";

import { urls } from "urls";

import Navigation from "./Navigation";
import { Label } from "./types";
import { authURLs } from "api/auth";
import { appendAPIBasePath } from "util/basePaths";

test("displays page links", () => {
  renderComponent(<Navigation />);
  expect(
    screen.getByRole("link", { name: Label.IDENTITY_PROVIDERS }),
  ).toHaveAttribute("href", urls.providers.index);
  expect(screen.getByRole("link", { name: Label.CLIENTS })).toHaveAttribute(
    "href",
    urls.clients.index,
  );
  expect(screen.getByRole("link", { name: Label.IDENTITIES })).toHaveAttribute(
    "href",
    urls.identities.index,
  );
  expect(screen.getByRole("link", { name: Label.SCHEMAS })).toHaveAttribute(
    "href",
    urls.schemas.index,
  );
  expect(screen.getByRole("link", { name: "Groups" })).toHaveAttribute(
    "href",
    urls.groups.index,
  );
  expect(screen.getByRole("link", { name: "Roles" })).toHaveAttribute(
    "href",
    urls.roles.index,
  );
});

test("displays username", () => {
  renderComponent(<Navigation username="test@example.com" />);
  const userNav = document.querySelector(".p-side-navigation--user-menu");
  expect(userNav).toBeInTheDocument();
  if (userNav) {
    expect(
      within(userNav as HTMLElement).getByText("test@example.com"),
    ).toBeInTheDocument();
  }
});

test("has a logout link", () => {
  renderComponent(<Navigation username="test@example.com" />);
  const userNav = document.querySelector(".p-side-navigation--user-menu");
  expect(userNav).toBeInTheDocument();
  if (userNav) {
    expect(
      within(userNav as HTMLElement).getByRole("link", { name: Label.LOGOUT }),
    ).toHaveAttribute("href", appendAPIBasePath(authURLs.logout));
  }
});
