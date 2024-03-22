import React, { FC } from "react";
import {
  Button,
  Form,
  Input,
  Select,
  Textarea,
} from "@canonical/react-components";
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
  subject_source?: string;
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
  const [advanced, setAdvanced] = React.useState(false);

  return (
    <Form onSubmit={formik.handleSubmit}>
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
          <Input
            {...formik.getFieldProps("microsoft_tenant")}
            id="microsoft_tenant"
            type="text"
            label="Tenant"
            error={
              formik.touched.microsoft_tenant
                ? formik.errors.microsoft_tenant
                : null
            }
          />
          <Input
            {...formik.getFieldProps("subject_source")}
            id="subject_source"
            type="text"
            label="Subject source"
            error={
              formik.touched.subject_source
                ? formik.errors.subject_source
                : null
            }
          />
        </>
      )}
      {formik.values.provider === "apple" && (
        <>
          <Textarea
            {...formik.getFieldProps("apple_private_key")}
            id="apple_private_key"
            label="Apple Private Key"
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
            label="Apple Team ID"
            error={
              formik.touched.apple_team_id ? formik.errors.apple_team_id : null
            }
          />
          <Input
            {...formik.getFieldProps("apple_private_key_id")}
            id="apple_private_key_id"
            type="text"
            label="Apple Private Key ID"
            error={
              formik.touched.apple_private_key_id
                ? formik.errors.apple_private_key_id
                : null
            }
          />
        </>
      )}
      <Button
        appearance="base"
        type="button"
        hasIcon
        onClick={() => {
          setAdvanced(!advanced);
        }}
        aria-label="Toggle advanced view"
        className="p-accordion__tab"
        aria-expanded={advanced ? "true" : "false"}
      >
        <strong className="p-heading--5 p-inline-list__item">Advanced</strong>
      </Button>
      {advanced && (
        <>
          {!["apple"].includes(formik.values.provider ?? "") && (
            <>
              <Input
                {...formik.getFieldProps("auth_url")}
                id="auth_url"
                type="text"
                label="Auth URL"
                error={formik.touched.auth_url ? formik.errors.auth_url : null}
              />
              <Input
                {...formik.getFieldProps("token_url")}
                id="token_url"
                type="text"
                label="Token URL"
                error={
                  formik.touched.token_url ? formik.errors.token_url : null
                }
              />
              <Input
                {...formik.getFieldProps("issuer_url")}
                id="issuer_url"
                type="text"
                label="Issuer URL"
                error={
                  formik.touched.issuer_url ? formik.errors.issuer_url : null
                }
              />
            </>
          )}
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
            error={formik.touched.mapper_url ? formik.errors.mapper_url : null}
            disabled={isEdit}
          />
          <Input
            {...formik.getFieldProps("scope")}
            id="scope"
            type="text"
            label="Scope"
            help="Scope specifies optional requested permissions"
            error={formik.touched.scope ? formik.errors.scope : null}
          />
        </>
      )}
    </Form>
  );
};

export default ProviderForm;
