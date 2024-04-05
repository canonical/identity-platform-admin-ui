import { FC } from "react";
import { NavLink } from "react-router-dom";
import { Button, Icon } from "@canonical/react-components";
import classnames from "classnames";
import Logo from "components/Logo";
import { GroupsLink, RolesLink } from "@canonical/rebac-admin";

const Navigation: FC = () => {
  return (
    <>
      <header className="l-navigation-bar">
        <div className="p-panel is-dark">
          <div className="p-panel__header">
            <Logo />
            <div className="p-panel__controls">
              <Button dense className="p-panel__toggle">
                Menu
              </Button>
            </div>
          </div>
        </div>
      </header>
      <nav aria-label="main navigation" className={classnames("l-navigation")}>
        <div className="l-navigation__drawer">
          <div className="p-panel is-dark">
            <div className="p-panel__header is-sticky">
              <Logo />
            </div>
            <div className="p-panel__content">
              <div className="p-side-navigation--icons is-dark">
                <ul className="p-side-navigation__list sidenav-top-ul">
                  <li className="p-side-navigation__item secondary">
                    <NavLink
                      className="p-side-navigation__link"
                      to={`/provider`}
                      title={`Provider list`}
                    >
                      <Icon
                        className="is-light p-side-navigation__icon"
                        name="plans"
                      />{" "}
                      Identity providers
                    </NavLink>
                  </li>
                  <li className="p-side-navigation__item secondary">
                    <NavLink
                      className="p-side-navigation__link"
                      to={`/client`}
                      title={`Client list`}
                    >
                      <Icon
                        className="is-light p-side-navigation__icon"
                        name="applications"
                      />{" "}
                      Clients
                    </NavLink>
                  </li>
                  <li className="p-side-navigation__item secondary">
                    <NavLink
                      className="p-side-navigation__link"
                      to={`/identity`}
                      title={`Identity list`}
                    >
                      <Icon
                        className="is-light p-side-navigation__icon"
                        name="user"
                      />{" "}
                      Identities
                    </NavLink>
                  </li>
                  <li className="p-side-navigation__item secondary">
                    <NavLink
                      className="p-side-navigation__link"
                      to={`/schema`}
                      title={`Schema list`}
                    >
                      <Icon
                        className="is-light p-side-navigation__icon"
                        name="profile"
                      />{" "}
                      Schemas
                    </NavLink>
                  </li>
                  <li className="p-side-navigation__item secondary">
                    <GroupsLink
                      className="p-side-navigation__link"
                      baseURL="/"
                      icon="user-group"
                      iconIsLight
                    />
                  </li>
                  <li className="p-side-navigation__item secondary">
                    <RolesLink
                      className="p-side-navigation__link"
                      baseURL="/"
                      icon="profile"
                      iconIsLight
                    />
                  </li>
                </ul>
              </div>
            </div>
          </div>
        </div>
      </nav>
    </>
  );
};

export default Navigation;
