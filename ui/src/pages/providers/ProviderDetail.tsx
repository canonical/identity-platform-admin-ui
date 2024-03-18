import React, { FC } from "react";
import { Row } from "@canonical/react-components";
import { Link, useParams } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { queryKeys } from "util/queryKeys";
import { NotificationConsumer } from "@canonical/react-components/dist/components/NotificationProvider/NotificationProvider";
import { fetchProvider } from "api/provider";
import DeleteProviderBtn from "pages/providers/DeleteProviderBtn";
import EditProviderBtn from "pages/providers/EditProviderBtn";

const ProviderDetail: FC = () => {
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
          <nav
            key="breadcrumbs"
            className="p-breadcrumbs"
            aria-label="Breadcrumbs"
          >
            <ol className="p-breadcrumbs__items">
              <li className="p-breadcrumbs__item">
                <Link to="/provider/list">Providers</Link>
              </li>
              <li className="p-breadcrumbs__item">Details</li>
            </ol>
          </nav>
        </div>
        {providerId && (
          <div className="p-panel__controls">
            {provider && <DeleteProviderBtn provider={provider} />}
            <EditProviderBtn providerId={providerId} />
          </div>
        )}
      </div>
      <div className="p-panel__content">
        <Row>
          <h1 className="p-heading--4">Provider {providerId}</h1>
          <NotificationConsumer />
          <h2 className="p-heading--5">raw data:</h2>
          <code>{JSON.stringify(provider)}</code>
        </Row>
      </div>
    </div>
  );
};

export default ProviderDetail;
