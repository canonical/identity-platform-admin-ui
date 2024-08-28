import { QueryClient } from "@tanstack/react-query";
import { render, renderHook } from "@testing-library/react";

import ComponentProviders from "./ComponentProviders";
import type { ComponentProps } from "./ComponentProviders";

type Options = {
  url?: string;
  path?: string;
  routeChildren?: ComponentProps["routeChildren"];
  queryClient?: QueryClient;
  setLocation?: ComponentProps["setLocation"];
};

export const changeURL = (url: string) => window.happyDOM.setURL(url);
const getQueryClient = (options: Options | null | undefined) =>
  options?.queryClient
    ? options.queryClient
    : new QueryClient({
        defaultOptions: {
          queries: {
            retry: false,
            staleTime: 1,
          },
        },
      });

export const renderComponent = (
  component: JSX.Element,
  options?: Options | null,
) => {
  const queryClient = getQueryClient(options);
  changeURL(options?.url ?? "/");
  const result = render(component, {
    wrapper: (props) => (
      <ComponentProviders
        {...props}
        routeChildren={options?.routeChildren}
        path={options?.path ?? "*"}
        queryClient={queryClient}
        setLocation={options?.setLocation}
      />
    ),
  });
  return { changeURL, result, queryClient };
};

export const renderWrappedHook = <Result, Props>(
  hook: (initialProps: Props) => Result,
  options?: Options | null,
) => {
  const queryClient = getQueryClient(options);
  changeURL(options?.url ?? "/");
  const { result } = renderHook(hook, {
    wrapper: (props) => (
      <ComponentProviders
        {...props}
        routeChildren={options?.routeChildren}
        path={options?.path ?? "*"}
        queryClient={queryClient}
        setLocation={options?.setLocation}
      />
    ),
  });
  return { changeURL, result, queryClient };
};
