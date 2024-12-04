import { screen } from "@testing-library/dom";
import { faker } from "@faker-js/faker";
import userEvent from "@testing-library/user-event";
import { Formik } from "formik";
import MockAdapter from "axios-mock-adapter";

import { axiosInstance } from "api/axios";
import { renderComponent } from "test/utils";

import ProviderForm from "./ProviderForm";
import { ProviderFormTypes } from "./ProviderForm";
import { Label } from "./types";

const mock = new MockAdapter(axiosInstance);

const initialValues = {
  id: "",
  client_id: "",
  client_secret: "",
  auth_url: "",
  issuer_url: "",
  token_url: "",
  subject_source: "userinfo",
  microsoft_tenant: "",
  provider: "",
  mapper_url: "",
  scope: "",
  apple_team_id: "",
  apple_private_key_id: "",
  apple_private_key: "",
  requested_claims: "",
} as const;

beforeEach(() => {
  mock.reset();
});

test("submits field data", async () => {
  const values = {
    provider: "auth0",
    id: faker.word.sample(),
    client_id: faker.word.sample(),
    client_secret: faker.word.sample(),
    scope: faker.word.sample(),
    requested_claims: faker.word.sample(),
    mapper_url: faker.word.sample(),
  };
  const onSubmit = vi.fn();
  renderComponent(
    <Formik<ProviderFormTypes>
      initialValues={initialValues}
      onSubmit={onSubmit}
    >
      {(formik) => (
        <>
          <ProviderForm formik={formik} />
          <button onClick={() => void formik.submitForm()}>Submit</button>
        </>
      )}
    </Formik>,
  );
  await userEvent.selectOptions(
    screen.getByRole("combobox", { name: Label.PROVIDER }),
    values.provider,
  );
  await userEvent.type(
    screen.getByRole("textbox", { name: Label.NAME }),
    values.id,
  );
  await userEvent.type(
    screen.getByRole("textbox", { name: Label.CLIENT_ID }),
    values.client_id,
  );
  await userEvent.type(
    screen.getByRole("textbox", { name: Label.CLIENT_SECRET }),
    values.client_secret,
  );
  await userEvent.type(
    screen.getByRole("textbox", { name: Label.SCOPES }),
    values.scope,
  );
  await userEvent.type(
    screen.getByRole("textbox", { name: Label.REQUESTED_CLAIMS }),
    values.requested_claims,
  );
  await userEvent.type(
    screen.getByRole("textbox", { name: Label.MAPPER }),
    values.mapper_url,
  );
  await userEvent.click(screen.getByRole("button"));
  expect(onSubmit.mock.calls[0][0]).toMatchObject(values);
});

test("displays fields for apple form", async () => {
  renderComponent(
    <Formik<ProviderFormTypes> initialValues={initialValues} onSubmit={vi.fn()}>
      {(formik) => <ProviderForm formik={formik} />}
    </Formik>,
  );
  await userEvent.selectOptions(
    screen.getByRole("combobox", { name: Label.PROVIDER }),
    "apple",
  );
  expect(
    screen.queryByRole("textbox", { name: Label.CLIENT_SECRET }),
  ).not.toBeInTheDocument();
  expect(
    screen.getByRole("textbox", { name: Label.PRIVATE_KEY }),
  ).toBeInTheDocument();
  expect(
    screen.getByRole("textbox", { name: Label.DEVELOPER_TEAM_ID }),
  ).toBeInTheDocument();
  expect(
    screen.getByRole("textbox", { name: Label.PRIVATE_KEY_ID }),
  ).toBeInTheDocument();
});

test("displays fields for microsoft form", async () => {
  renderComponent(
    <Formik<ProviderFormTypes> initialValues={initialValues} onSubmit={vi.fn()}>
      {(formik) => <ProviderForm formik={formik} />}
    </Formik>,
  );
  await userEvent.selectOptions(
    screen.getByRole("combobox", { name: Label.PROVIDER }),
    "microsoft",
  );
  expect(
    screen.getByRole("textbox", { name: Label.TENANT }),
  ).toBeInTheDocument();
  expect(
    screen.getByRole("radio", { name: Label.USERINFO }),
  ).toBeInTheDocument();
  expect(screen.getByRole("radio", { name: Label.ME })).toBeInTheDocument();
});

test("generic form displays auto discovery fields", async () => {
  renderComponent(
    <Formik<ProviderFormTypes> initialValues={initialValues} onSubmit={vi.fn()}>
      {(formik) => <ProviderForm formik={formik} />}
    </Formik>,
  );
  await userEvent.selectOptions(
    screen.getByRole("combobox", { name: Label.PROVIDER }),
    "generic",
  );
  await userEvent.click(screen.getByRole("radio", { name: Label.YES }));
  expect(
    screen.getByRole("textbox", { name: Label.OIDC_SERVER_URL }),
  ).toBeInTheDocument();
  expect(
    screen.queryByRole("textbox", { name: Label.AUTH_URL }),
  ).not.toBeInTheDocument();
  expect(
    screen.queryByRole("textbox", { name: Label.TOKEN_URL }),
  ).not.toBeInTheDocument();
});

test("generic form displays non auto discovery fields", async () => {
  renderComponent(
    <Formik<ProviderFormTypes> initialValues={initialValues} onSubmit={vi.fn()}>
      {(formik) => <ProviderForm formik={formik} />}
    </Formik>,
  );
  await userEvent.selectOptions(
    screen.getByRole("combobox", { name: Label.PROVIDER }),
    "generic",
  );
  await userEvent.click(screen.getByRole("radio", { name: Label.NO }));
  expect(
    screen.queryByRole("textbox", { name: Label.OIDC_SERVER_URL }),
  ).not.toBeInTheDocument();
  expect(
    screen.getByRole("textbox", { name: Label.AUTH_URL }),
  ).toBeInTheDocument();
  expect(
    screen.getByRole("textbox", { name: Label.TOKEN_URL }),
  ).toBeInTheDocument();
});

test("disables the mapper URL field when editing", () => {
  renderComponent(
    <Formik<ProviderFormTypes> initialValues={initialValues} onSubmit={vi.fn()}>
      {(formik) => <ProviderForm isEdit formik={formik} />}
    </Formik>,
  );
  expect(screen.getByRole("textbox", { name: Label.MAPPER })).toBeDisabled();
});
