import "@testing-library/jest-dom/vitest";
import type { Window as HappyDOMWindow } from "happy-dom";

declare global {
  interface Window extends HappyDOMWindow {}
}
