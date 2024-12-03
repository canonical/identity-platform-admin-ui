import { screen } from "@testing-library/dom";
import userEvent from "@testing-library/user-event";

import { renderComponent } from "test/utils";

import DeleteClientBtn from "./DeleteClientBtn";
import { mockClient } from "test/mocks/clients";
import { DeletePanelButtonLabel } from "components/DeletePanelButton";
import MockAdapter from "axios-mock-adapter";
import { axiosInstance } from "api/axios";
import { Label } from "./types";

const mock = new MockAdapter(axiosInstance);

beforeEach(() => {
  mock.reset();
});

test("deletes the client", async () => {
  const client = mockClient();
  renderComponent(<DeleteClientBtn client={client} />);
  await userEvent.click(
    screen.getByRole("button", { name: DeletePanelButtonLabel.DELETE }),
  );
  await userEvent.click(screen.getByRole("button", { name: Label.CONFIRM }));
  expect(mock.history.delete[0].url).toBe(`/clients/${client.client_id}`);
});
