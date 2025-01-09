import { screen, waitFor } from "@testing-library/dom";
import { faker } from "@faker-js/faker";
import userEvent from "@testing-library/user-event";
import { Location } from "react-router-dom";
import MockAdapter from "axios-mock-adapter";
import * as reactQuery from "@tanstack/react-query";

import { renderComponent } from "test/utils";
import { urls } from "urls";
import { axiosInstance } from "api/axios";

import ProviderCreate from "./ProviderCreate";
import { ProviderFormLabel } from "../ProviderForm";
import { Label } from "./types";
import { initialValues } from "./ProviderCreate";
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
  mock.onPost("/idps").reply(200);
});

test("can cancel", async () => {
  let location: Location | null = null;
  renderComponent(<ProviderCreate />, {
    url: "/",
    setLocation: (newLocation) => {
      location = newLocation;
    },
  });
  await userEvent.click(screen.getByRole("button", { name: Label.CANCEL }));
  expect((location as Location | null)?.pathname).toBe(urls.providers.index);
});

test("calls the API on submit", async () => {
  const values = {
    id: faker.word.sample(),
  };
  renderComponent(<ProviderCreate />);
  const input = screen.getByRole("textbox", { name: ProviderFormLabel.NAME });
  await userEvent.click(input);
  await userEvent.clear(input);
  await userEvent.type(input, values.id);
  await userEvent.click(screen.getByRole("button", { name: Label.SUBMIT }));
  expect(mock.history.post[0].url).toBe("/idps");
  expect(JSON.parse(mock.history.post[0].data as string)).toMatchObject({
    ...initialValues,
    scope: initialValues.scope.split(","),
    ...values,
  });
});

test("handles API success", async () => {
  let location: Location | null = null;
  const invalidateQueries = vi.fn();
  vi.spyOn(reactQuery, "useQueryClient").mockReturnValue({
    invalidateQueries,
  } as unknown as reactQuery.QueryClient);
  mock.onPost("/idps").reply(200);
  const values = {
    id: faker.word.sample(),
  };
  renderComponent(
    <NotificationProvider>
      <ProviderCreate />
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
    screen.getByRole("textbox", { name: ProviderFormLabel.NAME }),
    values.id,
  );
  await userEvent.click(screen.getByRole("button", { name: Label.SUBMIT }));
  await waitFor(() =>
    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: [queryKeys.providers],
    }),
  );
  expect(document.querySelector(".p-notification--positive")).toHaveTextContent(
    Label.SUCCESS,
  ),
    expect((location as Location | null)?.pathname).toBe(urls.providers.index);
});

test("handles API failure", async () => {
  mock.onPost("/idps").reply(400, {
    message: "oops",
  });
  const values = {
    id: faker.word.sample(),
  };
  renderComponent(
    <NotificationProvider>
      <ProviderCreate />
      <NotificationConsumer />
    </NotificationProvider>,
  );
  await userEvent.type(
    screen.getByRole("textbox", { name: ProviderFormLabel.NAME }),
    values.id,
  );
  await userEvent.click(screen.getByRole("button", { name: Label.SUBMIT }));
  expect(document.querySelector(".p-notification--negative")).toHaveTextContent(
    `${Label.ERROR}oops`,
  );
});
