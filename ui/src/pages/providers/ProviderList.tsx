import React, { FC } from "react";
import { Button, Col, MainTable, Row } from "@canonical/react-components";
import { useQuery } from "@tanstack/react-query";
import { queryKeys } from "util/queryKeys";
import { NotificationConsumer } from "@canonical/react-components/dist/components/NotificationProvider/NotificationProvider";
import { fetchProviders } from "api/provider";
import usePanelParams from "util/usePanelParams";
import EditProviderBtn from "pages/providers/EditProviderBtn";
import DeleteProviderBtn from "pages/providers/DeleteProviderBtn";

const ProviderList: FC = () => {
  const panelParams = usePanelParams();

  const { data: providers = [] } = useQuery({
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
            Add ID provider
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
              paginate={30}
              headers={[
                { content: "Name", sortKey: "id" },
                { content: "Provider", sortKey: "provider" },
                { content: "Actions" },
              ]}
              rows={providers.map((provider) => {
                return {
                  columns: [
                    {
                      content: provider.id,
                      role: "rowheader",
                      "aria-label": "Name",
                    },
                    {
                      content: provider.provider,
                      role: "rowheader",
                      "aria-label": "Provider",
                    },
                    {
                      content: (
                        <>
                          <EditProviderBtn providerId={provider.id ?? ""} />
                          <DeleteProviderBtn provider={provider} />
                        </>
                      ),
                      role: "rowheader",
                      "aria-label": "Actions",
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
