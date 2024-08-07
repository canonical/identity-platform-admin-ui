import { mergeConfig, defineConfig } from "vitest/config";
import viteConfig from "./vite.config";

export default defineConfig((configEnv) =>
  mergeConfig(
    viteConfig({ ...configEnv, mode: "development" }),
    defineConfig({
      test: {
        environment: "happy-dom",
        globals: true,
        include: ["./src/**/*.{test,spec}.{ts,tsx}"],
        mockReset: true,
        setupFiles: "src/test/setup.ts",
      },
    }),
  ),
);
