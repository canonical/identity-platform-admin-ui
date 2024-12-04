import React, { FC } from "react";
import { Form, Input, Select } from "@canonical/react-components";
import { FormikProps } from "formik";
import { useQuery } from "@tanstack/react-query";
import { queryKeys } from "util/queryKeys";
import { fetchSchemas } from "api/schema";
import { Label } from "./types";

export interface IdentityFormTypes {
  schemaId: string;
  email: string;
}

interface Props {
  formik: FormikProps<IdentityFormTypes>;
}

const IdentityForm: FC<Props> = ({ formik }) => {
  const { data } = useQuery({
    queryKey: [queryKeys.schemas],
    queryFn: () => fetchSchemas(""),
  });

  const schemaOptions =
    data?.data.map((schema) => ({
      label: schema.id,
      value: schema.id,
    })) ?? [];

  return (
    <Form onSubmit={formik.handleSubmit} stacked>
      <Input
        {...formik.getFieldProps("email")}
        type="text"
        label={Label.EMAIL}
        error={formik.touched.email ? formik.errors.email : null}
      />
      <Select
        {...formik.getFieldProps("schemaId")}
        options={[
          {
            label: "Select option",
            value: "",
            disabled: true,
          },
          ...schemaOptions,
        ]}
        label={Label.SCHEMA}
        error={formik.touched.schemaId ? formik.errors.schemaId : null}
      />
    </Form>
  );
};

export default IdentityForm;
