import { createRoot } from "react-dom/client";
import { BrowserRouter as Router } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import App from "./App";
import "./sass/styles.scss";
import { NotificationProvider } from "@canonical/react-components";
import { basePath } from "util/basePaths";
import { removeTrailingSlash } from "util/removeTrailingSlash";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: false,
      // Cache queries for 30 seconds by default.
      staleTime: 30000,
    },
  },
});

const rootElement = document.getElementById("app");
if (!rootElement) throw new Error("Failed to find the root element");
const root = createRoot(rootElement);
root.render(
  <Router basename={removeTrailingSlash(basePath)}>
    <QueryClientProvider client={queryClient}>
      <NotificationProvider>
        <App />
      </NotificationProvider>
    </QueryClientProvider>
  </Router>,
);
