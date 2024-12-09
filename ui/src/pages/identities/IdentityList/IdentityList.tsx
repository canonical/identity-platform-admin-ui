import React, { FC } from "react";
import {
  Button,
  Col,
  MainTable,
  Row,
  NotificationConsumer,
} from "@canonical/react-components";
import { useQuery } from "@tanstack/react-query";
import { queryKeys } from "util/queryKeys";
import { fetchIdentities } from "api/identities";
import { isoTimeToString } from "util/date";
import Loader from "components/Loader";
import usePanelParams from "util/usePanelParams";
import { usePagination } from "util/usePagination";
import Pagination from "components/Pagination";
import DeleteIdentityBtn from "pages/identities/DeleteIdentityBtn";
import { Label } from "./types";

const IdentityList: FC = () => {
  const panelParams = usePanelParams();
  const { pageToken } = usePagination();

  const { data: response, isLoading } = useQuery({
    queryKey: [queryKeys.identities, pageToken],
    queryFn: () => fetchIdentities(pageToken),
  });

  return (
    <div className="p-panel">
      <div className="p-panel__header ">
        <div className="p-panel__title">
          <h1 className="p-heading--4 u-no-margin--bottom">Identities</h1>
        </div>
        <div className="p-panel__controls">
          <Button
            appearance="positive"
            onClick={panelParams.openIdentityCreate}
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
              responsive
              headers={[
                { content: Label.HEADER_ID },
                { content: Label.HEADER_SCHEMA },
                { content: Label.HEADER_CREATED_AT },
                { content: Label.HEADER_ACTIONS },
              ]}
              rows={response?.data.map((identity) => {
                return {
                  columns: [
                    {
                      content: identity.traits?.email ?? identity.id,
                      role: "rowheader",
                    },
                    {
                      content: identity.schema_id,
                    },
                    {
                      content: identity.created_at
                        ? isoTimeToString(identity.created_at)
                        : "",
                    },
                    {
                      content: (
                        <>
                          <DeleteIdentityBtn identity={identity} />
                        </>
                      ),
                    },
                  ],
                };
              })}
              emptyStateMsg={
                isLoading ? (
                  <Loader text="Loading identities..." />
                ) : (
                  Label.NO_DATA
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

export default IdentityList;
