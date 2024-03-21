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
import { useQuery, useQueryClient } from "@tanstack/react-query";
import ProviderForm, { ProviderFormTypes } from "pages/providers/ProviderForm";
import { fetchProvider, updateProvider } from "api/provider";
import SidePanel from "components/SidePanel";
import usePanelParams from "util/usePanelParams";

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
      client_secret: provider?.client_secret,
      id: provider?.id,
      client_id: provider?.client_id,
      provider: provider?.provider,
      mapper: provider?.mapper_url,
      scope: provider?.scope.join(","),
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
      <SidePanel.Header>
        <SidePanel.HeaderTitle>Edit provider</SidePanel.HeaderTitle>
      </SidePanel.Header>
      <SidePanel.Content className="u-no-padding">
        <Row>
          <ProviderForm formik={formik} isEdit={true} />
        </Row>
      </SidePanel.Content>
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
    </SidePanel>
  );
};

export default ProviderEdit;
