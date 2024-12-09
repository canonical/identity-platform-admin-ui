import { FC, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useQueryClient } from "@tanstack/react-query";
import { ConfirmationButton, useNotify } from "@canonical/react-components";
import { Label, Props } from "./types";

const DeletePanelButton: FC<Props> = ({
  confirmButtonDisabled,
  confirmButtonLabel,
  confirmContent,
  confirmTitle = "Confirm delete",
  invalidateQuery,
  entityName,
  onDelete,
  successMessage,
  successPath,
  ...props
}) => {
  const notify = useNotify();
  const queryClient = useQueryClient();
  const [isLoading, setLoading] = useState(false);
  const navigate = useNavigate();

  const handleDelete = () => {
    setLoading(true);
    onDelete()
      .then(() => {
        navigate(successPath, notify.queue(notify.success(successMessage)));
      })
      .catch((error: unknown) => {
        notify.failure(
          `${entityName} deletion failed`,
          error instanceof Error ? error : null,
          typeof error === "string" ? error : null,
        );
      })
      .finally(() => {
        setLoading(false);
        void queryClient.invalidateQueries({
          queryKey: [invalidateQuery],
        });
      });
  };

  return (
    <ConfirmationButton
      className="u-no-margin--bottom"
      loading={isLoading}
      confirmationModalProps={{
        children: confirmContent,
        confirmButtonDisabled,
        confirmButtonLabel,
        onConfirm: handleDelete,
        title: confirmTitle,
      }}
      title={confirmTitle}
      {...props}
    >
      {Label.DELETE}
    </ConfirmationButton>
  );
};

export default DeletePanelButton;
