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
import IdentityForm, { IdentityFormTypes } from "pages/identities/IdentityForm";
import { createIdentity } from "api/identities";
import SidePanel from "components/SidePanel";
import ScrollableContainer from "components/ScrollableContainer";
import { TestId } from "./test-types";
import { testId } from "test/utils";
import { Label } from "./types";

const IdentityCreate: FC = () => {
  const navigate = useNavigate();
  const notify = useNotify();
  const queryClient = useQueryClient();

  const IdentityCreateSchema = Yup.object().shape({
    email: Yup.string().required("This field is required"),
    schemaId: Yup.string().required("This field is required"),
  });

  const formik = useFormik<IdentityFormTypes>({
    initialValues: {
      email: "",
      schemaId: "",
    },
    validationSchema: IdentityCreateSchema,
    validateOnMount: true,
    onSubmit: (values) => {
      const identity = {
        schema_id: values.schemaId,
        traits: { email: values.email },
      };
      createIdentity(JSON.stringify(identity))
        .then(() => {
          void queryClient.invalidateQueries({
            queryKey: [queryKeys.identities],
          });
          navigate("/identity", notify.queue(notify.success(Label.SUCCESS)));
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
          <SidePanel.HeaderTitle>Add identity</SidePanel.HeaderTitle>
        </SidePanel.Header>
        <SidePanel.Content>
          <Row>
            <IdentityForm formik={formik} />
          </Row>
        </SidePanel.Content>
      </ScrollableContainer>
      <div id="panel-footer">
        <SidePanel.Footer>
          <Row className="u-align-text--right">
            <Col size={12}>
              <Button appearance="base" onClick={() => navigate("/identity")}>
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

export default IdentityCreate;
