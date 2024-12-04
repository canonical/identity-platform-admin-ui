import { FC } from "react";
import { Button, Icon } from "@canonical/react-components";

import usePanelParams from "util/usePanelParams";

import { Label, Props } from "./types";

const EditPanelButton: FC<Props> = ({ openPanel }: Props) => {
  const panelParams = usePanelParams();

  return (
    <Button
      className="u-no-margin--bottom"
      hasIcon
      onClick={() => openPanel(panelParams)}
    >
      <Icon name="edit" />
      <span>{Label.EDIT}</span>
    </Button>
  );
};

export default EditPanelButton;
