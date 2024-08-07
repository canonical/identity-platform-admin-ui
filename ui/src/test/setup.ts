import "@testing-library/jest-dom/vitest";
import { vi } from "vitest";
import createFetchMock from "vitest-fetch-mock";
import type { Window as HappyDOMWindow } from "happy-dom";

declare global {
  interface Window extends HappyDOMWindow {}
}

const fetchMocker = createFetchMock(vi);
// sets globalThis.fetch and globalThis.fetchMock to our mocked version
fetchMocker.enableMocks();
