import { screen, within } from "@testing-library/dom";
import MockAdapter from "axios-mock-adapter";

import { axiosInstance } from "api/axios";
import { LoaderTestId } from "components/Loader";
import { customWithin } from "test/queries/within";
import { renderComponent } from "test/utils";

import SchemaList from "./SchemaList";
import { Label } from "./types";
import { PAGE_SIZE } from "util/api";
import { mockSchema } from "test/mocks/schemas";
import { mockPaginatedResponse } from "test/mocks/responses";

const mock = new MockAdapter(axiosInstance);

beforeEach(() => {
  mock.reset();
  mock
    .onGet(`/schemas?page_token=&page_size=${PAGE_SIZE}`)
    .reply(200, mockPaginatedResponse([], { _meta: {} }));
});

test("displays loading state", () => {
  renderComponent(<SchemaList />);
  expect(screen.getByTestId(LoaderTestId.COMPONENT)).toBeInTheDocument();
});

test("displays empty state", async () => {
  renderComponent(<SchemaList />);
  expect(
    await within(await screen.findByRole("grid")).findByText(Label.NO_DATA),
  ).toBeInTheDocument();
});

test("displays schema rows", async () => {
  const schemas = [
    mockSchema({
      schema: {
        test: "schema",
      },
    }),
  ];
  mock
    .onGet(`/schemas?page_token=&page_size=${PAGE_SIZE}`)
    .reply(200, mockPaginatedResponse(schemas, { _meta: {} }));
  renderComponent(<SchemaList />, {});
  const row = await screen.findByRole("row", {
    name: new RegExp(schemas[0].id),
  });
  expect(
    customWithin(row).getCellByHeader(Label.HEADER_ID, { role: "rowheader" }),
  ).toHaveTextContent(schemas[0].id);
  expect(
    customWithin(row).getCellByHeader(Label.HEADER_SCHEMA, {
      role: "gridcell",
      hasRowHeader: true,
    }),
  ).toHaveTextContent(JSON.stringify(schemas[0].schema));
});
