import { screen } from "@testing-library/dom";
import userEvent from "@testing-library/user-event";

import { renderComponent } from "test/utils";

import DeleteProviderBtn from "./DeleteProviderBtn";
import { Label } from "./types";
import { mockIdentityProvider } from "test/mocks/providers";
import { DeletePanelButtonLabel } from "components/DeletePanelButton";
import MockAdapter from "axios-mock-adapter";
import { axiosInstance } from "api/axios";
import {
  NotificationConsumer,
  NotificationProvider,
} from "@canonical/react-components";

const mock = new MockAdapter(axiosInstance);

beforeEach(() => {
  mock.reset();
});

test("disables confirmation if the confirm text hasn't been entered", async () => {
  const provider = mockIdentityProvider({ id: "provider1" });
  renderComponent(<DeleteProviderBtn provider={provider} />);
  await userEvent.click(
    screen.getByRole("button", { name: DeletePanelButtonLabel.DELETE }),
  );
  expect(screen.getByRole("button", { name: Label.CONFIRM })).toBeDisabled();
});

test("deletes the provider", async () => {
  const provider = mockIdentityProvider({ id: "provider1" });
  renderComponent(<DeleteProviderBtn provider={provider} />);
  await userEvent.click(
    screen.getByRole("button", { name: DeletePanelButtonLabel.DELETE }),
  );
  await userEvent.type(screen.getByRole("textbox"), `remove ${provider.id}`);
  await userEvent.click(screen.getByRole("button", { name: Label.CONFIRM }));
  expect(mock.history.delete[0].url).toBe(`/idps/${provider.id}`);
});

test("displays an error if the provider doesn't have an id", async () => {
  const provider = mockIdentityProvider({ id: undefined });
  renderComponent(
    <NotificationProvider>
      <NotificationConsumer />
      <DeleteProviderBtn provider={provider} />
    </NotificationProvider>,
  );
  await userEvent.click(
    screen.getByRole("button", { name: DeletePanelButtonLabel.DELETE }),
  );
  await userEvent.type(screen.getByRole("textbox"), "remove ");
  await userEvent.click(screen.getByRole("button", { name: Label.CONFIRM }));
  expect(mock.history.delete).toHaveLength(0);
  expect(
    screen
      .getByText("Provider deletion failed")
      .closest(".p-notification--negative"),
  ).toBeInTheDocument();
});
