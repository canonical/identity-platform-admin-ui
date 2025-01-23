import { FC } from "react";
import { queryKeys } from "util/queryKeys";
import { Identity } from "types/identity";
import { deleteIdentity } from "api/identities";
import { urls } from "urls";
import DeletePanelButton from "components/DeletePanelButton";
import { Label } from "./types";
import { testId } from "test/utils";
import { TestId } from "./test-types";

interface Props {
  identity: Identity;
}

const DeleteIdentityBtn: FC<Props> = ({ identity }) => {
  return (
    <DeletePanelButton
      confirmButtonLabel={Label.CONFIRM}
      confirmContent={
        <p>
          This will permanently delete identity <b>{identity.traits?.email}</b>.
        </p>
      }
      entityName="Identity"
      invalidateQuery={queryKeys.identities}
      onDelete={() => deleteIdentity(identity.id)}
      successPath={urls.identities.index}
      successMessage={`Identity ${identity.traits?.email} deleted.`}
      {...testId(TestId.COMPONENT)}
    />
  );
};

export default DeleteIdentityBtn;
