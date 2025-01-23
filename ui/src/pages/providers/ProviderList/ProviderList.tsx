import { FC } from "react";
import {
  Button,
  Col,
  MainTable,
  NotificationConsumer,
  Row,
} from "@canonical/react-components";
import { useQuery } from "@tanstack/react-query";
import { queryKeys } from "util/queryKeys";
import { fetchProviders } from "api/provider";
import usePanelParams from "util/usePanelParams";
import EditProviderBtn from "pages/providers/EditProviderBtn";
import DeleteProviderBtn from "pages/providers/DeleteProviderBtn";
import { Label } from "./types";

const ProviderList: FC = () => {
  const panelParams = usePanelParams();

  const { data: providers } = useQuery({
    queryKey: [queryKeys.providers],
    queryFn: fetchProviders,
  });

  return (
    <div className="p-panel">
      <div className="p-panel__header ">
        <div className="p-panel__title">
          <h1 className="p-heading--4 u-no-margin--bottom">
            Identity Providers
          </h1>
        </div>
        <div className="p-panel__controls">
          <Button
            appearance="positive"
            onClick={panelParams.openProviderCreate}
          >
            {Label.ADD}
          </Button>
        </div>
      </div>
      <div className="p-panel__content">
        <Row>
          <Col size={12}>
            <NotificationConsumer />
            <MainTable
              className="u-table-layout--auto"
              sortable
              responsive
              headers={[
                { content: Label.HEADER_NAME, sortKey: "id" },
                { content: Label.HEADER_PROVIDER, sortKey: "provider" },
                { content: Label.HEADER_ACTIONS },
              ]}
              rows={providers?.data.map((provider) => {
                return {
                  columns: [
                    {
                      content: provider.id,
                      role: "rowheader",
                    },
                    {
                      content: provider.provider,
                    },
                    {
                      content: (
                        <>
                          <EditProviderBtn providerId={provider.id ?? ""} />
                          <DeleteProviderBtn provider={provider} />
                        </>
                      ),
                    },
                  ],
                  sortData: {
                    id: provider.id,
                    provider: provider.provider,
                  },
                };
              })}
            />
          </Col>
        </Row>
      </div>
    </div>
  );
};

export default ProviderList;
