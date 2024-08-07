import {
  Button,
  ButtonAppearance,
  CodeSnippet,
  LoginPageLayout,
  Spinner,
} from "@canonical/react-components";
import { authURLs } from "api/auth";
import { FC, ReactNode } from "react";
import { SITE_NAME } from "consts";
import { Label } from "./types";
import { appendAPIBasePath } from "util/basePaths";

type Props = {
  isLoading?: boolean;
  error?: string;
};

const Login: FC<Props> = ({ error, isLoading }) => {
  let loginContent: ReactNode;
  if (isLoading) {
    loginContent = <Spinner />;
  } else if (error) {
    loginContent = (
      <CodeSnippet
        blocks={[
          {
            code: error,
            wrapLines: true,
          },
        ]}
      />
    );
  } else {
    loginContent = (
      <Button
        appearance={
          isLoading ? ButtonAppearance.DEFAULT : ButtonAppearance.POSITIVE
        }
        disabled={isLoading}
        element="a"
        href={appendAPIBasePath(authURLs.login)}
      >
        Sign in to {SITE_NAME}
      </Button>
    );
  }
  return (
    <LoginPageLayout
      logo={{
        src: "https://assets.ubuntu.com/v1/82818827-CoF_white.svg",
        title: SITE_NAME,
        url: "/",
      }}
      title={Label.TITLE}
    >
      {loginContent}
    </LoginPageLayout>
  );
};

export default Login;
