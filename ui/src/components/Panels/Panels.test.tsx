import { screen } from "@testing-library/dom";

import { renderComponent } from "test/utils";
import MockAdapter from "axios-mock-adapter";

import { axiosInstance } from "api/axios";
import { panels } from "util/usePanelParams";
import { ProviderCreateTestId } from "pages/providers/ProviderCreate";
import { ProviderEditTestId } from "pages/providers/ProviderEdit";
import { ClientCreateTestId } from "pages/clients/ClientCreate";
import { ClientEditTestId } from "pages/clients/ClientEdit";
import { IdentityCreateTestId } from "pages/identities/IdentityCreate";

import Panels from "./Panels";

const mock = new MockAdapter(axiosInstance);

beforeEach(() => {
  mock.reset();
});

test("can display no panel", () => {
  const { result } = renderComponent(<Panels />);
  expect(result.container.firstChild).toBeNull();
});

test("can display the create provider panel", async () => {
  renderComponent(<Panels />, {
    url: `/?panel=${panels.providerCreate}`,
  });
  expect(
    await screen.findByTestId(ProviderCreateTestId.COMPONENT),
  ).toBeInTheDocument();
});

test("can display the edit provider panel", async () => {
  renderComponent(<Panels />, {
    url: `/?panel=${panels.providerEdit}&id=testid`,
  });
  expect(
    await screen.findByTestId(ProviderEditTestId.COMPONENT),
  ).toBeInTheDocument();
});

test("can display the create client panel", async () => {
  renderComponent(<Panels />, {
    url: `/?panel=${panels.clientCreate}`,
  });
  expect(
    await screen.findByTestId(ClientCreateTestId.COMPONENT),
  ).toBeInTheDocument();
});

test("can display the edit client panel", async () => {
  renderComponent(<Panels />, {
    url: `/?panel=${panels.clientEdit}&id=testid`,
  });
  expect(
    await screen.findByTestId(ClientEditTestId.COMPONENT),
  ).toBeInTheDocument();
});

test("can display the create identity panel", async () => {
  renderComponent(<Panels />, {
    url: `/?panel=${panels.identityCreate}`,
  });
  expect(
    await screen.findByTestId(IdentityCreateTestId.COMPONENT),
  ).toBeInTheDocument();
});
