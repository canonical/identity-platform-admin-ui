import { FC } from "react";
import { Form, Input } from "@canonical/react-components";
import { FormikProps } from "formik";
import CheckboxList from "components/CheckboxList";
import { Label } from "./types";

export interface ClientFormTypes {
  client_uri?: string;
  client_name?: string;
  grant_types?: string[];
  response_types?: string[];
  scope?: string;
  redirect_uris?: string[];
  request_object_signing_alg?: string;
}

interface Props {
  formik: FormikProps<ClientFormTypes>;
}

const ClientForm: FC<Props> = ({ formik }) => {
  const toggle = (currentValues: string[], field: string, value: string) => {
    const newValues = currentValues.includes(value)
      ? currentValues.filter((c) => c !== value)
      : [...currentValues, value];

    void formik.setFieldValue(field, newValues);
  };

  return (
    <Form onSubmit={formik.handleSubmit}>
      <Input
        id="client_name"
        name="client_name"
        type="text"
        label={Label.NAME}
        onBlur={formik.handleBlur}
        onChange={formik.handleChange}
        value={formik.values.client_name}
        error={formik.touched.client_name ? formik.errors.client_name : null}
        required
      />
      <Input
        id="scope"
        name="scope"
        type="text"
        label={Label.SCOPE}
        onBlur={formik.handleBlur}
        onChange={formik.handleChange}
        value={formik.values.scope}
        error={formik.touched.scope ? formik.errors.scope : null}
        required
      />
      <Input
        id="redirect_uris"
        name="redirect_uris"
        type="text"
        label={Label.REDIRECT_URI}
        onBlur={formik.handleBlur}
        onChange={(e) =>
          void formik.setFieldValue("redirect_uris", [e.target.value])
        }
        value={formik.values.redirect_uris}
        error={
          formik.touched.redirect_uris ? formik.errors.redirect_uris : null
        }
        required
      />
      <Input
        id="client_uri"
        name="client_uri"
        type="text"
        label={Label.CLIENT_URI}
        onBlur={formik.handleBlur}
        onChange={formik.handleChange}
        value={formik.values.client_uri}
        error={formik.touched.client_uri ? formik.errors.client_uri : null}
        required
      />
      <Input
        id="request_object_signing_alg"
        name="request_object_signing_alg"
        type="text"
        label={Label.SIGNING_ALGORITHM}
        onBlur={formik.handleBlur}
        onChange={formik.handleChange}
        value={formik.values.request_object_signing_alg}
        error={
          formik.touched.request_object_signing_alg
            ? formik.errors.request_object_signing_alg
            : null
        }
        required
      />
      <CheckboxList
        label={Label.GRANT_TYPES}
        values={["authorization_code", "refresh_token"]}
        checkedValues={formik.values.grant_types ?? []}
        toggleValue={(value: string) =>
          toggle(formik.values.grant_types ?? [], "grant_types", value)
        }
      />
      <CheckboxList
        label={Label.RESPONSE_TYPES}
        values={["code", "id_token"]}
        checkedValues={formik.values.response_types ?? []}
        toggleValue={(value: string) =>
          toggle(formik.values.response_types ?? [], "response_types", value)
        }
      />
    </Form>
  );
};

export default ClientForm;
