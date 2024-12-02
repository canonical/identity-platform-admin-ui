import { FC } from "react";

import EditPanelButton from "components/EditPanelButton";

interface Props {
  clientId: string;
}

const EditClientBtn: FC<Props> = ({ clientId }) => {
  return (
    <EditPanelButton
      openPanel={(panelParams) => panelParams.openClientEdit(clientId)}
    />
  );
};

export default EditClientBtn;
