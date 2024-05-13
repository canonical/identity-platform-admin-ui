import { FC, Suspense } from "react";
import { Navigate, Route, Routes } from "react-router-dom";
import { ApplicationLayout } from "@canonical/react-components";
import { ReBACAdmin } from "@canonical/rebac-admin";
import Loader from "components/Loader";
import Login from "components/Login";
import ClientList from "pages/clients/ClientList";
import NoMatch from "components/NoMatch";
import ProviderList from "pages/providers/ProviderList";
import IdentityList from "pages/identities/IdentityList";
import SchemaList from "pages/schemas/SchemaList";
import Panels from "components/Panels";
import useLocalStorage from "util/useLocalStorage";
import Logo from "components/Logo";
import Navigation from "components/Navigation";

const VITE_API_URL =
  typeof import.meta.env.VITE_API_URL === "string"
    ? import.meta.env.VITE_API_URL
    : "/";

const API_URL = (
  VITE_API_URL.startsWith("http") || VITE_API_URL.startsWith("/")
    ? VITE_API_URL
    : "/"
) as `${"http" | "/"}${string}`;

const App: FC = () => {
  // Store a user token that will be passed to the API using the
  // X-Authorization header so that the user can be identified. This will be
  // replaced by API authentication when it has been implemented.
  const [authUser, setAuthUser] = useLocalStorage<{
    username: string;
    token: string;
  } | null>("user", null);
  return (
    <ApplicationLayout
      aside={<Panels />}
      logo={<Logo />}
      sideNavigation={
        <Navigation
          username={authUser?.username}
          logout={() => {
            setAuthUser(null);
            window.location.reload();
          }}
        />
      }
    >
      <Suspense fallback={<Loader />}>
        <Routes>
          <Route
            path="/"
            element={
              <Login isAuthenticated={!!authUser} setAuthUser={setAuthUser} />
            }
          >
            <Route
              path="/"
              element={<Navigate to="/provider" replace={true} />}
            />
            <Route path="/provider" element={<ProviderList />} />
            <Route path="/client" element={<ClientList />} />
            <Route path="/identity" element={<IdentityList />} />
            <Route path="/schema" element={<SchemaList />} />
            <Route
              path="/*"
              element={
                <ReBACAdmin
                  apiURL={API_URL}
                  asidePanelId="rebac-admin-panel"
                  authToken={authUser?.token}
                />
              }
            />
            <Route path="*" element={<NoMatch />} />
          </Route>
        </Routes>
      </Suspense>
    </ApplicationLayout>
  );
};

export default App;
