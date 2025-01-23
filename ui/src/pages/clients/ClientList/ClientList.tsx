import { FC } from "react";
import {
  Button,
  Col,
  MainTable,
  Row,
  NotificationConsumer,
} from "@canonical/react-components";
import { useQuery } from "@tanstack/react-query";
import { queryKeys } from "util/queryKeys";
import { fetchClients } from "api/client";
import { isoTimeToString } from "util/date";
import usePanelParams from "util/usePanelParams";
import EditClientBtn from "pages/clients/EditClientBtn";
import DeleteClientBtn from "pages/clients/DeleteClientBtn";
import Loader from "components/Loader";
import Pagination from "components/Pagination";
import { usePagination } from "util/usePagination";
import { Label } from "./types";

const ClientList: FC = () => {
  const panelParams = usePanelParams();
  const { pageToken } = usePagination();

  const { data: response, isLoading } = useQuery({
    queryKey: [queryKeys.clients, pageToken],
    queryFn: () => fetchClients(pageToken),
  });

  return (
    <div className="p-panel">
      <div className="p-panel__header ">
        <div className="p-panel__title">
          <h1 className="p-heading--4 u-no-margin--bottom">Clients</h1>
        </div>
        <div className="p-panel__controls">
          <Button appearance="positive" onClick={panelParams.openClientCreate}>
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
                { content: Label.HEADER_NAME },
                { content: Label.HEADER_DATE },
                { content: Label.HEADER_ACTIONS },
              ]}
              rows={response?.data.map((client) => {
                return {
                  columns: [
                    {
                      content: client.client_id,
                      role: "rowheader",
                    },
                    {
                      content: client.client_name,
                    },
                    {
                      content: (
                        <>
                          <div>
                            Created: {isoTimeToString(client.created_at)}
                          </div>
                          <div className="u-text--muted">
                            Updated: {isoTimeToString(client.updated_at)}
                          </div>
                        </>
                      ),
                    },
                    {
                      content: (
                        <>
                          <EditClientBtn clientId={client.client_id} />
                          <DeleteClientBtn client={client} />
                        </>
                      ),
                    },
                  ],
                };
              })}
              emptyStateMsg={
                isLoading ? <Loader text="Loading clients..." /> : Label.NO_DATA
              }
            />
            <Pagination response={response} />
          </Col>
        </Row>
      </div>
    </div>
  );
};

export default ClientList;
