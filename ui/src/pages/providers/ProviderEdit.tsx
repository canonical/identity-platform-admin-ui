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
import { useQuery, useQueryClient } from "@tanstack/react-query";
import ProviderForm, { ProviderFormTypes } from "pages/providers/ProviderForm";
import { fetchProvider, updateProvider } from "api/provider";
import SidePanel from "components/SidePanel";
import usePanelParams from "util/usePanelParams";
import ScrollableContainer from "components/ScrollableContainer";

const ProviderEdit: FC = () => {
  const notify = useNotify();
  const queryClient = useQueryClient();
  const panelParams = usePanelParams();
  const providerId = panelParams.id;

  if (!providerId) {
    return;
  }

  const { data: provider } = useQuery({
    queryKey: [queryKeys.providers, providerId],
    queryFn: () => fetchProvider(providerId),
  });

  const ProviderEditSchema = Yup.object().shape({
    id: Yup.string().required("This field is required"),
  });

  const formik = useFormik<ProviderFormTypes>({
    initialValues: {
      apple_private_key: provider?.apple_private_key,
      apple_private_key_id: provider?.apple_private_key_id,
      apple_team_id: provider?.apple_team_id,
      auth_url: provider?.auth_url,
      client_id: provider?.client_id,
      client_secret: provider?.client_secret,
      id: provider?.id,
      issuer_url: provider?.issuer_url,
      mapper_url: provider?.mapper_url,
      microsoft_tenant: provider?.microsoft_tenant,
      provider: provider?.provider,
      requested_claims: provider?.requested_claims,
      scope: provider?.scope?.join(","),
      subject_source: provider?.subject_source,
      token_url: provider?.token_url,
    },
    enableReinitialize: true,
    validationSchema: ProviderEditSchema,
    onSubmit: (values) => {
      updateProvider(
        provider?.id ?? "",
        JSON.stringify({ ...values, scope: values.scope?.split(",") }),
      )
        .then(() => {
          void queryClient.invalidateQueries({
            queryKey: [queryKeys.providers],
          });
          notify.success("Provider updated");
          panelParams.clear();
        })
        .catch((e) => {
          formik.setSubmitting(false);
          notify.failure("Provider update failed", e);
        });
    },
  });

  return (
    <SidePanel hasError={false} loading={false} className="p-panel">
      <ScrollableContainer dependencies={[]} belowId="panel-footer">
        <SidePanel.Header>
          <SidePanel.HeaderTitle>Edit provider</SidePanel.HeaderTitle>
        </SidePanel.Header>
        <SidePanel.Content className="u-no-padding">
          <Row>
            <ProviderForm formik={formik} isEdit={true} />
          </Row>
        </SidePanel.Content>
      </ScrollableContainer>
      <div id="panel-footer">
        <SidePanel.Footer className="u-align--right">
          <Row>
            <Col size={12}>
              <Button
                appearance="base"
                className="u-no-margin--bottom u-sv2"
                onClick={panelParams.clear}
              >
                Cancel
              </Button>
              <ActionButton
                appearance="positive"
                className="u-no-margin--bottom"
                loading={formik.isSubmitting}
                disabled={!formik.isValid}
                onClick={() => void formik.submitForm()}
              >
                Update
              </ActionButton>
            </Col>
          </Row>
        </SidePanel.Footer>
      </div>
    </SidePanel>
  );
};

export default ProviderEdit;
