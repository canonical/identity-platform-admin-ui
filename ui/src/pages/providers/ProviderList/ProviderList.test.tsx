import { screen, waitFor } from "@testing-library/dom";
import MockAdapter from "axios-mock-adapter";
import userEvent from "@testing-library/user-event";
import { faker } from "@faker-js/faker";

import { axiosInstance } from "api/axios";
import { customWithin } from "test/queries/within";
import { mockIdentityProvider } from "test/mocks/providers";
import { mockPaginatedResponse } from "test/mocks/responses";
import { renderComponent } from "test/utils";

import { DeleteProviderBtnTestId } from "../DeleteProviderBtn";
import { EditProviderBtnTestId } from "../EditProviderBtn";

import { Label } from "./types";
import ProviderList from "./ProviderList";
import { Location } from "react-router";
import { panels } from "util/usePanelParams";

const mock = new MockAdapter(axiosInstance);

beforeEach(() => {
  mock.reset();
  mock.onGet("/idps").reply(200, mockPaginatedResponse([], { _meta: {} }));
});

test("displays provider rows", async () => {
  const provider = {
    id: faker.word.sample(),
    provider: faker.word.sample(),
  };
  const providers = [mockIdentityProvider(provider)];
  mock
    .onGet("/idps")
    .reply(200, mockPaginatedResponse(providers, { _meta: {} }));
  renderComponent(<ProviderList />);
  const row = await screen.findByRole("row", {
    name: new RegExp(provider.id),
  });
  expect(
    customWithin(row).getCellByHeader(Label.HEADER_NAME, { role: "rowheader" }),
  ).toHaveTextContent(provider.id);
  expect(
    customWithin(row).getCellByHeader(Label.HEADER_PROVIDER, {
      role: "gridcell",
      hasRowHeader: true,
    }),
  ).toHaveTextContent(provider.provider);
});

test("opens add provider panel", async () => {
  let location: Location | null = null;
  renderComponent(<ProviderList />, {
    setLocation: (newLocation) => {
      location = newLocation;
    },
  });
  await userEvent.click(screen.getByRole("button", { name: Label.ADD }));
  await waitFor(() => {
    expect(location?.search).toBe(`?panel=${panels.providerCreate}`);
  });
});

test("displays edit and delete buttons", async () => {
  const providers = [mockIdentityProvider()];
  mock
    .onGet("/idps")
    .reply(200, mockPaginatedResponse(providers, { _meta: {} }));
  renderComponent(<ProviderList />);
  expect(
    await screen.findByTestId(EditProviderBtnTestId.COMPONENT),
  ).toBeInTheDocument();
  expect(
    await screen.findByTestId(DeleteProviderBtnTestId.COMPONENT),
  ).toBeInTheDocument();
});
