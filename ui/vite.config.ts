import { defineConfig, loadEnv } from "vite";
import react from "@vitejs/plugin-react";
import tsconfigPaths from "vite-tsconfig-paths";

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), "");
  return {
    plugins: [tsconfigPaths(), react()],
    server: {
      port: 3000,
      proxy: {
        "/api": {
          target: env.VITE_PROXY_API_URL ?? "/",
        },
      },
    },
    build: {
      outDir: "./dist",
      minify: "esbuild",
    },
  };
});
