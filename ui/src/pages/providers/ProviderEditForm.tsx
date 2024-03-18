import React, { FC } from "react";
import { Button, Col, Row, useNotify } from "@canonical/react-components";
import { useFormik } from "formik";
import * as Yup from "yup";
import { queryKeys } from "util/queryKeys";
import { useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "react-router-dom";
import { NotificationConsumer } from "@canonical/react-components/dist/components/NotificationProvider/NotificationProvider";
import SubmitButton from "components/SubmitButton";
import { IdentityProvider } from "types/provider";
import ProviderForm, { ProviderFormTypes } from "pages/providers/ProviderForm";
import { updateProvider } from "api/provider";

interface Props {
  provider: IdentityProvider;
}

const ProviderEditForm: FC<Props> = ({ provider }) => {
  const navigate = useNavigate();
  const notify = useNotify();
  const queryClient = useQueryClient();

  const ProviderEditSchema = Yup.object().shape({
    id: Yup.string().required("This field is required"),
  });

  const formik = useFormik<ProviderFormTypes>({
    initialValues: {
      client_secret: provider.client_secret,
      id: provider.id,
      client_id: provider.client_id,
      provider: provider.provider,
      mapper: provider.mapper_url,
      scope: provider.scope,
    },
    validationSchema: ProviderEditSchema,
    onSubmit: (values) => {
      updateProvider(provider.id, JSON.stringify(values))
        .then(() => {
          void queryClient.invalidateQueries({
            queryKey: [queryKeys.providers],
          });
          navigate(
            `/provider/detail/${provider.id}`,
            notify.queue(notify.success("Provider updated")),
          );
        })
        .catch((e) => {
          formik.setSubmitting(false);
          notify.failure("Provider update failed", e);
        });
    },
  });

  const submitForm = () => {
    void formik.submitForm();
  };

  return (
    <>
      <Row>
        <Col size={12}>
          <NotificationConsumer />
          <ProviderForm formik={formik} />
        </Col>
      </Row>
      <hr />
      <Row className="u-align--right u-sv2">
        <Col size={12}>
          <Button
            appearance="base"
            className="u-no-margin--bottom u-sv2"
            onClick={() => navigate(`/provider/detail/${provider.id}`)}
          >
            Cancel
          </Button>
          <SubmitButton
            isSubmitting={formik.isSubmitting}
            isDisabled={!formik.isValid}
            onClick={submitForm}
            buttonLabel="Update provider"
            className="u-no-margin--bottom"
          />
        </Col>
      </Row>
    </>
  );
};

export default ProviderEditForm;
