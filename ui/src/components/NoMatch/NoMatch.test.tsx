import { screen } from "@testing-library/dom";
import { render } from "@testing-library/react";

import NoMatch from "./NoMatch";
import { Label } from "./types";

test("displays a no-match message", () => {
  render(<NoMatch />);
  expect(
    screen.getByRole("heading", { name: Label.TITLE }),
  ).toBeInTheDocument();
});
