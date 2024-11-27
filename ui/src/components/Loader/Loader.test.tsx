import { render } from "@testing-library/react";
import { screen } from "@testing-library/dom";

import Loader from "./Loader";

test("displays a spinner", () => {
  render(<Loader />);
  const spinner = document.querySelector(".p-icon--spinner");
  expect(spinner).toBeInTheDocument();
  expect(screen.getByText("Loading...")).toBeInTheDocument();
});

test("displays custom loading text", () => {
  render(<Loader text="Spinning" />);
  const spinner = document.querySelector(".p-icon--spinner");
  expect(spinner).toBeInTheDocument();
  expect(screen.getByText("Spinning")).toBeInTheDocument();
});
