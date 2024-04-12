import { FC } from "react";
import { Button, Col, MainTable, Row } from "@canonical/react-components";
import { useQuery } from "@tanstack/react-query";
import { queryKeys } from "util/queryKeys";
import { NotificationConsumer } from "@canonical/react-components/dist/components/NotificationProvider/NotificationProvider";
import { fetchClients } from "api/client";
import { isoTimeToString } from "util/date";
import usePanelParams from "util/usePanelParams";
import EditClientBtn from "pages/clients/EditClientBtn";
import DeleteClientBtn from "pages/clients/DeleteClientBtn";
import Loader from "components/Loader";
import Pagination from "components/Pagination";
import { usePagination } from "util/usePagination";

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
            Add client
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
                { content: "Id" },
                { content: "Name" },
                { content: "Date" },
                { content: "Actions" },
              ]}
              rows={response?.data.map((client) => {
                return {
                  columns: [
                    {
                      content: client.client_id,
                      role: "rowheader",
                      "aria-label": "Id",
                    },
                    {
                      content: client.client_name,
                      role: "rowheader",
                      "aria-label": "Name",
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
                      role: "rowheader",
                      "aria-label": "Created",
                    },
                    {
                      content: (
                        <>
                          <EditClientBtn clientId={client.client_id} />
                          <DeleteClientBtn client={client} />
                        </>
                      ),
                      role: "rowheader",
                      "aria-label": "Actions",
                    },
                  ],
                };
              })}
              emptyStateMsg={
                isLoading ? (
                  <Loader text="Loading clients..." />
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

export default ClientList;
