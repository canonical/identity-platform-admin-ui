import React, { FC } from "react";
import { Button, Col, MainTable, Row } from "@canonical/react-components";
import { useQuery } from "@tanstack/react-query";
import { queryKeys } from "util/queryKeys";
import { Link, useNavigate } from "react-router-dom";
import { NotificationConsumer } from "@canonical/react-components/dist/components/NotificationProvider/NotificationProvider";
import { fetchProviders } from "api/provider";

const ProviderList: FC = () => {
  const navigate = useNavigate();

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
            onClick={() => navigate("/provider/create")}
          >
            Create provider
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
                { content: "Id", sortKey: "id" },
                { content: "Provider", sortKey: "provider" },
                { content: "Scope", sortKey: "scope" },
              ]}
              rows={providers.map((provider) => {
                return {
                  columns: [
                    {
                      content: (
                        <Link to={`/provider/detail/${provider.id}`}>
                          {provider.id}
                        </Link>
                      ),
                      role: "rowheader",
                      "aria-label": "Id",
                    },
                    {
                      content: provider.provider,
                      role: "rowheader",
                      "aria-label": "Name",
                    },
                    {
                      content: provider.scope.length,
                      role: "rowheader",
                      "aria-label": "Response types",
                    },
                  ],
                  sortData: {
                    id: provider.id,
                    provider: provider.provider,
                    scope: provider.scope.length,
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
