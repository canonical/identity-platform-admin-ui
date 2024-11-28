import { getParentsBottomSpacing, getAbsoluteHeightBelow } from "./helpers";

test("getParentsBottomSpacing", () => {
  const child = document.createElement("div");
  const parent1 = document.createElement("div");
  const parent2 = document.createElement("div");
  parent1.setAttribute("style", "margin-bottom: 20px; padding-bottom: 10px;");
  parent1.appendChild(child);
  parent2.setAttribute("style", "margin-bottom: 40px; padding-bottom: 30px;");
  parent2.appendChild(parent1);
  document.body.appendChild(parent2);
  expect(getParentsBottomSpacing(child)).toBe(100);
  document.body.removeChild(parent2);
});

test("getAbsoluteHeightBelow", () => {
  const below = document.createElement("div");
  below.setAttribute("id", "below");
  below.setAttribute(
    "style",
    "margin-top: 10px; margin-bottom: 20px; padding-top: 30px; padding-bottom: 40px;",
  );
  Object.defineProperty(below, "offsetHeight", {
    configurable: true,
    value: 50,
  });
  document.body.appendChild(below);
  expect(getAbsoluteHeightBelow("below")).toBe(151);
  document.body.removeChild(below);
});
