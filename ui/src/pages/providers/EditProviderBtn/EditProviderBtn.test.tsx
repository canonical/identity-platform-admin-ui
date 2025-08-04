import { screen, waitFor } from "@testing-library/dom";
import userEvent from "@testing-library/user-event";
import { Location } from "react-router";

import { EditPanelButtonLabel } from "components/EditPanelButton";
import { renderComponent } from "test/utils";

import EditProviderBtn from "./EditProviderBtn";

test("opens the edit provider panel", async () => {
  let location: Location | null = null;
  renderComponent(<EditProviderBtn providerId="provider1" />, {
    setLocation: (newLocation) => {
      location = newLocation;
    },
  });
  await userEvent.click(
    screen.getByRole("button", { name: EditPanelButtonLabel.EDIT }),
  );
  await waitFor(() => {
    expect(location?.search).toBe("?panel=provider-edit&id=provider1");
  });
});
