import { screen } from "@testing-library/dom";
import { render } from "@testing-library/react";

import ScrollableContainer from "./ScrollableContainer";
import { useState } from "react";
import userEvent from "@testing-library/user-event";

test("displays children", () => {
  const content = "Content";
  render(
    <ScrollableContainer dependencies={[]}>{content}</ScrollableContainer>,
  );
  expect(screen.getByText(content)).toBeInTheDocument();
});

test("sets the height of the content wrapper", () => {
  render(
    <>
      <ScrollableContainer belowId="below" dependencies={[]}>
        Content
      </ScrollableContainer>
      <div id="below">Below</div>
    </>,
  );
  expect(document.querySelector(".content-details")).toHaveAttribute(
    "style",
    "height: calc(100vh - 2px); min-height: calc(100vh - 2px)",
  );
});

test("adjust the height if the dependencies change", async () => {
  const Test = () => {
    const [count, setCount] = useState(0);
    return (
      <>
        <ScrollableContainer belowId="below" dependencies={[count]}>
          <button onClick={() => setCount(count + 1)}>Update</button>
        </ScrollableContainer>
        <div id="below">Below</div>
      </>
    );
  };
  render(<Test />);
  expect(document.querySelector(".content-details")).toHaveAttribute(
    "style",
    "height: calc(100vh - 2px); min-height: calc(100vh - 2px)",
  );
  // Adjust the DOM so that there is a size change to test for once the dependencies change.
  Object.defineProperty(document.getElementById("below"), "offsetHeight", {
    configurable: true,
    value: 50,
  });
  await userEvent.click(screen.getByRole("button", { name: "Update" }));
  expect(document.querySelector(".content-details")).toHaveAttribute(
    "style",
    "height: calc(100vh - 52px); min-height: calc(100vh - 52px)",
  );
});

test("adjust the height if the window is resized", () => {
  const Test = () => (
    <>
      <ScrollableContainer belowId="below" dependencies={[]}>
        Content
      </ScrollableContainer>
      <div id="below">Below</div>
    </>
  );
  render(<Test />);
  expect(document.querySelector(".content-details")).toHaveAttribute(
    "style",
    "height: calc(100vh - 2px); min-height: calc(100vh - 2px)",
  );
  // Adjust the DOM so that there is a size change to test for once the resize
  // event occurs.
  Object.defineProperty(document.getElementById("below"), "offsetHeight", {
    configurable: true,
    value: 50,
  });
  // Resize the window.
  window.happyDOM.setViewport({ height: 100, width: 100 });
  expect(document.querySelector(".content-details")).toHaveAttribute(
    "style",
    "height: calc(100vh - 52px); min-height: calc(100vh - 52px)",
  );
});
