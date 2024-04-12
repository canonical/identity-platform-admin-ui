import React, { FC } from "react";
import { Col, MainTable, Row } from "@canonical/react-components";
import { useQuery } from "@tanstack/react-query";
import { queryKeys } from "util/queryKeys";
import { NotificationConsumer } from "@canonical/react-components/dist/components/NotificationProvider/NotificationProvider";
import { fetchSchemas } from "api/schema";
import Loader from "components/Loader";
import Pagination from "components/Pagination";
import { usePagination } from "util/usePagination";

const SchemaList: FC = () => {
  const { pageToken } = usePagination();

  const { data: response, isLoading } = useQuery({
    queryKey: [queryKeys.schemas, pageToken],
    queryFn: () => fetchSchemas(pageToken),
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
              responsive
              headers={[{ content: "Id" }, { content: "Schema" }]}
              rows={response?.data.map((schema) => {
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
                };
              })}
              emptyStateMsg={
                isLoading ? (
                  <Loader text="Loading schemas..." />
                ) : (
                  "No data to display"
                )
              }
            />
            <Pagination response={response} />
          </Col>
        </Row>
      </div>
    </div>
  );
};

export default SchemaList;
