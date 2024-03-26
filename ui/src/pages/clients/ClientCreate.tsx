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
import ClientForm, { ClientFormTypes } from "pages/clients/ClientForm";
import { createClient } from "api/client";
import SidePanel from "components/SidePanel";
import ScrollableContainer from "components/ScrollableContainer";

const ClientCreate: FC = () => {
  const navigate = useNavigate();
  const notify = useNotify();
  const queryClient = useQueryClient();

  const ClientCreateSchema = Yup.object().shape({
    client_name: Yup.string().required("This field is required"),
  });

  const formik = useFormik<ClientFormTypes>({
    initialValues: {
      client_uri: "",
      client_name: "grafana",
      grant_types: ["authorization_code", "refresh_token"],
      response_types: ["code", "id_token"],
      scope: "openid offline_access email",
      redirect_uris: ["http://localhost:2345/login/generic_oauth"],
      request_object_signing_alg: "RS256",
    },
    validationSchema: ClientCreateSchema,
    onSubmit: (values) => {
      createClient(JSON.stringify(values))
        .then((result) => {
          void queryClient.invalidateQueries({
            queryKey: [queryKeys.clients],
          });
          const msg = `Client created. Id: ${result.client_id} Secret: ${result.client_secret}`;
          navigate("/client", notify.queue(notify.success(msg)));
        })
        .catch((e) => {
          formik.setSubmitting(false);
          notify.failure("Client creation failed", e);
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
          <SidePanel.HeaderTitle>Add client</SidePanel.HeaderTitle>
        </SidePanel.Header>
        <SidePanel.Content>
          <Row>
            <ClientForm formik={formik} />
          </Row>
        </SidePanel.Content>
      </ScrollableContainer>
      <div id="panel-footer">
        <SidePanel.Footer>
          <Row className="u-align-text--right">
            <Col size={12}>
              <Button appearance="base" onClick={() => navigate("/client")}>
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

export default ClientCreate;
