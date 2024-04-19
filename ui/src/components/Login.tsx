import { Button, FormikField } from "@canonical/react-components";
import { Form, Formik } from "formik";
import { FC } from "react";
import { Outlet } from "react-router-dom";
import * as Yup from "yup";
import { AuthUser } from "types/auth-user";

type Props = {
  isAuthenticated?: boolean;
  setAuthUser: (user: AuthUser) => void;
};

const schema = Yup.object().shape({
  username: Yup.string().required("Required"),
});

const Login: FC<Props> = ({ isAuthenticated, setAuthUser }) => {
  return isAuthenticated ? (
    <Outlet />
  ) : (
    <div className="p-login">
      <div className="p-login__inner p-card--highlighted">
        <div className="p-login__tagged-logo">
          <span className="p-login__logo-container">
            <div className="p-login__logo-tag">
              <img
                className="p-login__logo-icon"
                src="https://assets.ubuntu.com/v1/82818827-CoF_white.svg"
                alt=""
              />
            </div>
            <span className="p-login__logo-title">Identity platform</span>
          </span>
        </div>
        <Formik
          initialValues={{ username: "" }}
          onSubmit={({ username }) => {
            setAuthUser({ username, token: btoa(username) });
          }}
          validationSchema={schema}
        >
          <Form>
            <FormikField
              label="Username"
              name="username"
              takeFocus
              type="text"
            />
            <Button appearance="positive" type="submit">
              Log in
            </Button>
          </Form>
        </Formik>
      </div>
    </div>
  );
};

export default Login;
