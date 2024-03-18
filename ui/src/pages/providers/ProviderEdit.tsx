import React, { FC } from "react";
import { queryKeys } from "util/queryKeys";
import { useQuery } from "@tanstack/react-query";
import { useParams } from "react-router-dom";
import { fetchProvider } from "api/provider";
import ProviderEditForm from "pages/providers/ProviderEditForm";

const ProviderEdit: FC = () => {
  const { providerId } = useParams<{ providerId: string }>();

  if (!providerId) {
    return <></>;
  }

  const { data: provider } = useQuery({
    queryKey: [queryKeys.providers, providerId],
    queryFn: () => fetchProvider(providerId),
  });

  return (
    <div className="p-panel">
      <div className="p-panel__header ">
        <div className="p-panel__title">
          <h1 className="p-heading--4 u-no-margin--bottom">Edit provider</h1>
        </div>
      </div>
      <div className="p-panel__content">
        {provider && <ProviderEditForm provider={provider} />}
      </div>
    </div>
  );
};

export default ProviderEdit;
