import { FC, Suspense } from "react";
import { Navigate, Route, Routes } from "react-router-dom";
import { ReBACAdmin } from "@canonical/rebac-admin";
import Loader from "components/Loader";
import Login from "components/Login";
import ClientList from "pages/clients/ClientList";
import NoMatch from "components/NoMatch";
import ProviderList from "pages/providers/ProviderList";
import IdentityList from "pages/identities/IdentityList";
import SchemaList from "pages/schemas/SchemaList";
import Navigation from "components/Navigation";
import Panels from "components/Panels";
import useLocalStorage from "util/useLocalStorage";
import { apiBasePath, basePath } from "util/basePaths";

const App: FC = () => {
  // Store a user token that will be passed to the API using the
  // X-Authorization header so that the user can be identified. This will be
  // replaced by API authentication when it has been implemented.
  const [authUser, setAuthUser] = useLocalStorage<{
    username: string;
    token: string;
  } | null>("user", null);
  return (
    <div className="l-application" role="presentation">
      <Navigation
        username={authUser?.username}
        logout={() => {
          setAuthUser(null);
          window.location.reload();
        }}
      />
      <main className="l-main">
        <Suspense fallback={<Loader />}>
          <Routes>
            <Route
              path={basePath}
              element={
                <Login isAuthenticated={!!authUser} setAuthUser={setAuthUser} />
              }
            >
              <Route
                path={basePath}
                element={<Navigate to={`${basePath}provider`} replace={true} />}
              />
              <Route path={`${basePath}provider`} element={<ProviderList />} />
              <Route path={`${basePath}client`} element={<ClientList />} />
              <Route path={`${basePath}identity`} element={<IdentityList />} />
              <Route path={`${basePath}schema`} element={<SchemaList />} />
              <Route
                path={basePath + "*"}
                element={
                  <ReBACAdmin
                    apiURL={apiBasePath}
                    asidePanelId="rebac-admin-panel"
                    authToken={authUser?.token}
                  />
                }
              />
              <Route path="*" element={<NoMatch />} />
            </Route>
          </Routes>
        </Suspense>
      </main>
      <Panels />
    </div>
  );
};

export default App;
