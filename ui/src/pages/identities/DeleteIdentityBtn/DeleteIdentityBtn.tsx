import { FC, useState } from "react";
import { useNavigate } from "react-router-dom";
import { queryKeys } from "util/queryKeys";
import { useQueryClient } from "@tanstack/react-query";
import { ConfirmationButton, useNotify } from "@canonical/react-components";
import { Identity } from "types/identity";
import { deleteIdentity } from "api/identities";

interface Props {
  identity: Identity;
}

const DeleteIdentityBtn: FC<Props> = ({ identity }) => {
  const notify = useNotify();
  const queryClient = useQueryClient();
  const [isLoading, setLoading] = useState(false);
  const navigate = useNavigate();

  const handleDelete = () => {
    setLoading(true);
    deleteIdentity(identity.id)
      .then(() => {
        navigate(
          "/identity",
          notify.queue(
            notify.success(`Identity ${identity.traits?.email} deleted.`),
          ),
        );
      })
      .catch((e) => {
        notify.failure("Identity deletion failed", e);
      })
      .finally(() => {
        setLoading(false);
        void queryClient.invalidateQueries({
          queryKey: [queryKeys.identities],
        });
      });
  };

  return (
    <ConfirmationButton
      className="u-no-margin--bottom"
      loading={isLoading}
      confirmationModalProps={{
        title: "Confirm delete",
        children: (
          <p>
            This will permanently delete identity{" "}
            <b>{identity.traits?.email}</b>.
          </p>
        ),
        confirmButtonLabel: "Delete identity",
        onConfirm: handleDelete,
      }}
      title="Confirm delete"
    >
      Delete
    </ConfirmationButton>
  );
};

export default DeleteIdentityBtn;
