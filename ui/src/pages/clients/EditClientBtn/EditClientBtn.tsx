import { FC } from "react";

import EditPanelButton from "components/EditPanelButton";
import { testId } from "test/utils";

import { TestId } from "./test-types";

interface Props {
  clientId: string;
}

const EditClientBtn: FC<Props> = ({ clientId }) => {
  return (
    <EditPanelButton
      openPanel={(panelParams) => panelParams.openClientEdit(clientId)}
      {...testId(TestId.COMPONENT)}
    />
  );
};

export default EditClientBtn;
