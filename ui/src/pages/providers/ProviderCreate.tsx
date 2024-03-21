import React, { FC } from "react";
import {
  ActionButton,
  Button,
  Col,
  Row,
  useNotify,
} from "@canonical/react-components";
import { useFormik } from "formik";
import * as Yup from "yup";
import { queryKeys } from "util/queryKeys";
import { useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "react-router-dom";
import ProviderForm, { ProviderFormTypes } from "pages/providers/ProviderForm";
import { createProvider } from "api/provider";
import SidePanel from "components/SidePanel";

const ProviderCreate: FC = () => {
  const navigate = useNavigate();
  const notify = useNotify();
  const queryClient = useQueryClient();

  const ProviderCreateSchema = Yup.object().shape({
    id: Yup.string().required("This field is required"),
  });

  const formik = useFormik<ProviderFormTypes>({
    initialValues: {
      client_secret: "secret-9",
      id: "okta_347646e49b484037b83690b020f9f629",
      client_id: "347646e4-9b48-4037-b836-90b020f9f629",
      provider: "okta",
      mapper: "file:///etc/config/kratos/okta_schema.jsonnet",
      scope: "email",
    },
    validationSchema: ProviderCreateSchema,
    onSubmit: (values) => {
      createProvider(
        JSON.stringify({ ...values, scope: values.scope?.split(",") }),
      )
        .then(() => {
          void queryClient.invalidateQueries({
            queryKey: [queryKeys.providers],
          });
          const msg = `Provider created.`;
          navigate("/provider", notify.queue(notify.success(msg)));
        })
        .catch((e) => {
          formik.setSubmitting(false);
          notify.failure("Provider creation failed", e);
        });
    },
  });

  const submitForm = () => {
    void formik.submitForm();
  };

  return (
    <SidePanel hasError={false} loading={false} className="p-panel">
      <SidePanel.Header>
        <SidePanel.HeaderTitle>Add ID provider</SidePanel.HeaderTitle>
      </SidePanel.Header>
      <SidePanel.Content>
        <Row>
          <ProviderForm formik={formik} />
        </Row>
      </SidePanel.Content>
      <SidePanel.Footer>
        <Row className="u-align-text--right">
          <Col size={12}>
            <Button appearance="base" onClick={() => navigate("/provider")}>
              Cancel
            </Button>
            <ActionButton
              appearance="positive"
              loading={formik.isSubmitting}
              disabled={!formik.isValid}
              onClick={submitForm}
            >
              Save
            </ActionButton>
          </Col>
        </Row>
      </SidePanel.Footer>
    </SidePanel>
  );
};

export default ProviderCreate;
