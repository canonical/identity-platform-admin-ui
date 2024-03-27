import { FC, useState } from "react";
import { Form, Input, Select, Textarea } from "@canonical/react-components";
import { FormikProps } from "formik";

const providerOptions = [
  {
    label: "Generic OAuth2 / OIDC provider",
    value: "generic",
  },
  {
    label: "Apple",
    value: "apple",
  },
  {
    label: "Auth0",
    value: "auth0",
  },
  {
    label: "AWS",
    value: "aws",
  },
  {
    label: "Dingtalk",
    value: "dingtalk",
  },
  {
    label: "Discord",
    value: "discord",
  },
  {
    label: "Facebook",
    value: "facebook",
  },
  {
    label: "GitHub",
    value: "github",
  },
  {
    label: "GitHub-App",
    value: "github-app",
  },
  {
    label: "Gitlab",
    value: "gitlab",
  },
  {
    label: "Google",
    value: "google",
  },
  {
    label: "Linkedin",
    value: "linkedin",
  },
  {
    label: "Microsoft",
    value: "microsoft",
  },
  {
    label: "Netid",
    value: "netid",
  },
  {
    label: "Okta",
    value: "okta",
  },
  {
    label: "Patreon",
    value: "patreon",
  },
  {
    label: "Slack",
    value: "slack",
  },
  {
    label: "Spotify",
    value: "spotify",
  },
  {
    label: "VK",
    value: "vk",
  },
  {
    label: "Yandex",
    value: "yandex",
  },
];

export interface ProviderFormTypes {
  id?: string;
  client_id?: string;
  client_secret?: string;
  auth_url?: string;
  issuer_url?: string;
  token_url?: string;
  subject_source?: "userinfo" | "me";
  microsoft_tenant?: string;
  provider?: string;
  mapper_url?: string;
  scope?: string;
  apple_team_id?: string;
  apple_private_key_id?: string;
  apple_private_key?: string;
  requested_claims?: string;
}

interface Props {
  formik: FormikProps<ProviderFormTypes>;
  isEdit?: boolean;
}

