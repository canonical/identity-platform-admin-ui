import { screen, within } from "@testing-library/dom";
import { render } from "@testing-library/react";

import SidePanel from "./SidePanel";
import { Label } from "./types";
import { LoaderTestId } from "components/Loader";

describe("SidePanel", () => {
  test("displays content", () => {
    const content = "Content";
    render(<SidePanel>{content}</SidePanel>);
    expect(
      screen.getByRole("complementary", {
        name: Label.SIDE_PANEL,
      }),
    ).toHaveTextContent(content);
  });

  test("can display as normal width", () => {
    render(<SidePanel>Content</SidePanel>);
    const sidePanel = screen.getByRole("complementary", {
      name: Label.SIDE_PANEL,
    });
    expect(sidePanel).not.toHaveClass("is-wide");
    expect(sidePanel).not.toHaveClass("is-narrow");
  });

  test("can display as wide", () => {
    render(<SidePanel width="wide">Content</SidePanel>);
    expect(
      screen.getByRole("complementary", {
        name: Label.SIDE_PANEL,
      }),
    ).toHaveClass("is-wide");
  });

  test("can display as narrow", () => {
    render(<SidePanel width="narrow">Content</SidePanel>);
    expect(
      screen.getByRole("complementary", {
        name: Label.SIDE_PANEL,
      }),
    ).toHaveClass("is-narrow");
  });

  test("can display as split", () => {
    render(<SidePanel isSplit>Content</SidePanel>);
    expect(
      screen.getByRole("complementary", {
        name: Label.SIDE_PANEL,
      }),
    ).toHaveClass("is-split");
  });

  test("can display as an overlay", () => {
    render(<SidePanel isOverlay>Content</SidePanel>);
    expect(
      screen.getByRole("complementary", {
        name: Label.SIDE_PANEL,
      }),
    ).toHaveClass("is-overlay");
  });

  test("can display as pinned", () => {
    render(<SidePanel pinned>Content</SidePanel>);
    expect(
      screen.getByRole("complementary", {
        name: Label.SIDE_PANEL,
      }),
    ).toHaveClass("is-pinned");
  });

  test("can display loading state", () => {
    render(<SidePanel loading>Content</SidePanel>);
    expect(
      within(
        screen.getByRole("complementary", {
          name: Label.SIDE_PANEL,
        }),
      ).getByTestId(LoaderTestId.COMPONENT),
    ).toBeInTheDocument();
  });

  test("can display loading error", () => {
    render(<SidePanel hasError>Content</SidePanel>);
    expect(
      within(
        screen.getByRole("complementary", {
          name: Label.SIDE_PANEL,
        }),
      ).getByText(Label.ERROR_LOADING),
    ).toBeInTheDocument();
  });
});

test("HeaderControls", () => {
  const content = "Content";
  render(<SidePanel.HeaderControls>{content}</SidePanel.HeaderControls>);
  expect(screen.getByText(content)).toHaveClass("p-panel__controls");
});

test("HeaderTitle", () => {
  const content = "Content";
  render(<SidePanel.HeaderTitle>{content}</SidePanel.HeaderTitle>);
  expect(screen.getByText(content)).toHaveClass("p-panel__title");
});
test("Sticky", () => {
  const content = "Content";
  render(<SidePanel.Sticky>{content}</SidePanel.Sticky>);
  expect(screen.getByText(content)).toHaveClass("sticky-wrapper");
});

test("Header", () => {
  const content = "Content";
  render(<SidePanel.Header>{content}</SidePanel.Header>);
  expect(screen.getByText(content)).toHaveClass("p-panel__header");
});

test("Container", () => {
  const content = "Content";
  render(<SidePanel.Container>{content}</SidePanel.Container>);
  expect(screen.getByText(content).parentElement).toHaveClass("p-panel");
});

test("Content", () => {
  const content = "Content";
  render(<SidePanel.Content>{content}</SidePanel.Content>);
  expect(screen.getByText(content)).toHaveClass("p-panel__content");
});

test("Footer", () => {
  const content = "Content";
  render(<SidePanel.Footer>{content}</SidePanel.Footer>);
  expect(screen.getByText(content)).toHaveClass("panel-footer");
});
