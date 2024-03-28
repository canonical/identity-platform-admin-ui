import React, { FC } from "react";
import { Col, MainTable, Row } from "@canonical/react-components";
import { useQuery } from "@tanstack/react-query";
import { queryKeys } from "util/queryKeys";
import { NotificationConsumer } from "@canonical/react-components/dist/components/NotificationProvider/NotificationProvider";
import { fetchIdentities } from "api/identities";
import { isoTimeToString } from "util/date";

const IdentityList: FC = () => {
  const { data: identities = [] } = useQuery({
    queryKey: [queryKeys.identities],
    queryFn: fetchIdentities,
  });

  return (
    <div className="p-panel">
      <div className="p-panel__header ">
        <div className="p-panel__title">
          <h1 className="p-heading--4 u-no-margin--bottom">Identities</h1>
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
                { content: "Schema", sortKey: "schema" },
                { content: "Created at", sortKey: "createdAt" },
              ]}
              rows={identities.map((identity) => {
                return {
                  columns: [
                    {
                      content: identity.traits?.email ?? identity.id,
                      role: "rowheader",
                      "aria-label": "Id",
                    },
                    {
                      content: identity.schema_id,
                      role: "rowheader",
                      "aria-label": "Schema",
                    },
                    {
                      content: identity.created_at
                        ? isoTimeToString(identity.created_at)
                        : "",
                      role: "rowheader",
                      "aria-label": "Created at",
                    },
                  ],
                  sortData: {
                    id: identity.id,
                    schema: identity.schema_id,
                    createdAt: identity.created_at,
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

export default IdentityList;