const ProviderForm: FC<Props> = ({ formik, isEdit = false }) => {
  const [hasAutoDiscovery, setAutoDiscovery] = useState(
    !(formik.values.auth_url || formik.values.token_url),
  );

  return (
    <Form onSubmit={formik.handleSubmit}>
      <h2 className="p-heading--5 u-no-margin--bottom">General</h2>
      <Select
        {...formik.getFieldProps("provider")}
        id="provider"
        options={providerOptions}
        label="Provider"
        error={formik.touched.provider ? formik.errors.provider : null}
      />
      <Input
        {...formik.getFieldProps("id")}
        id="id"
        type="text"
        label="Name"
        error={formik.touched.id ? formik.errors.id : null}
        disabled={isEdit}
      />
      <Input
        {...formik.getFieldProps("client_id")}
        id="client_id"
        type="text"
        label="Client ID"
        error={formik.touched.client_id ? formik.errors.client_id : null}
        disabled={isEdit}
      />
      {!["apple"].includes(formik.values.provider ?? "") && (
        <Input
          {...formik.getFieldProps("client_secret")}
          id="client_secret"
          type="text"
          label="Client secret"
          error={
            formik.touched.client_secret ? formik.errors.client_secret : null
          }
        />
      )}
      {formik.values.provider === "microsoft" && (
        <>
          <h2 className="p-heading--5 u-no-margin--bottom">
            Microsoft configuration
          </h2>
          <Input
            {...formik.getFieldProps("microsoft_tenant")}
            id="microsoft_tenant"
            type="text"
            label="Tenant"
            help={
              <>
                The Azure AD Tenant to use for authentication. Can either be{" "}
                <code>common</code>, <code>organizations</code>,
                <code>consumers</code> for a multitenant application or a
                specific tenant like <code>contoso.onmicrosoft.com</code>.
              </>
            }
            error={
              formik.touched.microsoft_tenant
                ? formik.errors.microsoft_tenant
                : null
            }
          />
          <p className="u-no-margin--bottom">Subject source</p>
          <Input
            name="subject_source"
            type="radio"
            label="Userinfo"
            help="The subject identifier is taken from sub field of userifo standard endpoint response"
            checked={formik.values.subject_source === "userinfo"}
            onChange={() =>
              void formik.setFieldValue("subject_source", "userinfo")
            }
          />
          <Input
            name="subject_source"
            type="radio"
            label="Me"
            help={
              <>
                The <code>id</code> field of https://graph.microsoft.com/v1.0/me
                response is taken as subject
              </>
            }
            checked={formik.values.subject_source === "me"}
            onChange={() => void formik.setFieldValue("subject_source", "me")}
          />
        </>
      )}
      {formik.values.provider === "apple" && (
        <>
          <h2 className="p-heading--5 u-no-margin--bottom">
            Apple configuration
          </h2>
          <Textarea
            {...formik.getFieldProps("apple_private_key")}
            id="apple_private_key"
            label="Private Key"
            error={
              formik.touched.apple_private_key
                ? formik.errors.apple_private_key
                : null
            }
          />
          <Input
            {...formik.getFieldProps("apple_team_id")}
            id="apple_team_id"
            type="text"
            label="Developer Team ID"
            help={
              <>
                The Apple Developer Team ID can be found at{" "}
                <a
                  href="https://developer.apple.com/"
                  target="_blank"
                  rel="noreferrer"
                >
                  developer.apple.com
                </a>
              </>
            }
            error={
              formik.touched.apple_team_id ? formik.errors.apple_team_id : null
            }
          />
          <Input
            {...formik.getFieldProps("apple_private_key_id")}
            id="apple_private_key_id"
            type="text"
            label="Private Key ID"
            error={
              formik.touched.apple_private_key_id
                ? formik.errors.apple_private_key_id
                : null
            }
          />
        </>
      )}
      {formik.values.provider === "generic" && (
        <>
          <h2 className="p-heading--5 u-no-margin--bottom">Urls</h2>
          <p className="u-no-margin--bottom">
            Does the OIDC server support OIDC Discovery?
          </p>
          <p className="u-text--muted u-no-margin--bottom">
            You can read more about OIDC Discovery in the{" "}
            <a
              href="https://openid.net/specs/openid-connect-discovery-1_0.html#IssuerDiscovery"
              target="_blank"
              rel="noreferrer"
            >
              OIDC specs
            </a>
            .
          </p>
          <Input
            type="radio"
            id="discovery_on"
            label="Yes"
            name="oidc_discovery"
            checked={hasAutoDiscovery}
            onChange={() => setAutoDiscovery(true)}
          />
          {hasAutoDiscovery && (
            <div className="radio-section">
              <Input
                {...formik.getFieldProps("issuer_url")}
                id="issuer_url"
                type="text"
                label="OIDC server URL"
                error={
                  formik.touched.issuer_url ? formik.errors.issuer_url : null
                }
              />
            </div>
          )}
          <Input
            type="radio"
            id="discovery_off"
            label="No"
            name="oidc_discovery"
            checked={!hasAutoDiscovery}
            onChange={() => setAutoDiscovery(false)}
          />
          {!hasAutoDiscovery && (
            <div className="radio-section">
              <Input
                {...formik.getFieldProps("auth_url")}
                id="auth_url"
                type="text"
                label="Auth URL"
                help="I.e. https://example.org/oauth2/auth"
                error={formik.touched.auth_url ? formik.errors.auth_url : null}
              />
              <Input
                {...formik.getFieldProps("token_url")}
                id="token_url"
                type="text"
                label="Token URL"
                help="I.e. https://example.org/oauth2/token"
                error={
                  formik.touched.token_url ? formik.errors.token_url : null
                }
              />
            </div>
          )}
        </>
      )}
      <h2 className="p-heading--5 u-no-margin--bottom">
        Optional configurations
      </h2>
      <Input
        {...formik.getFieldProps("scope")}
        id="scope"
        type="text"
        label="Scopes"
        help={
          <>
            Comma seperated list of optional requested permissions. Common
            values are <code>email</code>, <code>openid</code>,{" "}
            <code>profile</code>, <code>address</code>, <code>phone</code> or
            custom values.
          </>
        }
        error={formik.touched.scope ? formik.errors.scope : null}
      />
      <Input
        {...formik.getFieldProps("requested_claims")}
        id="requested_claims"
        type="text"
        label="Requested claims"
        error={
          formik.touched.requested_claims
            ? formik.errors.requested_claims
            : null
        }
      />
      <Input
        {...formik.getFieldProps("mapper_url")}
        id="mapper_url"
        type="text"
        label="Mapper"
        help="Mapper specifies the JSONNet code snippet which uses the OpenID Connect Provider's data to hydrate the identity's data. Supported file types are .jsonnet"
        error={formik.touched.mapper_url ? formik.errors.mapper_url : null}
        disabled={isEdit}
      />
    </Form>
  );
};

export default ProviderForm;
