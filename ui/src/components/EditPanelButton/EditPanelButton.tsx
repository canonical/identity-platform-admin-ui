import { FC } from "react";
import { Button, Icon } from "@canonical/react-components";

import usePanelParams from "util/usePanelParams";

import { Label, Props } from "./types";

const EditPanelButton: FC<Props> = ({ openPanel, ...props }: Props) => {
  const panelParams = usePanelParams();

  return (
    <Button
      className="u-no-margin--bottom"
      hasIcon
      onClick={() => openPanel(panelParams)}
      {...props}
    >
      <Icon name="edit" />
      <span>{Label.EDIT}</span>
    </Button>
  );
};

export default EditPanelButton;
