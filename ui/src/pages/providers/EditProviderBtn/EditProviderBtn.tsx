import { FC } from "react";

import EditPanelButton from "components/EditPanelButton";
import { testId } from "test/utils";

import { TestId } from "./test-types";

interface Props {
  providerId: string;
}

const EditProviderBtn: FC<Props> = ({ providerId }) => {
  return (
    <EditPanelButton
      openPanel={(panelParams) => panelParams.openProviderEdit(providerId)}
      {...testId(TestId.COMPONENT)}
    />
  );
};

export default EditProviderBtn;
