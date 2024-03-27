import { FC } from "react";
import { Button, Icon } from "@canonical/react-components";
import usePanelParams from "util/usePanelParams";

interface Props {
  providerId: string;
}

const EditProviderBtn: FC<Props> = ({ providerId }) => {
  const panelParams = usePanelParams();

  return (
    <Button
      className="u-no-margin--bottom"
      hasIcon
      onClick={() => panelParams.openProviderEdit(providerId)}
    >
      <Icon name="edit" />
      <span>Edit</span>
    </Button>
  );
};

export default EditProviderBtn;
