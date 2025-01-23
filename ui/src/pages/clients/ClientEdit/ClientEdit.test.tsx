import { screen, waitFor } from "@testing-library/dom";
import { faker } from "@faker-js/faker";
import userEvent from "@testing-library/user-event";
import { Location } from "react-router-dom";
import MockAdapter from "axios-mock-adapter";
import * as reactQuery from "@tanstack/react-query";

import { renderComponent } from "test/utils";
import { urls } from "urls";
import { axiosInstance } from "api/axios";

import ClientEdit from "./ClientEdit";
import { ClientFormLabel } from "../ClientForm";
import { Label } from "./types";
import {
  NotificationProvider,
  NotificationConsumer,
} from "@canonical/react-components";
import { queryKeys } from "util/queryKeys";
import { mockClient } from "test/mocks/clients";
import { Client } from "types/client";

vi.mock("@tanstack/react-query", async () => {
  const actual = await vi.importActual("@tanstack/react-query");
  return {
    ...actual,
    useQueryClient: vi.fn(),
  };
});

const mock = new MockAdapter(axiosInstance);

let client: Client;

beforeEach(() => {
  vi.spyOn(reactQuery, "useQueryClient").mockReturnValue({
    invalidateQueries: vi.fn(),
  } as unknown as reactQuery.QueryClient);
  mock.reset();
  client = mockClient();
  mock.onGet(`/clients/${client.client_id}`).reply(200, { data: client });
  mock.onPut(`/clients/${client.client_id}`).reply(200);
});

test("can cancel", async () => {
  let location: Location | null = null;
  renderComponent(<ClientEdit />, {
    url: `/?id=${client.client_id}`,
    setLocation: (newLocation) => {
      location = newLocation;
    },
  });
  await userEvent.click(screen.getByRole("button", { name: Label.CANCEL }));
  expect((location as Location | null)?.pathname).toBe(urls.clients.index);
});

test("calls the API on submit", async () => {
  const values = {
    client_name: faker.word.sample(),
  };
  renderComponent(<ClientEdit />, {
    url: `/?id=${client.client_id}`,
  });
  const input = screen.getByRole("textbox", { name: ClientFormLabel.NAME });
  await userEvent.click(input);
  await userEvent.clear(input);
  await userEvent.type(input, values.client_name);
  await userEvent.click(screen.getByRole("button", { name: Label.SUBMIT }));
  expect(mock.history.put[0].url).toBe(`/clients/${client.client_id}`);
  expect(JSON.parse(mock.history.put[0].data as string)).toMatchObject({
    client_uri: client.client_uri,
    grant_types: client.grant_types,
    response_types: client.response_types,
    scope: client.scope,
    redirect_uris: client.redirect_uris,
    request_object_signing_alg: client.request_object_signing_alg,
    ...values,
  });
});

test("handles API success", async () => {
  let location: Location | null = null;
  const invalidateQueries = vi.fn();
  vi.spyOn(reactQuery, "useQueryClient").mockReturnValue({
    invalidateQueries,
  } as unknown as reactQuery.QueryClient);
  mock.onPut(`/clients/${client.client_id}`).reply(200, {
    data: { client_id: "client1", client_secret: "secret1" },
  });
  const values = {
    client_name: faker.word.sample(),
  };
  renderComponent(
    <NotificationProvider>
      <ClientEdit />
      <NotificationConsumer />
    </NotificationProvider>,
    {
      url: `/?id=${client.client_id}`,
      setLocation: (newLocation) => {
        location = newLocation;
      },
    },
  );
  await userEvent.type(
    screen.getByRole("textbox", { name: ClientFormLabel.NAME }),
    values.client_name,
  );
  await userEvent.click(screen.getByRole("button", { name: Label.SUBMIT }));
  await waitFor(() =>
    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: [queryKeys.clients],
    }),
  );
  expect(document.querySelector(".p-notification--positive")).toHaveTextContent(
    Label.SUCCESS,
  ),
    expect((location as Location | null)?.pathname).toBe(urls.clients.index);
});

test("handles API failure", async () => {
  mock.onPut(`/clients/${client.client_id}`).reply(400, {
    message: "oops",
  });
  const values = {
    client_name: faker.word.sample(),
  };
  renderComponent(
    <NotificationProvider>
      <ClientEdit />
      <NotificationConsumer />
    </NotificationProvider>,
    { url: `/?id=${client.client_id}` },
  );
  await userEvent.type(
    screen.getByRole("textbox", { name: ClientFormLabel.NAME }),
    values.client_name,
  );
  await userEvent.click(screen.getByRole("button", { name: Label.SUBMIT }));
  expect(document.querySelector(".p-notification--negative")).toHaveTextContent(
    `${Label.ERROR}oops`,
  );
});
