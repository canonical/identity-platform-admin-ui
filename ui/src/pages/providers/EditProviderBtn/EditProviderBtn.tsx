import { FC } from "react";

import EditPanelButton from "components/EditPanelButton";

interface Props {
  providerId: string;
}

const EditProviderBtn: FC<Props> = ({ providerId }) => {
  return (
    <EditPanelButton
      openPanel={(panelParams) => panelParams.openProviderEdit(providerId)}
    />
  );
};

export default EditProviderBtn;
