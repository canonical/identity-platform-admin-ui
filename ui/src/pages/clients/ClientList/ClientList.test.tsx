import { screen, within } from "@testing-library/dom";
import MockAdapter from "axios-mock-adapter";
import userEvent from "@testing-library/user-event";

import { axiosInstance } from "api/axios";
import { customWithin } from "test/queries/within";
import { isoTimeToString } from "util/date";
import { LoaderTestId } from "components/Loader";
import { mockClient } from "test/mocks/clients";
import { mockPaginatedResponse } from "test/mocks/responses";
import { PAGE_SIZE } from "util/api";
import { renderComponent } from "test/utils";

import { DeleteClientBtnTestId } from "../DeleteClientBtn";
import { EditClientBtnTestId } from "../EditClientBtn";

import { Label } from "./types";
import ClientList from "./ClientList";
import { Location } from "react-router-dom";
import { panels } from "util/usePanelParams";

const mock = new MockAdapter(axiosInstance);

beforeEach(() => {
  mock.reset();
  mock
    .onGet(`/clients?page_token=&page_size=${PAGE_SIZE}`)
    .reply(200, mockPaginatedResponse([], { _meta: {} }));
});

test("displays loading state", () => {
  renderComponent(<ClientList />);
  expect(screen.getByTestId(LoaderTestId.COMPONENT)).toBeInTheDocument();
});

test("displays empty state", async () => {
  renderComponent(<ClientList />);
  expect(
    await within(await screen.findByRole("grid")).findByText(Label.NO_DATA),
  ).toBeInTheDocument();
});

test("displays client rows", async () => {
  const client = mockClient();
  const clients = [client];
  mock
    .onGet(`/clients?page_token=&size=${PAGE_SIZE}`)
    .reply(200, mockPaginatedResponse(clients, { _meta: {} }));
  renderComponent(<ClientList />);
  const row = await screen.findByRole("row", {
    name: new RegExp(client.client_id),
  });
  expect(
    customWithin(row).getCellByHeader(Label.HEADER_ID, { role: "rowheader" }),
  ).toHaveTextContent(client.client_id);
  expect(
    customWithin(row).getCellByHeader(Label.HEADER_NAME, {
      role: "gridcell",
      hasRowHeader: true,
    }),
  ).toHaveTextContent(client.client_name);
  expect(
    customWithin(row).getCellByHeader(Label.HEADER_DATE, {
      role: "gridcell",
      hasRowHeader: true,
    }),
  ).toHaveTextContent(
    `Created: ${isoTimeToString(client.created_at)}Updated: ${isoTimeToString(client.updated_at)}`,
  );
});

test("opens add client panel", async () => {
  let location: Location | null = null;
  renderComponent(<ClientList />, {
    setLocation: (newLocation) => {
      location = newLocation;
    },
  });
  await userEvent.click(screen.getByRole("button", { name: Label.ADD }));
  expect((location as Location | null)?.search).toBe(
    `?panel=${panels.clientCreate}`,
  );
});

test("displays edit and delete buttons", async () => {
  const clients = [mockClient()];
  mock
    .onGet(`/clients?page_token=&size=${PAGE_SIZE}`)
    .reply(200, mockPaginatedResponse(clients, { _meta: {} }));
  renderComponent(<ClientList />);
  expect(
    await screen.findByTestId(EditClientBtnTestId.COMPONENT),
  ).toBeInTheDocument();
  expect(
    await screen.findByTestId(DeleteClientBtnTestId.COMPONENT),
  ).toBeInTheDocument();
});
