import { FC } from "react";
import { NavLink } from "react-router-dom";
import { Button, Icon } from "@canonical/react-components";
import classnames from "classnames";
import Logo from "components/Logo";

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
