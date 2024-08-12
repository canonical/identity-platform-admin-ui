import { SideNavigation } from "@canonical/react-components";
import { FC } from "react";
import { NavLink, NavLinkProps } from "react-router-dom";
import { Button } from "@canonical/react-components";
import { GroupsLink, RolesLink } from "@canonical/rebac-admin";
import { authURLs } from "api/auth";
import { appendAPIBasePath } from "util/basePaths";

type Props = {
  username?: string;
};

const Navigation: FC<Props> = ({ username }) => {
  return (
    <>
      <SideNavigation<NavLinkProps>
        dark={true}
        items={[
          {
            items: [
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
            ],
          },
        ]}
        linkComponent={NavLink}
      />
      <SideNavigation
        className="p-side-navigation--user-menu"
        dark={true}
        items={[
          {
            items: [
              {
                icon: "user",
                label: username,
                nonInteractive: true,
              },
              <Button
                element="a"
                appearance="link"
                href={appendAPIBasePath(authURLs.logout)}
                className="p-side-navigation__link"
                key="logout"
              >
                Logout
              </Button>,
            ],
          },
        ]}
        linkComponent={NavLink}
      />
    </>
  );
};

export default Navigation;
