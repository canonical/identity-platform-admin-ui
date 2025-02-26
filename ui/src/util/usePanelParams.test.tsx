import { act, screen, waitFor } from "@testing-library/react";
import { Location } from "react-router";

import { renderComponent, renderWrappedHook } from "test/utils";

import usePanelParams from "./usePanelParams";
import {
  NotificationProvider,
  NotificationConsumer,
  useNotify,
} from "@canonical/react-components";
import userEvent from "@testing-library/user-event";
import { useEffect } from "react";

test("fetches the current panel", () => {
  const { result } = renderWrappedHook(() => usePanelParams(), {
    url: "/?panel=testpanel",
  });
  expect(result.current.panel).toBe("testpanel");
});

test("fetches the current id", () => {
  const { result } = renderWrappedHook(() => usePanelParams(), {
    url: "/?id=testid",
  });
  expect(result.current.id).toBe("testid");
});

test("sets the panel and args", () => {
  let location: Location | null = null;
  const { result } = renderWrappedHook(() => usePanelParams(), {
    url: "/",
    setLocation: (newLocation) => {
      location = newLocation;
    },
  });
  act(() => result.current.openProviderEdit("testid"));
  expect((location as Location | null)?.search).toBe(
    "?panel=provider-edit&id=testid",
  );
});

test("dispatches a resize event", async () => {
  const listener = vi.fn();
  addEventListener("resize", listener);
  const { result } = renderWrappedHook(() => usePanelParams(), {
    url: "/",
  });
  act(() => result.current.openProviderCreate());
  await waitFor(() => expect(listener).toHaveBeenCalled());
  removeEventListener("resize", listener);
});

test("clears notifications", async () => {
  const Test = () => {
    const panelParams = usePanelParams();
    const notify = useNotify();
    useEffect(() => {
      notify.queue(notify.success("A notification"));
    }, []);
    return <button onClick={panelParams.openClientCreate}>Test</button>;
  };
  renderComponent(
    <NotificationProvider>
      <Test />
      <NotificationConsumer />
    </NotificationProvider>,
  );
  expect(
    document.querySelector(".p-notification--positive"),
  ).toBeInTheDocument();
  await userEvent.click(screen.getByRole("button", { name: "Test" }));
  expect(
    document.querySelector(".p-notification--positive"),
  ).not.toBeInTheDocument();
});

test("clears the panel", async () => {
  let location: Location | null = null;
  const listener = vi.fn();
  addEventListener("resize", listener);
  const { result } = renderWrappedHook(() => usePanelParams(), {
    url: "/?panel=testpanel",
    setLocation: (newLocation) => {
      location = newLocation;
    },
  });
  act(() => result.current.clear());
  await waitFor(() => expect(listener).toHaveBeenCalled());
  expect((location as Location | null)?.search).toBe("");
  removeEventListener("resize", listener);
});

test("can open the create provider panel", () => {
  let location: Location | null = null;
  const { result } = renderWrappedHook(() => usePanelParams(), {
    url: "/?panel=testpanel",
    setLocation: (newLocation) => {
      location = newLocation;
    },
  });
  act(() => result.current.openProviderCreate());
  expect((location as Location | null)?.search).toBe("?panel=provider-create");
});

test("can open the edit provider panel", () => {
  let location: Location | null = null;
  const { result } = renderWrappedHook(() => usePanelParams(), {
    url: "/?search=query",
    setLocation: (newLocation) => {
      location = newLocation;
    },
  });
  act(() => result.current.openProviderEdit("testid"));
  expect((location as Location | null)?.search).toBe(
    "?panel=provider-edit&id=testid",
  );
});

test("can open the create client panel", () => {
  let location: Location | null = null;
  const { result } = renderWrappedHook(() => usePanelParams(), {
    url: "/?search=query",
    setLocation: (newLocation) => {
      location = newLocation;
    },
  });
  act(() => result.current.openClientCreate());
  expect((location as Location | null)?.search).toBe("?panel=client-create");
});

test("can open the edit client panel", () => {
  let location: Location | null = null;
  const { result } = renderWrappedHook(() => usePanelParams(), {
    url: "/?search=query",
    setLocation: (newLocation) => {
      location = newLocation;
    },
  });
  act(() => result.current.openClientEdit("testid"));
  expect((location as Location | null)?.search).toBe(
    "?panel=client-edit&id=testid",
  );
});

test("can open the create identity panel", () => {
  let location: Location | null = null;
  const { result } = renderWrappedHook(() => usePanelParams(), {
    url: "/?search=query",
    setLocation: (newLocation) => {
      location = newLocation;
    },
  });
  act(() => result.current.openIdentityCreate());
  expect((location as Location | null)?.search).toBe("?panel=identity-create");
});

test("can add panel params", () => {
  let location: Location | null = null;
  const { result } = renderWrappedHook(() => usePanelParams(), {
    url: "/?panel=testpanel",
    setLocation: (newLocation) => {
      location = newLocation;
    },
  });
  act(() => result.current.updatePanelParams("testkey", "testvalue"));
  expect((location as Location | null)?.search).toBe(
    "?panel=testpanel&testkey=testvalue",
  );
});

test("can update panel params", () => {
  let location: Location | null = null;
  const { result } = renderWrappedHook(() => usePanelParams(), {
    url: "/?panel=testpanel&testkey=oldvalue",
    setLocation: (newLocation) => {
      location = newLocation;
    },
  });
  act(() => result.current.updatePanelParams("testkey", "newvalue"));
  expect((location as Location | null)?.search).toBe(
    "?panel=testpanel&testkey=newvalue",
  );
});

test("removes empty panel params", () => {
  let location: Location | null = null;
  const { result } = renderWrappedHook(() => usePanelParams(), {
    url: "/?panel=testpanel&testkey=oldvalue",
    setLocation: (newLocation) => {
      location = newLocation;
    },
  });
  act(() => result.current.updatePanelParams("testkey", ""));
  expect((location as Location | null)?.search).toBe("?panel=testpanel");
});
