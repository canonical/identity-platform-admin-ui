import { SITE_NAME } from "consts";
import { FC } from "react";
import { NavLink } from "react-router";

const Logo: FC = () => {
  return (
    <NavLink className="p-panel__logo" to="/">
      <div className="p-navigation__logo-tag">
        <img
          className="p-navigation__logo-icon"
          src="https://assets.ubuntu.com/v1/82818827-CoF_white.svg"
          alt="Circle of friends"
        />
      </div>
      <div className="logo-text p-heading--4">{SITE_NAME}</div>
    </NavLink>
  );
};

export default Logo;
