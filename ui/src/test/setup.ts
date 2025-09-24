import "@testing-library/jest-dom/vitest";
import type { Window as HappyDOMWindow } from "happy-dom";

declare global {
  // eslint-disable-next-line @typescript-eslint/no-empty-object-type
  interface Window extends HappyDOMWindow {}
  var jest: object;
}
// Fix for RTL using fake timers:
// https://github.com/testing-library/user-event/issues/1115#issuecomment-1565730917
globalThis.jest = {
  advanceTimersByTime: vi.advanceTimersByTime.bind(vi),
};
