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
import ScrollableContainer from "components/ScrollableContainer";

const ProviderCreate: FC = () => {
  const navigate = useNavigate();
  const notify = useNotify();
  const queryClient = useQueryClient();

  const ProviderCreateSchema = Yup.object().shape({
    id: Yup.string().required("This field is required"),
  });

  const formik = useFormik<ProviderFormTypes>({
    initialValues: {
      provider: "generic",
      id: "",
      client_id: "",
      client_secret: "",
      mapper_url: "file:///etc/config/kratos/okta_schema.jsonnet",
      scope: "email",
      subject_source: "userinfo",
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
      <ScrollableContainer dependencies={[]} belowId="panel-footer">
        <SidePanel.Header>
          <SidePanel.HeaderTitle>Add ID provider</SidePanel.HeaderTitle>
        </SidePanel.Header>
        <SidePanel.Content>
          <Row>
            <ProviderForm formik={formik} />
          </Row>
        </SidePanel.Content>
      </ScrollableContainer>
      <div id="panel-footer">
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
      </div>
    </SidePanel>
  );
};

export default ProviderCreate;
