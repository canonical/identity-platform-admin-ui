import React, { FC } from "react";
import { Col, MainTable, Row } from "@canonical/react-components";
import { useQuery } from "@tanstack/react-query";
import { queryKeys } from "util/queryKeys";
import { NotificationConsumer } from "@canonical/react-components/dist/components/NotificationProvider/NotificationProvider";
import { fetchSchemas } from "api/schema";

const SchemaList: FC = () => {
  const { data: schemas = [] } = useQuery({
    queryKey: [queryKeys.schemas],
    queryFn: fetchSchemas,
  });

  return (
    <div className="p-panel">
      <div className="p-panel__header ">
        <div className="p-panel__title">
          <h1 className="p-heading--4 u-no-margin--bottom">Schemas</h1>
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
                { content: "Schema" },
              ]}
              rows={schemas.map((schema) => {
                return {
                  columns: [
                    {
                      content: schema.id,
                      role: "rowheader",
                      "aria-label": "Id",
                    },
                    {
                      content: JSON.stringify(schema.schema),
                      role: "rowheader",
                      "aria-label": "Name",
                    },
                  ],
                  sortData: {
                    id: schema.id,
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

export default SchemaList;
