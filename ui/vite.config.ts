import { defineConfig, loadEnv } from "vite";
import react from "@vitejs/plugin-react";
import tsconfigPaths from "vite-tsconfig-paths";
import { createHtmlPlugin } from "vite-plugin-html";

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), "");
  return {
    css: {
      preprocessorOptions: {
        scss: {
          quietDeps: true,
          silenceDeprecations: ["import", "global-builtin", "mixed-decls"],
        },
      },
    },
    plugins: [tsconfigPaths(), react(), createHtmlPlugin()],
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
