import { removeTrailingSlash } from "./removeTrailingSlash";

test("removes trailing slash", () => {
  expect(removeTrailingSlash("/trailing/")).toBe("/trailing");
});

test("handles no trailing slash", () => {
  expect(removeTrailingSlash("/no-trailing")).toBe("/no-trailing");
});
