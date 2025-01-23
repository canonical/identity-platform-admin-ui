import { screen, waitFor } from "@testing-library/dom";
import { faker } from "@faker-js/faker";
import userEvent from "@testing-library/user-event";
import { Location } from "react-router-dom";
import MockAdapter from "axios-mock-adapter";
import * as reactQuery from "@tanstack/react-query";

import { renderComponent } from "test/utils";
import { urls } from "urls";
import { axiosInstance } from "api/axios";

import ClientCreate from "./ClientCreate";
import { ClientFormLabel } from "../ClientForm";
import { Label } from "./types";
import { initialValues } from "./ClientCreate";
import {
  NotificationProvider,
  NotificationConsumer,
} from "@canonical/react-components";
import { queryKeys } from "util/queryKeys";

vi.mock("@tanstack/react-query", async () => {
  const actual = await vi.importActual("@tanstack/react-query");
  return {
    ...actual,
    useQueryClient: vi.fn(),
  };
});

const mock = new MockAdapter(axiosInstance);

beforeEach(() => {
  mock.reset();
  vi.spyOn(reactQuery, "useQueryClient").mockReturnValue({
    invalidateQueries: vi.fn(),
  } as unknown as reactQuery.QueryClient);
  mock.onPost("/clients").reply(200);
});

test("can cancel", async () => {
  let location: Location | null = null;
  renderComponent(<ClientCreate />, {
    url: "/",
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
  renderComponent(<ClientCreate />);
  const input = screen.getByRole("textbox", { name: ClientFormLabel.NAME });
  await userEvent.click(input);
  await userEvent.clear(input);
  await userEvent.type(input, values.client_name);
  await userEvent.click(screen.getByRole("button", { name: Label.SUBMIT }));
  expect(mock.history.post[0].url).toBe("/clients");
  expect(mock.history.post[0].data).toBe(
    JSON.stringify({
      ...initialValues,
      ...values,
    }),
  );
});

test("handles API success", async () => {
  let location: Location | null = null;
  const invalidateQueries = vi.fn();
  vi.spyOn(reactQuery, "useQueryClient").mockReturnValue({
    invalidateQueries,
  } as unknown as reactQuery.QueryClient);
  mock.onPost("/clients").reply(200, {
    data: { client_id: "client1", client_secret: "secret1" },
  });
  const values = {
    client_name: faker.word.sample(),
  };
  renderComponent(
    <NotificationProvider>
      <ClientCreate />
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
    "Client created. Id: client1 Secret: secret1",
  ),
    expect((location as Location | null)?.pathname).toBe(urls.clients.index);
});

test("handles API failure", async () => {
  mock.onPost("/clients").reply(400, {
    message: "oops",
  });
  const values = {
    client_name: faker.word.sample(),
  };
  renderComponent(
    <NotificationProvider>
      <ClientCreate />
      <NotificationConsumer />
    </NotificationProvider>,
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
