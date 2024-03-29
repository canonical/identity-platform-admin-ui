import ProviderCreate from "pages/providers/ProviderCreate";
import usePanelParams, { panels } from "util/usePanelParams";
import ProviderEdit from "pages/providers/ProviderEdit";
import ClientCreate from "pages/clients/ClientCreate";
import ClientEdit from "pages/clients/ClientEdit";

const Panels = () => {
  const panelParams = usePanelParams();

  const generatePanel = () => {
    switch (panelParams.panel) {
      case panels.providerCreate:
        return <ProviderCreate />;
      case panels.providerEdit:
        return <ProviderEdit />;
      case panels.clientCreate:
        return <ClientCreate />;
      case panels.clientEdit:
        return <ClientEdit />;
      default:
        return null;
    }
  };
  return <>{panelParams.panel && generatePanel()}</>;
};

export default Panels;
