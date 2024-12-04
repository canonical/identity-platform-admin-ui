import { screen } from "@testing-library/dom";
import userEvent from "@testing-library/user-event";
import { Location } from "react-router-dom";

import { EditPanelButtonLabel } from "components/EditPanelButton";
import { renderComponent } from "test/utils";

import EditClientBtn from "./EditClientBtn";

test("opens the edit client panel", async () => {
  let location: Location | null = null;
  renderComponent(<EditClientBtn clientId="client1" />, {
    setLocation: (newLocation) => {
      location = newLocation;
    },
  });
  await userEvent.click(
    screen.getByRole("button", { name: EditPanelButtonLabel.EDIT }),
  );
  expect((location as Location | null)?.search).toBe(
    "?panel=client-edit&id=client1",
  );
});
