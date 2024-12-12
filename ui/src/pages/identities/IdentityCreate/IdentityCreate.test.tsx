import { screen, waitFor } from "@testing-library/dom";
import { faker } from "@faker-js/faker";
import userEvent from "@testing-library/user-event";
import { Location } from "react-router-dom";
import MockAdapter from "axios-mock-adapter";
import * as reactQuery from "@tanstack/react-query";

import { renderComponent } from "test/utils";
import { urls } from "urls";
import { axiosInstance } from "api/axios";

import IdentityCreate from "./IdentityCreate";
import { IdentityFormLabel } from "../IdentityForm";
import { Label } from "./types";
import {
  NotificationProvider,
  NotificationConsumer,
} from "@canonical/react-components";
import { queryKeys } from "util/queryKeys";
import { mockSchema } from "test/mocks/schemas";
import { Schema } from "types/schema";

vi.mock("@tanstack/react-query", async () => {
  const actual = await vi.importActual("@tanstack/react-query");
  return {
    ...actual,
    useQueryIdentity: vi.fn(),
  };
});

const mock = new MockAdapter(axiosInstance);

let schema: Schema;
beforeEach(() => {
  mock.reset();
  vi.spyOn(reactQuery, "useQueryClient").mockReturnValue({
    invalidateQueries: vi.fn(),
  } as unknown as reactQuery.QueryClient);
  mock.onPost("/identities").reply(200);
  schema = mockSchema();
  mock
    .onGet("/schemas?page_token=&page_size=50")
    .reply(200, { data: [schema] });
});

test("can cancel", async () => {
  let location: Location | null = null;
  renderComponent(<IdentityCreate />, {
    url: "/",
    setLocation: (newLocation) => {
      location = newLocation;
    },
  });
  await userEvent.click(screen.getByRole("button", { name: Label.CANCEL }));
  expect((location as Location | null)?.pathname).toBe(urls.identities.index);
});

test("calls the API on submit", async () => {
  const values = {
    email: faker.word.sample(),
    schemaId: schema.id,
  };
  renderComponent(<IdentityCreate />);
  const emailInput = screen.getByRole("textbox", {
    name: IdentityFormLabel.EMAIL,
  });
  await userEvent.click(emailInput);
  await userEvent.clear(emailInput);
  await userEvent.type(emailInput, values.email);
  await userEvent.selectOptions(
    screen.getByRole("combobox", { name: IdentityFormLabel.SCHEMA }),
    values.schemaId,
  );
  await userEvent.click(screen.getByRole("button", { name: Label.SUBMIT }));
  expect(mock.history.post[0].url).toBe("/identities");
  expect(JSON.parse(mock.history.post[0].data as string)).toMatchObject({
    schema_id: values.schemaId,
    traits: {
      email: values.email,
    },
  });
});

test("handles API success", async () => {
  let location: Location | null = null;
  const invalidateQueries = vi.fn();
  vi.spyOn(reactQuery, "useQueryClient").mockReturnValue({
    invalidateQueries,
  } as unknown as reactQuery.QueryClient);
  mock.onPost("/identities").reply(200);
  renderComponent(
    <NotificationProvider>
      <IdentityCreate />
      <NotificationConsumer />
    </NotificationProvider>,
    {
      url: "/",
      setLocation: (newLocation) => {
        location = newLocation;
      },
    },
  );
  await userEvent.type(
    screen.getByRole("textbox", {
      name: IdentityFormLabel.EMAIL,
    }),
    faker.word.sample(),
  );
  await userEvent.selectOptions(
    screen.getByRole("combobox", { name: IdentityFormLabel.SCHEMA }),
    schema.id,
  );
  await userEvent.click(screen.getByRole("button", { name: Label.SUBMIT }));
  await waitFor(() =>
    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: [queryKeys.identities],
    }),
  );
  expect(document.querySelector(".p-notification--positive")).toHaveTextContent(
    Label.SUCCESS,
  ),
    expect((location as Location | null)?.pathname).toBe(urls.identities.index);
});

test("handles API failure", async () => {
  mock.onPost("/identities").reply(400, {
    message: "oops",
  });
  renderComponent(
    <NotificationProvider>
      <IdentityCreate />
      <NotificationConsumer />
    </NotificationProvider>,
  );
  await userEvent.type(
    screen.getByRole("textbox", {
      name: IdentityFormLabel.EMAIL,
    }),
    faker.word.sample(),
  );
  await userEvent.selectOptions(
    screen.getByRole("combobox", { name: IdentityFormLabel.SCHEMA }),
    schema.id,
  );
  await userEvent.click(screen.getByRole("button", { name: Label.SUBMIT }));
  expect(document.querySelector(".p-notification--negative")).toHaveTextContent(
    `${Label.ERROR}oops`,
  );
});
