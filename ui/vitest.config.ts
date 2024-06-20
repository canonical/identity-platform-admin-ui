import { mergeConfig, defineConfig } from "vitest/config";
import viteConfig from "./vite.config";

export default mergeConfig(
  viteConfig({ mode: "development" }),
  defineConfig({
    test: {
      environment: "jsdom",
      globals: true,
      include: ["./src/**/*.spec.{ts,tsx}"],
    },
  }),
);
