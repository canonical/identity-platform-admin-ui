import { FC, Suspense } from "react";
import { Outlet } from "react-router-dom";
import { ApplicationLayout } from "@canonical/react-components";
import Loader from "components/Loader";
import Login from "components/Login";
import Logo from "components/Logo";
import Navigation from "components/Navigation";
import Panels from "components/Panels";
import { useQuery } from "@tanstack/react-query";
import { queryKeys } from "util/queryKeys";
import { fetchMe } from "api/auth";

const Layout: FC = () => {
  const {
    data: user,
    isLoading,
    error,
  } = useQuery({
    queryKey: [queryKeys.auth],
    queryFn: fetchMe,
  });
  return user ? (
    <ApplicationLayout
      aside={<Panels />}
      id="app-layout"
      logo={<Logo />}
      sideNavigation={<Navigation username={user.name || user.email} />}
    >
      <Suspense fallback={<Loader />}>
        <Outlet />
      </Suspense>
    </ApplicationLayout>
  ) : (
    <Login isLoading={isLoading} error={error?.message}></Login>
  );
};

export default Layout;
