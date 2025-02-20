import { useSearchParams } from "react-router";
import { useNotify } from "@canonical/react-components";

export interface PanelHelper {
  panel: string | null;
  id: string | null;
  clear: () => void;
  openProviderCreate: () => void;
  openProviderEdit: (id: string) => void;
  openClientCreate: () => void;
  openClientEdit: (id: string) => void;
  openIdentityCreate: () => void;
  updatePanelParams: (key: string, value: string) => void;
}

export const panels = {
  providerCreate: "provider-create",
  providerEdit: "provider-edit",
  clientCreate: "client-create",
  clientEdit: "client-edit",
  identityCreate: "identity-create",
};

type ParamMap = Record<string, string>;

const usePanelParams = (): PanelHelper => {
  const [params, setParams] = useSearchParams();
  const notify = useNotify();

  const craftResizeEvent = () => {
    setTimeout(() => window.dispatchEvent(new Event("resize")), 100);
  };

  const setPanelParams = (panel: string, args: ParamMap = {}) => {
    const newParams = new URLSearchParams();
    newParams.set("panel", panel);
    for (const [key, value] of Object.entries(args)) {
      newParams.set(key, value);
    }
    setParams(newParams);
    craftResizeEvent();
    notify.clear();
  };

  const clearParams = () => {
    setParams(new URLSearchParams());
    craftResizeEvent();
  };

  return {
    panel: params.get("panel"),
    id: params.get("id"),

    clear: () => {
      clearParams();
    },

    openProviderCreate: () => {
      setPanelParams(panels.providerCreate);
    },

    openProviderEdit: (id: string) => {
      setPanelParams(panels.providerEdit, { id });
    },

    openClientCreate: () => {
      setPanelParams(panels.clientCreate);
    },

    openClientEdit: (id: string) => {
      setPanelParams(panels.clientEdit, { id });
    },

    openIdentityCreate: () => {
      setPanelParams(panels.identityCreate);
    },

    updatePanelParams: (key: string, value: string) => {
      const newParams = new URLSearchParams(params);
      newParams.set(key, value);
      if (!value) {
        newParams.delete(key);
      }
      setParams(newParams);
    },
  };
};

export default usePanelParams;
