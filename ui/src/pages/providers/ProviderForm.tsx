import React, { FC } from "react";
import { Form, Input, Select } from "@canonical/react-components";
import { FormikProps } from "formik";

export interface ProviderFormTypes {
  id?: string;
  client_id?: string;
  client_secret?: string;
  provider?: string;
  mapper?: string;
  scope?: string;
}

interface Props {
  formik: FormikProps<ProviderFormTypes>;
  isEdit?: boolean;
}

const ProviderForm: FC<Props> = ({ formik, isEdit = false }) => {
  return (
    <Form onSubmit={formik.handleSubmit}>
      <Input
        id="id"
        name="id"
        type="text"
        label="Name"
        onBlur={formik.handleBlur}
        onChange={formik.handleChange}
        value={formik.values.id}
        error={formik.touched.id ? formik.errors.id : null}
        disabled={isEdit}
        required
      />
      <Select
        id="provider"
        name="provider"
        options={[
          "apple",
          "auth0",
          "aws",
          "generic",
          "github",
          "google",
          "microsoft",
          "okta",
          "spotify",
        ].map((item) => {
          return { label: item, value: item };
        })}
        label="Provider"
        onBlur={formik.handleBlur}
        onChange={formik.handleChange}
        value={formik.values.provider}
        error={formik.touched.provider ? formik.errors.provider : null}
      />
      <Input
        id="scope"
        name="scope"
        type="text"
        label="Scope"
        help="Scope specifies optional requested permissions"
        onBlur={formik.handleBlur}
        onChange={formik.handleChange}
        value={formik.values.scope}
        error={formik.touched.scope ? formik.errors.scope : null}
      />
      <Input
        id="client_id"
        name="client_id"
        type="text"
        label="Client ID"
        onBlur={formik.handleBlur}
        onChange={formik.handleChange}
        value={formik.values.client_id}
        error={formik.touched.client_id ? formik.errors.client_id : null}
        disabled={isEdit}
      />
      <Input
        id="client_secret"
        name="client_secret"
        type="text"
        label="Client secret"
        onBlur={formik.handleBlur}
        onChange={formik.handleChange}
        value={formik.values.client_secret}
        error={
          formik.touched.client_secret ? formik.errors.client_secret : null
        }
      />
      <Input
        id="mapper"
        name="mapper"
        type="text"
        label="Mapper"
        onBlur={formik.handleBlur}
        onChange={formik.handleChange}
        value={formik.values.mapper}
        error={formik.touched.mapper ? formik.errors.mapper : null}
        disabled={isEdit}
      />
    </Form>
  );
};

export default ProviderForm;
