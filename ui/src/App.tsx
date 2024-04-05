import { FC, Suspense } from "react";
import { Navigate, Route, Routes } from "react-router-dom";
import Loader from "components/Loader";
import ClientList from "pages/clients/ClientList";
import NoMatch from "components/NoMatch";
import ProviderList from "pages/providers/ProviderList";
import IdentityList from "pages/identities/IdentityList";
import SchemaList from "pages/schemas/SchemaList";
import { ReBACAdmin } from "@canonical/rebac-admin";

const App: FC = () => {
  return (
    <Suspense fallback={<Loader />}>
      <Routes>
        <Route path="/" element={<Navigate to="/provider" replace={true} />} />
        <Route path="/provider" element={<ProviderList />} />
        <Route path="/client" element={<ClientList />} />
        <Route path="/identity" element={<IdentityList />} />
        <Route path="/schema" element={<SchemaList />} />
        <Route
          path="/*"
          element={<ReBACAdmin apiURL={import.meta.env.VITE_API_URL} />}
        />
        <Route path="*" element={<NoMatch />} />
      </Routes>
    </Suspense>
  );
};

export default App;
