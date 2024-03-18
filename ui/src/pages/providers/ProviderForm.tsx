import React, { FC } from "react";
import { Form, Input } from "@canonical/react-components";
import { FormikProps } from "formik";

export interface ProviderFormTypes {
  id: string;
  client_id: string;
  client_secret: string;
  provider: string;
  mapper: string;
  scope: string[];
}

interface Props {
  formik: FormikProps<ProviderFormTypes>;
}

const ProviderForm: FC<Props> = ({ formik }) => {
  return (
    <Form onSubmit={formik.handleSubmit} stacked>
      <Input
        id="id"
        name="id"
        type="text"
        label="ID"
        onBlur={formik.handleBlur}
        onChange={formik.handleChange}
        value={formik.values.id}
        error={formik.touched.id ? formik.errors.id : null}
        required
        stacked
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
        stacked
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
        stacked
      />
      <Input
        id="provider"
        name="provider"
        type="text"
        label="Provider"
        onBlur={formik.handleBlur}
        onChange={formik.handleChange}
        value={formik.values.provider}
        error={formik.touched.provider ? formik.errors.provider : null}
        stacked
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
        stacked
      />
      <Input
        id="scope"
        name="scope"
        type="text"
        label="Scope"
        onBlur={formik.handleBlur}
        onChange={formik.handleChange}
        value={formik.values.scope}
        error={formik.touched.scope ? formik.errors.scope : null}
        stacked
      />
    </Form>
  );
};

export default ProviderForm;
