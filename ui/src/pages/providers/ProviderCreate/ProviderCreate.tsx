// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

import { FC } from "react";
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
import { TestId } from "./test-types";
import { testId } from "test/utils";
import { Label } from "./types";

export const initialValues = {
  provider: "generic",
  id: "",
  client_id: "",
  client_secret: "",
  mapper_url: "",
  scope: "email",
  subject_source: "userinfo",
} as const;

const ProviderCreate: FC = () => {
  const navigate = useNavigate();
  const notify = useNotify();
  const queryClient = useQueryClient();

  const ProviderCreateSchema = Yup.object().shape({
    id: Yup.string().required("This field is required"),
  });

  const formik = useFormik<ProviderFormTypes>({
    initialValues,
    validationSchema: ProviderCreateSchema,
    onSubmit: (values) => {
      createProvider(
        JSON.stringify({ ...values, scope: values.scope?.split(",") }),
      )
        .then(() => {
          void queryClient.invalidateQueries({
            queryKey: [queryKeys.providers],
          });
          navigate("/provider", notify.queue(notify.success(Label.SUCCESS)));
        })
        .catch((error: unknown) => {
          formik.setSubmitting(false);
          notify.failure(
            Label.ERROR,
            error instanceof Error ? error : null,
            typeof error === "string" ? error : null,
          );
        });
    },
  });

  const submitForm = () => {
    void formik.submitForm();
  };

  return (
    <SidePanel
      hasError={false}
      loading={false}
      className="p-panel"
      {...testId(TestId.COMPONENT)}
    >
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
                {Label.CANCEL}
              </Button>
              <ActionButton
                appearance="positive"
                loading={formik.isSubmitting}
                disabled={!formik.isValid}
                onClick={submitForm}
              >
                {Label.SUBMIT}
              </ActionButton>
            </Col>
          </Row>
        </SidePanel.Footer>
      </div>
    </SidePanel>
  );
};

export default ProviderCreate;
