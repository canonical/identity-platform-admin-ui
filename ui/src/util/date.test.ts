import { isoTimeToString } from "./date";

test("outputs local time", () => {
  expect(isoTimeToString("2024-12-02T03:16:20Z")).toBe("Dec 2, 2024, 03:16 AM");
});
