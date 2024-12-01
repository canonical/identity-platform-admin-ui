import { SideNavigation } from "@canonical/react-components";
import { FC } from "react";
import { NavLink, NavLinkProps } from "react-router-dom";
import { Button } from "@canonical/react-components";
import { GroupsLink, RolesLink } from "@canonical/rebac-admin-admin-ui";
import { authURLs } from "api/auth";
import { appendAPIBasePath } from "util/basePaths";
import { urls } from "urls";
import { Label } from "./types";

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
                label: Label.IDENTITY_PROVIDERS,
                to: urls.providers.index,
              },
              {
                icon: "applications",
                label: Label.CLIENTS,
                to: urls.clients.index,
              },
              {
                icon: "user",
                label: Label.IDENTITIES,
                to: urls.identities.index,
              },
              {
                icon: "profile",
                label: Label.SCHEMAS,
                to: urls.schemas.index,
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
                {Label.LOGOUT}
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
