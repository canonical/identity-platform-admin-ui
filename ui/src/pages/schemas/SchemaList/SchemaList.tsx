import React, { FC } from "react";
import {
  Col,
  MainTable,
  Row,
  NotificationConsumer,
} from "@canonical/react-components";
import { useQuery } from "@tanstack/react-query";
import { queryKeys } from "util/queryKeys";
import { fetchSchemas } from "api/schema";
import Loader from "components/Loader";
import Pagination from "components/Pagination";
import { usePagination } from "util/usePagination";
import { Label } from "./types";

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
              headers={[
                { content: Label.HEADER_ID },
                { content: Label.HEADER_SCHEMA },
              ]}
              rows={response?.data.map((schema) => {
                return {
                  columns: [
                    {
                      content: schema.id,
                      role: "rowheader",
                    },
                    {
                      content: JSON.stringify(schema.schema),
                    },
                  ],
                };
              })}
              emptyStateMsg={
                isLoading ? <Loader text="Loading schemas..." /> : Label.NO_DATA
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
