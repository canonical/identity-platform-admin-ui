import { FC } from "react";
import { NavLink } from "react-router-dom";
import { Button } from "@canonical/react-components";
import { GroupsLink, RolesLink } from "@canonical/rebac-admin";
import SideNavigation from "@canonical/react-components/dist/components/SideNavigation";

type Props = {
  username?: string;
  logout: () => void;
};

const Navigation: FC<Props> = ({ username, logout }) => {
  return (
    <>
      <SideNavigation
        dark={true}
        items={[
          {
            icon: "plans",
            label: "Identity providers",
            to: "/provider",
          },
          {
            icon: "applications",
            label: "Clients",
            to: "/client",
          },
          {
            icon: "user",
            label: "Identities",
            to: "/identity",
          },
          {
            icon: "profile",
            label: "Schemas",
            to: "/schema",
          },
          <GroupsLink
            className="p-side-navigation__link"
            baseURL="/"
            icon="user-group"
            iconIsLight
            key="groups"
          />,
          <RolesLink
            className="p-side-navigation__link"
            baseURL="/"
            icon="profile"
            iconIsLight
            key="roles"
          />,
        ]}
        linkComponent={NavLink}
      />
      <SideNavigation
        className="p-side-navigation--user-menu"
        dark={true}
        items={[
          {
            icon: "user",
            label: username,
            nonInteractive: true,
          },
          <Button
            appearance="link"
            className="p-side-navigation__link"
            onClick={() => {
              logout();
            }}
            key="logout"
          >
            Logout
          </Button>,
        ]}
        linkComponent={NavLink}
      />
    </>
  );
};

export default Navigation;
