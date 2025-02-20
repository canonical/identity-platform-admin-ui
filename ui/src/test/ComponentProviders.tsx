import { QueryClient } from "@tanstack/react-query";
import { QueryClientProvider } from "@tanstack/react-query";
import { useEffect, type PropsWithChildren } from "react";
import type { Location, RouteObject } from "react-router";
import { createBrowserRouter, RouterProvider, useLocation } from "react-router";

export type ComponentProps = {
  path: string;
  routeChildren?: RouteObject[];
  queryClient?: QueryClient;
  setLocation?: (location: Location) => void;
} & PropsWithChildren;

const Wrapper = ({
  children,
  setLocation,
}: PropsWithChildren & { setLocation?: (location: Location) => void }) => {
  const location = useLocation();
  useEffect(() => {
    setLocation?.(location);
  }, [location, setLocation]);
  return children;
};

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
  setLocation,
}: ComponentProps) => {
  const router = createBrowserRouter([
    {
      path,
      element: <Wrapper setLocation={setLocation}>{children}</Wrapper>,
      children: routeChildren,
    },
    {
      // Capture other paths to prevent warnings when navigating in tests.
      path: "*",
      element: <Wrapper setLocation={setLocation} />,
    },
  ]);
  return (
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  );
};

export default ComponentProviders;
