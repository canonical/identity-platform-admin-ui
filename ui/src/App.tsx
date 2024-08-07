import { FC } from "react";
import { Navigate, Route, Routes } from "react-router-dom";
import { ReBACAdmin } from "@canonical/rebac-admin";
import ClientList from "pages/clients/ClientList";
import NoMatch from "components/NoMatch";
import ProviderList from "pages/providers/ProviderList";
import IdentityList from "pages/identities/IdentityList";
import SchemaList from "pages/schemas/SchemaList";
import { apiBasePath } from "util/basePaths";
import Layout from "components/Layout/Layout";

const App: FC = () => {
  return (
    <Routes>
      <Route path="/" element={<Layout />}>
        <Route path="/" element={<Navigate to="/provider" replace={true} />} />
        <Route path="/provider" element={<ProviderList />} />
        <Route path="/client" element={<ClientList />} />
        <Route path="/identity" element={<IdentityList />} />
        <Route path="/schema" element={<SchemaList />} />
        <Route
          path="/*"
          element={
            <ReBACAdmin apiURL={apiBasePath} asidePanelId="app-layout" />
          }
        />
        <Route path="*" element={<NoMatch />} />
      </Route>
    </Routes>
  );
};

export default App;
