import { FC, useState } from "react";
import { useNavigate } from "react-router-dom";
import { queryKeys } from "util/queryKeys";
import { useQueryClient } from "@tanstack/react-query";
import {
  ActionButton,
  Button,
  Icon,
  Input,
  Modal,
  useNotify,
} from "@canonical/react-components";
import { deleteProvider } from "api/provider";
import { IdentityProvider } from "types/provider";
import usePortal from "react-useportal";

interface Props {
  provider: IdentityProvider;
}

const DeleteProviderBtn: FC<Props> = ({ provider }) => {
  const notify = useNotify();
  const queryClient = useQueryClient();
  const [isLoading, setLoading] = useState(false);
  const navigate = useNavigate();
  const { openPortal, closePortal, isOpen, Portal } = usePortal();
  const [confirmText, setConfirmText] = useState("");

  const handleDelete = () => {
    setLoading(true);
    deleteProvider(provider.id)
      .then(() => {
        navigate(
          "/provider",
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

  const expectedConfirmText = `remove ${provider.id}`;

  return (
    <>
      {isOpen && (
        <Portal>
          <Modal
            close={closePortal}
            title="Remove ID provider"
            buttonRow={
              <>
                <Button
                  className="u-no-margin--bottom"
                  type="button"
                  onClick={closePortal}
                >
                  Cancel
                </Button>
                <ActionButton
                  appearance="negative"
                  className="u-no-margin--bottom has-icon"
                  disabled={confirmText !== expectedConfirmText}
                  loading={isLoading}
                  onClick={handleDelete}
                >
                  <Icon name="delete" light />
                  <span>Remove</span>
                </ActionButton>
              </>
            }
          >
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
          </Modal>
        </Portal>
      )}
      <Button
        className="u-no-margin--bottom"
        onClick={openPortal}
        title="Confirm delete"
      >
        Delete
      </Button>
    </>
  );
};

export default DeleteProviderBtn;
