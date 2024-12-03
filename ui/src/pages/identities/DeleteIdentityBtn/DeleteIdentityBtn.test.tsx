import { screen } from "@testing-library/dom";
import userEvent from "@testing-library/user-event";

import { renderComponent } from "test/utils";

import DeleteIdentityBtn from "./DeleteIdentityBtn";
import { Label } from "./types";
import { mockIdentity } from "test/mocks/identities";
import { DeletePanelButtonLabel } from "components/DeletePanelButton";
import MockAdapter from "axios-mock-adapter";
import { axiosInstance } from "api/axios";

const mock = new MockAdapter(axiosInstance);

beforeEach(() => {
  mock.reset();
});

test("deletes the identity", async () => {
  const identity = mockIdentity();
  renderComponent(<DeleteIdentityBtn identity={identity} />);
  await userEvent.click(
    screen.getByRole("button", { name: DeletePanelButtonLabel.DELETE }),
  );
  await userEvent.click(screen.getByRole("button", { name: Label.CONFIRM }));
  expect(mock.history.delete[0].url).toBe(`/identities/${identity.id}`);
});
