import { FC, useState } from "react";
import { queryKeys } from "util/queryKeys";
import { Input } from "@canonical/react-components";
import { deleteProvider } from "api/provider";
import { IdentityProvider } from "types/provider";
import { urls } from "urls";
import DeletePanelButton from "components/DeletePanelButton";
import { Label } from "./types";

interface Props {
  provider: IdentityProvider;
}

const DeleteProviderBtn: FC<Props> = ({ provider }) => {
  const [confirmText, setConfirmText] = useState("");
  const expectedConfirmText = `remove ${provider.id || ""}`;

  return (
    <DeletePanelButton
      confirmButtonDisabled={confirmText !== expectedConfirmText}
      confirmButtonLabel={Label.CONFIRM}
      confirmContent={
        <>
          <p>
            Are you sure you want to remove {'"'}
            {provider.id}
            {'"'} as an ID provider? The removal of {provider.id} as an ID
            provider is irreversible and might adversely affect your system.
          </p>
          <Input
            onChange={(e) => setConfirmText(e.target.value)}
            value={confirmText}
            type="text"
            placeholder={expectedConfirmText}
            label={
              <>
                Type <b>{expectedConfirmText}</b> to confirm
              </>
            }
          />
        </>
      }
      confirmTitle="Remove ID provider"
      entityName="Provider"
      invalidateQuery={queryKeys.providers}
      onDelete={() => {
        if (!provider.id) {
          const error = "Cannot delete provider without id";
          console.error(error, provider);
          return Promise.reject(error);
        }
        return deleteProvider(provider.id);
      }}
      successPath={urls.providers.index}
      successMessage={`Provider ${provider.id} deleted.`}
    />
  );
};

export default DeleteProviderBtn;
