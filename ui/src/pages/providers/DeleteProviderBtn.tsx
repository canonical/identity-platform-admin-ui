import React, { FC, useState } from "react";
import { useNavigate } from "react-router-dom";
import { queryKeys } from "util/queryKeys";
import { useQueryClient } from "@tanstack/react-query";
import { ConfirmationButton, useNotify } from "@canonical/react-components";
import { deleteProvider } from "api/provider";
import { IdentityProvider } from "types/provider";

interface Props {
  provider: IdentityProvider;
}

const DeleteProviderBtn: FC<Props> = ({ provider }) => {
  const notify = useNotify();
  const queryClient = useQueryClient();
  const [isLoading, setLoading] = useState(false);
  const navigate = useNavigate();

  const handleDelete = () => {
    setLoading(true);
    deleteProvider(provider.id)
      .then(() => {
        navigate(
          "/provider/list",
          notify.queue(notify.success(`Provider ${provider.id} deleted.`)),
        );
      })
      .catch((e) => {
        notify.failure("Provider deletion failed", e);
      })
      .finally(() => {
        setLoading(false);
        void queryClient.invalidateQueries({
          queryKey: [queryKeys.providers],
        });
      });
  };

  return (
    <ConfirmationButton
      loading={isLoading}
      confirmationModalProps={{
        title: "Confirm delete",
        children: (
          <p>
            This will permanently delete provider <b>{provider.id}</b>.
          </p>
        ),
        confirmButtonLabel: "Delete provider",
        onConfirm: handleDelete,
      }}
      title="Confirm delete"
      appearance="base"
    >
      Delete
    </ConfirmationButton>
  );
};

export default DeleteProviderBtn;
