import "@testing-library/jest-dom/vitest";
import type { Window as HappyDOMWindow } from "happy-dom";

declare global {
  // eslint-disable-next-line @typescript-eslint/no-empty-object-type
  interface Window extends HappyDOMWindow {}
}
