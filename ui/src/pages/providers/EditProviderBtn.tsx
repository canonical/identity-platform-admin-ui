import React, { FC } from "react";
import { Button, Icon } from "@canonical/react-components";
import { useNavigate } from "react-router-dom";

interface Props {
  providerId: string;
}

const EditProviderBtn: FC<Props> = ({ providerId }) => {
  const navigate = useNavigate();

  return (
    <Button
      appearance=""
      hasIcon
      onClick={() => navigate(`/provider/edit/${providerId}`)}
    >
      <Icon name="edit" />
      <span>Edit</span>
    </Button>
  );
};

export default EditProviderBtn;
