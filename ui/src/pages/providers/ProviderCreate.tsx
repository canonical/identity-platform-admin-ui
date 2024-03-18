import React, { FC } from "react";
import { Button, Col, Row, useNotify } from "@canonical/react-components";
import { useFormik } from "formik";
import * as Yup from "yup";
import { queryKeys } from "util/queryKeys";
import { useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "react-router-dom";
import { NotificationConsumer } from "@canonical/react-components/dist/components/NotificationProvider/NotificationProvider";
import SubmitButton from "components/SubmitButton";
import ProviderForm, { ProviderFormTypes } from "pages/providers/ProviderForm";
import { createProvider } from "api/provider";

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
      scope: ["email"],
    },
    validationSchema: ProviderCreateSchema,
    onSubmit: (values) => {
      createProvider(JSON.stringify(values))
        .then(() => {
          void queryClient.invalidateQueries({
            queryKey: [queryKeys.providers],
          });
          const msg = `Provider created.`;
          navigate("/provider/list", notify.queue(notify.success(msg)));
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
    <div className="p-panel">
      <div className="p-panel__header ">
        <div className="p-panel__title">
          <h1 className="p-heading--4 u-no-margin--bottom">
            Create new provider
          </h1>
        </div>
      </div>
      <div className="p-panel__content">
        <Row>
          <Col size={12}>
            <NotificationConsumer />
            <ProviderForm formik={formik} />
          </Col>
        </Row>
        <hr />
        <Row className="u-align--right">
          <Col size={12}>
            <Button
              appearance="base"
              onClick={() => navigate("/provider/list")}
            >
              Cancel
            </Button>
            <SubmitButton
              isSubmitting={formik.isSubmitting}
              isDisabled={!formik.isValid}
              onClick={submitForm}
              buttonLabel="Create"
            />
          </Col>
        </Row>
      </div>
    </div>
  );
};

export default ProviderCreate;
