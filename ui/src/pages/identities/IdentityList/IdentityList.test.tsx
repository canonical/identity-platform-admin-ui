import { screen, within } from "@testing-library/dom";
import MockAdapter from "axios-mock-adapter";
import userEvent from "@testing-library/user-event";
import { faker } from "@faker-js/faker";

import { axiosInstance } from "api/axios";
import { customWithin } from "test/queries/within";
import { isoTimeToString } from "util/date";
import { LoaderTestId } from "components/Loader";
import { mockIdentity } from "test/mocks/identities";
import { mockPaginatedResponse } from "test/mocks/responses";
import { PAGE_SIZE } from "util/api";
import { renderComponent } from "test/utils";

import { DeleteIdentityBtnTestId } from "../DeleteIdentityBtn";

import { Label } from "./types";
import IdentityList from "./IdentityList";
import { Location } from "react-router-dom";
import { panels } from "util/usePanelParams";

const mock = new MockAdapter(axiosInstance);

beforeEach(() => {
  mock.reset();
  mock
    .onGet(`/identities?page_token=&page_size=${PAGE_SIZE}`)
    .reply(200, mockPaginatedResponse([], { _meta: {} }));
});

test("displays loading state", () => {
  renderComponent(<IdentityList />);
  expect(screen.getByTestId(LoaderTestId.COMPONENT)).toBeInTheDocument();
});

test("displays empty state", async () => {
  renderComponent(<IdentityList />);
  expect(
    await within(await screen.findByRole("grid")).findByText(Label.NO_DATA),
  ).toBeInTheDocument();
});

test("displays identity rows", async () => {
  const createdAt = faker.date.anytime().toISOString();
  const identity = mockIdentity({
    created_at: createdAt,
  });
  const identities = [identity];
  mock
    .onGet(`/identities?page_token=&size=${PAGE_SIZE}`)
    .reply(200, mockPaginatedResponse(identities, { _meta: {} }));
  renderComponent(<IdentityList />);
  const row = await screen.findByRole("row", {
    name: new RegExp(identity.id),
  });
  expect(
    customWithin(row).getCellByHeader(Label.HEADER_ID, { role: "rowheader" }),
  ).toHaveTextContent(identity.id);
  expect(
    customWithin(row).getCellByHeader(Label.HEADER_SCHEMA, {
      role: "gridcell",
      hasRowHeader: true,
    }),
  ).toHaveTextContent(identity.schema_id);
  expect(
    customWithin(row).getCellByHeader(Label.HEADER_CREATED_AT, {
      role: "gridcell",
      hasRowHeader: true,
    }),
  ).toHaveTextContent(isoTimeToString(createdAt));
});

test("opens add identity panel", async () => {
  let location: Location | null = null;
  renderComponent(<IdentityList />, {
    setLocation: (newLocation) => {
      location = newLocation;
    },
  });
  await userEvent.click(screen.getByRole("button", { name: Label.ADD }));
  expect((location as Location | null)?.search).toBe(
    `?panel=${panels.identityCreate}`,
  );
});

test("displays delete button", async () => {
  const identities = [mockIdentity()];
  mock
    .onGet(`/identities?page_token=&size=${PAGE_SIZE}`)
    .reply(200, mockPaginatedResponse(identities, { _meta: {} }));
  renderComponent(<IdentityList />);
  expect(
    await screen.findByTestId(DeleteIdentityBtnTestId.COMPONENT),
  ).toBeInTheDocument();
});
