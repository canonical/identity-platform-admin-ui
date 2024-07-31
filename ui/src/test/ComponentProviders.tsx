import { QueryClient } from "@tanstack/react-query";
import { QueryClientProvider } from "@tanstack/react-query";
import { type PropsWithChildren } from "react";
import type { RouteObject } from "react-router-dom";
import { createBrowserRouter, RouterProvider } from "react-router-dom";

export type ComponentProps = {
  path: string;
  routeChildren?: RouteObject[];
  queryClient?: QueryClient;
} & PropsWithChildren;

const ComponentProviders = ({
  children,
  routeChildren,
  path,
  queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  }),
}: ComponentProps) => {
  const router = createBrowserRouter([
    {
      path,
      element: children,
      children: routeChildren,
    },
    {
      // Capture other paths to prevent warnings when navigating in tests.
      path: "*",
      element: <span />,
    },
  ]);
  return (
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  );
};

export default ComponentProviders;
