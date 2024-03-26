import { FC } from "react";
import { Button, Icon } from "@canonical/react-components";
import usePanelParams from "util/usePanelParams";

interface Props {
  clientId: string;
}

const EditClientBtn: FC<Props> = ({ clientId }) => {
  const panelParams = usePanelParams();

  return (
    <Button
      className="u-no-margin--bottom"
      hasIcon
      onClick={() => panelParams.openClientEdit(clientId)}
    >
      <Icon name="edit" />
      <span>Edit</span>
    </Button>
  );
};

export default EditClientBtn;
