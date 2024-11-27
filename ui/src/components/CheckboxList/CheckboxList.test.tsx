import { render } from "@testing-library/react";
import { screen } from "@testing-library/dom";
import userEvent from "@testing-library/user-event";

import CheckboxList from "./CheckboxList";

test("displays a label", () => {
  const label = "This is a list";
  render(
    <CheckboxList
      label={label}
      values={[]}
      checkedValues={[]}
      toggleValue={vi.fn()}
    />,
  );
  expect(screen.getByText(label)).toBeInTheDocument();
});

test("displays checkboxes", () => {
  render(
    <CheckboxList
      label="This is a list"
      values={["val1", "val2"]}
      checkedValues={[]}
      toggleValue={vi.fn()}
    />,
  );
  expect(screen.getByRole("checkbox", { name: "val1" })).toBeInTheDocument();
  expect(screen.getByRole("checkbox", { name: "val2" })).toBeInTheDocument();
});

test("checks selected checkboxes", () => {
  render(
    <CheckboxList
      label="This is a list"
      values={["val1", "val2"]}
      checkedValues={["val2"]}
      toggleValue={vi.fn()}
    />,
  );
  expect(screen.getByRole("checkbox", { name: "val1" })).not.toBeChecked();
  expect(screen.getByRole("checkbox", { name: "val2" })).toBeChecked();
});

test("changes checkbox state", async () => {
  const toggleValue = vi.fn();
  render(
    <CheckboxList
      label="This is a list"
      values={["val1", "val2"]}
      checkedValues={[]}
      toggleValue={toggleValue}
    />,
  );
  await userEvent.click(screen.getByRole("checkbox", { name: "val2" }));
  expect(toggleValue).toHaveBeenCalledWith("val2");
});
