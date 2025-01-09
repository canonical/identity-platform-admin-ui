import { FC } from "react";
import { queryKeys } from "util/queryKeys";
import { deleteClient } from "api/client";
import { Client } from "types/client";
import DeletePanelButton from "components/DeletePanelButton";
import { urls } from "urls";
import { Label } from "./types";
import { testId } from "test/utils";
import { TestId } from "./test-types";

interface Props {
  client: Client;
}

const DeleteClientBtn: FC<Props> = ({ client }) => {
  return (
    <DeletePanelButton
      confirmButtonLabel={Label.CONFIRM}
      confirmContent={
        <p>
          This will permanently delete client <b>{client.client_name}</b>.
        </p>
      }
      entityName="Client"
      invalidateQuery={queryKeys.clients}
      onDelete={() => deleteClient(client.client_id)}
      successPath={urls.clients.index}
      successMessage={`Client ${client.client_name} deleted.`}
      {...testId(TestId.COMPONENT)}
    />
  );
};

export default DeleteClientBtn;
