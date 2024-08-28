import { useQueryClient } from "@tanstack/react-query";
import { FC, useRef } from "react";
import { Navigate, Route, Routes } from "react-router-dom";
import { ReBACAdmin } from "@canonical/rebac-admin";
import { AxiosError } from "axios";

import { authURLs } from "api/auth";
import ClientList from "pages/clients/ClientList";
import NoMatch from "components/NoMatch";
import ProviderList from "pages/providers/ProviderList";
import IdentityList from "pages/identities/IdentityList";
import SchemaList from "pages/schemas/SchemaList";
import Layout from "components/Layout/Layout";
import { queryKeys } from "util/queryKeys";
import { axiosInstance } from "./api/axios";
import { urls } from "urls";
import { useNext } from "util/useNext";

const App: FC = () => {
  const queryClient = useQueryClient();
  // Redirect to the ?next=/... URL returned by the authentication step.
  useNext();

  useRef(
    axiosInstance.interceptors.response.use(
      (response) => response,
      (error: AxiosError) => {
        if (
          error.response?.status === 401 &&
          // The 'me' endpoint is used to check whether the user is
          // authenticated, so don't invalidate the cache for these requests as
          // that would cause it to be fetched again and cause an infinite loop.
          error.config?.url !== authURLs.me
        ) {
          // Handle any unauthenticated requests and clear the cache for the
          // auth endpoints so that the user details will get requested again
          // and display the login screen if needed.
          void queryClient.invalidateQueries({
            queryKey: [queryKeys.auth],
          });
          return null;
        }
        return Promise.reject(error);
      },
    ),
  );

  return (
    <Routes>
      <Route path={urls.index} element={<Layout />}>
        <Route
          path={urls.index}
          element={<Navigate to={urls.providers.index} replace={true} />}
        />
        <Route path={urls.providers.index} element={<ProviderList />} />
        <Route path={urls.clients.index} element={<ClientList />} />
        <Route path={urls.identities.index} element={<IdentityList />} />
        <Route path={urls.schemas.index} element={<SchemaList />} />
        <Route
          path={`${urls.index}*`}
          element={
            <ReBACAdmin
              asidePanelId="app-layout"
              axiosInstance={axiosInstance}
            />
          }
        />
        <Route path="*" element={<NoMatch />} />
      </Route>
    </Routes>
  );
};

export default App;
