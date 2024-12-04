import { screen } from "@testing-library/dom";
import userEvent from "@testing-library/user-event";

import { renderComponent } from "test/utils";

import EditPanelButton from "./EditPanelButton";
import { Label } from "./types";

test("displays a button", () => {
  renderComponent(<EditPanelButton openPanel={vi.fn()} />);
  expect(screen.getByRole("button", { name: Label.EDIT })).toBeInTheDocument();
});

test("can open a panel", async () => {
  const openPanel = vi.fn();
  renderComponent(<EditPanelButton openPanel={openPanel} />);
  await userEvent.click(screen.getByRole("button", { name: Label.EDIT }));
  expect(openPanel).toHaveBeenCalled();
});
