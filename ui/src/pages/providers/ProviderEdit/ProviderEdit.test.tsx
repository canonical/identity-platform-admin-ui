import { screen, waitFor } from "@testing-library/dom";
import { faker } from "@faker-js/faker";
import userEvent from "@testing-library/user-event";
import { Location } from "react-router-dom";
import MockAdapter from "axios-mock-adapter";
import * as reactQuery from "@tanstack/react-query";

import { renderComponent } from "test/utils";
import { axiosInstance } from "api/axios";

import ProviderEdit from "./ProviderEdit";
import { ProviderFormLabel } from "../ProviderForm";
import { Label } from "./types";
import {
  NotificationProvider,
  NotificationConsumer,
} from "@canonical/react-components";
import { queryKeys } from "util/queryKeys";
import { mockIdentityProvider } from "test/mocks/providers";
import { IdentityProvider } from "types/provider";

vi.mock("@tanstack/react-query", async () => {
  const actual = await vi.importActual("@tanstack/react-query");
  return {
    ...actual,
    useQueryClient: vi.fn(),
  };
});

const mock = new MockAdapter(axiosInstance);

let provider: IdentityProvider;

beforeEach(() => {
  vi.spyOn(reactQuery, "useQueryClient").mockReturnValue({
    invalidateQueries: vi.fn(),
  } as unknown as reactQuery.QueryClient);
  mock.reset();
  provider = mockIdentityProvider({
    id: faker.word.sample(),
    apple_private_key: faker.word.sample(),
    apple_private_key_id: faker.word.sample(),
    apple_team_id: faker.word.sample(),
    auth_url: faker.word.sample(),
    client_id: faker.word.sample(),
    client_secret: faker.word.sample(),
    issuer_url: faker.word.sample(),
    mapper_url: faker.word.sample(),
    microsoft_tenant: faker.word.sample(),
    provider: faker.word.sample(),
    requested_claims: faker.word.sample(),
    subject_source: "userinfo",
    token_url: faker.word.sample(),
    scope: ["email"],
  });
  mock.onGet(`/idps/${provider.id}`).reply(200, { data: [provider] });
  mock.onPatch(`/idps/${provider.id}`).reply(200);
});

test("can cancel", async () => {
  let location: Location | null = null;
  renderComponent(<ProviderEdit />, {
    url: `/?id=${provider.id}`,
    setLocation: (newLocation) => {
      location = newLocation;
    },
  });
  await userEvent.click(screen.getByRole("button", { name: Label.CANCEL }));
  expect((location as Location | null)?.pathname).toBe("/");
  expect((location as Location | null)?.search).toBe("");
});

test("calls the API on submit", async () => {
  const values = {
    scope: faker.word.sample(),
  };
  renderComponent(<ProviderEdit />, {
    url: `/?id=${provider.id}`,
  });
  const input = screen.getByRole("textbox", { name: ProviderFormLabel.SCOPES });
  await userEvent.click(input);
  await userEvent.clear(input);
  await userEvent.type(input, values.scope);
  await userEvent.click(screen.getByRole("button", { name: Label.SUBMIT }));
  expect(mock.history.patch[0].url).toBe(`/idps/${provider.id}`);
  expect(JSON.parse(mock.history.patch[0].data as string)).toMatchObject({
    apple_private_key: provider.apple_private_key,
    apple_private_key_id: provider.apple_private_key_id,
    apple_team_id: provider.apple_team_id,
    auth_url: provider.auth_url,
    client_id: provider.client_id,
    client_secret: provider.client_secret,
    id: provider.id,
    issuer_url: provider.issuer_url,
    mapper_url: provider.mapper_url,
    microsoft_tenant: provider.microsoft_tenant,
    provider: provider.provider,
    requested_claims: provider.requested_claims,
    subject_source: "userinfo",
    token_url: provider.token_url,
    scope: [values.scope],
  });
});

test("handles API success", async () => {
  let location: Location | null = null;
  const invalidateQueries = vi.fn();
  vi.spyOn(reactQuery, "useQueryClient").mockReturnValue({
    invalidateQueries,
  } as unknown as reactQuery.QueryClient);
  mock.onPatch(`/idps/${provider.id}`).reply(200);
  const values = {
    scope: faker.word.sample(),
  };
  renderComponent(
    <NotificationProvider>
      <ProviderEdit />
      <NotificationConsumer />
    </NotificationProvider>,
    {
      url: `/?id=${provider.id}`,
      setLocation: (newLocation) => {
        location = newLocation;
      },
    },
  );
  await userEvent.type(
    screen.getByRole("textbox", { name: ProviderFormLabel.SCOPES }),
    values.scope,
  );
  await userEvent.click(screen.getByRole("button", { name: Label.SUBMIT }));
  await waitFor(() =>
    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: [queryKeys.providers],
    }),
  );
  expect(document.querySelector(".p-notification--positive")).toHaveTextContent(
    Label.SUCCESS,
  );
  expect((location as Location | null)?.pathname).toBe("/");
  expect((location as Location | null)?.search).toBe("");
});

test("handles API failure", async () => {
  mock.onPatch(`/idps/${provider.id}`).reply(400, {
    message: "oops",
  });
  const values = {
    scope: faker.word.sample(),
  };
  renderComponent(
    <NotificationProvider>
      <ProviderEdit />
      <NotificationConsumer />
    </NotificationProvider>,
    { url: `/?id=${provider.id}` },
  );
  await userEvent.type(
    screen.getByRole("textbox", { name: ProviderFormLabel.SCOPES }),
    values.scope,
  );
  await userEvent.click(screen.getByRole("button", { name: Label.SUBMIT }));
  expect(document.querySelector(".p-notification--negative")).toHaveTextContent(
    `${Label.ERROR}oops`,
  );
});
