import { screen, waitFor } from "@testing-library/dom";
import * as reactQuery from "@tanstack/react-query";
import userEvent from "@testing-library/user-event";
import {
  NotificationConsumer,
  NotificationProvider,
} from "@canonical/react-components";

import { renderComponent } from "test/utils";

import DeletePanelButton from "./DeletePanelButton";
import { Label } from "./types";
import { Location } from "react-router";

vi.mock("@tanstack/react-query", async () => {
  const actual = await vi.importActual("@tanstack/react-query");
  return {
    ...actual,
    useQueryClient: vi.fn(),
  };
});

beforeEach(() => {
  vi.spyOn(reactQuery, "useQueryClient").mockReturnValue({
    invalidateQueries: vi.fn(),
  } as unknown as reactQuery.QueryClient);
});

test("displays the delete button", () => {
  renderComponent(
    <DeletePanelButton
      confirmButtonLabel="Confirm"
      confirmContent="Content"
      entityName="Nebulous"
      invalidateQuery="nebulous"
      onDelete={() => Promise.resolve()}
      successPath="/nebulous"
      successMessage="successfully formed"
    />,
  );
  expect(
    screen.getByRole("button", { name: Label.DELETE }),
  ).toBeInTheDocument();
});

test("displays a confirmation", async () => {
  renderComponent(
    <DeletePanelButton
      confirmButtonLabel="Confirm"
      confirmContent="Content"
      confirmTitle="Define"
      entityName="Nebulous"
      invalidateQuery="nebulous"
      onDelete={() => Promise.resolve()}
      successPath="/nebulous"
      successMessage="successfully formed"
    />,
  );
  await userEvent.click(screen.getByRole("button", { name: Label.DELETE }));
  expect(screen.getByRole("dialog", { name: "Define" })).toBeInTheDocument();
});

test("can disable the confirm button", async () => {
  renderComponent(
    <DeletePanelButton
      confirmButtonLabel="Confirm"
      confirmButtonDisabled
      confirmContent="Content"
      entityName="Nebulous"
      invalidateQuery="nebulous"
      onDelete={() => Promise.resolve()}
      successPath="/nebulous"
      successMessage="successfully formed"
    />,
  );
  await userEvent.click(screen.getByRole("button", { name: Label.DELETE }));
  expect(screen.getByRole("button", { name: "Confirm" })).toBeDisabled();
});

test("starts deletion", async () => {
  renderComponent(
    <DeletePanelButton
      confirmButtonLabel="Confirm"
      confirmContent="Content"
      entityName="Nebulous"
      invalidateQuery="nebulous"
      onDelete={() => Promise.resolve()}
      successPath="/nebulous"
      successMessage="successfully formed"
    />,
  );
  await userEvent.click(screen.getByRole("button", { name: Label.DELETE }));
  await userEvent.click(screen.getByRole("button", { name: "Confirm" }));
  expect(document.querySelector(".u-animation--spin")).toBeInTheDocument();
});

test("calls the delete method", async () => {
  const onDelete = vi.fn().mockImplementation(() => Promise.resolve());
  renderComponent(
    <DeletePanelButton
      confirmButtonLabel="Confirm"
      confirmContent="Content"
      entityName="Nebulous"
      invalidateQuery="nebulous"
      onDelete={onDelete}
      successPath="/nebulous"
      successMessage="successfully formed"
    />,
  );
  await userEvent.click(screen.getByRole("button", { name: Label.DELETE }));
  await userEvent.click(screen.getByRole("button", { name: "Confirm" }));
  expect(onDelete).toHaveBeenCalled();
});

test("handles a successful delete call", async () => {
  let location: Location | null = null;
  renderComponent(
    <NotificationProvider>
      <NotificationConsumer />
      <DeletePanelButton
        confirmButtonLabel="Confirm"
        confirmContent="Content"
        entityName="Nebulous"
        invalidateQuery="nebulous"
        onDelete={() => Promise.resolve()}
        successPath="/nebulous"
        successMessage="successfully formed"
      />
    </NotificationProvider>,
    {
      setLocation: (newLocation) => {
        location = newLocation;
      },
    },
  );
  await userEvent.click(screen.getByRole("button", { name: Label.DELETE }));
  await userEvent.click(screen.getByRole("button", { name: "Confirm" }));
  const message = await screen.findByText("successfully formed");
  expect(message.closest(".p-notification--positive")).toBeInTheDocument();
  await waitFor(() => {
    expect(location?.pathname).toBe("/nebulous");
  });
});

test("notifies on error", async () => {
  renderComponent(
    <NotificationProvider>
      <NotificationConsumer />
      <DeletePanelButton
        confirmButtonLabel="Confirm"
        confirmContent="Content"
        entityName="Nebulous"
        invalidateQuery="nebulous"
        onDelete={() => Promise.reject("Oops")}
        successPath="/nebulous"
        successMessage="successfully formed"
      />
    </NotificationProvider>,
  );
  await userEvent.click(screen.getByRole("button", { name: Label.DELETE }));
  await userEvent.click(screen.getByRole("button", { name: "Confirm" }));
  expect(
    screen
      .getByText("Nebulous deletion failed")
      .closest(".p-notification--negative"),
  ).toBeInTheDocument();
  expect(screen.getByText("Oops")).toHaveClass("p-notification__message");
});

test("notifies on error object", async () => {
  renderComponent(
    <NotificationProvider>
      <NotificationConsumer />
      <DeletePanelButton
        confirmButtonLabel="Confirm"
        confirmContent="Content"
        entityName="Nebulous"
        invalidateQuery="nebulous"
        onDelete={() => Promise.reject(new Error("Oops"))}
        successPath="/nebulous"
        successMessage="successfully formed"
      />
    </NotificationProvider>,
  );
  await userEvent.click(screen.getByRole("button", { name: Label.DELETE }));
  await userEvent.click(screen.getByRole("button", { name: "Confirm" }));
  expect(
    screen
      .getByText("Nebulous deletion failed")
      .closest(".p-notification--negative"),
  ).toBeInTheDocument();
  expect(screen.getByText("Oops")).toHaveClass("p-notification__message");
});

test("invlidates queries and hides the spinner on success", async () => {
  const invalidateQueries = vi.fn();
  vi.spyOn(reactQuery, "useQueryClient").mockReturnValue({
    invalidateQueries,
  } as unknown as reactQuery.QueryClient);
  renderComponent(
    <DeletePanelButton
      confirmButtonLabel="Confirm"
      confirmContent="Content"
      entityName="Nebulous"
      invalidateQuery="nebulous"
      onDelete={() => Promise.resolve()}
      successPath="/nebulous"
      successMessage="successfully formed"
    />,
  );
  await userEvent.click(screen.getByRole("button", { name: Label.DELETE }));
  await userEvent.click(screen.getByRole("button", { name: "Confirm" }));
  await waitFor(() =>
    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: ["nebulous"],
    }),
  );
  await waitFor(() =>
    expect(
      document.querySelector(".u-animation--spin"),
    ).not.toBeInTheDocument(),
  );
});
