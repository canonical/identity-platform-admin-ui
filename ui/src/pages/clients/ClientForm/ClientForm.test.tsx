import { screen } from "@testing-library/dom";
import { render } from "@testing-library/react";
import { faker } from "@faker-js/faker";
import { Formik } from "formik";
import userEvent from "@testing-library/user-event";

import ClientForm from "./ClientForm";
import { ClientFormTypes } from "./ClientForm";
import { Label } from "./types";

const initialValues = {
  client_uri: "",
  client_name: "",
  grant_types: [],
  response_types: [],
  scope: "",
  redirect_uris: [],
  request_object_signing_alg: "",
};

test("submits field data", async () => {
  const values = {
    client_uri: faker.word.sample(),
    client_name: faker.word.sample(),
    scope: faker.word.sample(),
    redirect_uris: [faker.word.sample()],
    request_object_signing_alg: faker.word.sample(),
  };
  const onSubmit = vi.fn();
  render(
    <Formik<ClientFormTypes> initialValues={initialValues} onSubmit={onSubmit}>
      {(formik) => (
        <>
          <ClientForm formik={formik} />
          <button onClick={() => void formik.submitForm()}>Submit</button>
        </>
      )}
    </Formik>,
  );
  await userEvent.type(
    screen.getByRole("textbox", { name: Label.NAME }),
    values.client_name,
  );
  await userEvent.type(
    screen.getByRole("textbox", { name: Label.SCOPE }),
    values.scope,
  );
  await userEvent.type(
    screen.getByRole("textbox", { name: Label.REDIRECT_URI }),
    values.redirect_uris[0],
  );
  await userEvent.type(
    screen.getByRole("textbox", { name: Label.CLIENT_URI }),
    values.client_uri,
  );
  await userEvent.type(
    screen.getByRole("textbox", { name: Label.SIGNING_ALGORITHM }),
    values.request_object_signing_alg,
  );
  await userEvent.click(
    screen.getByRole("checkbox", { name: "authorization_code" }),
  );
  await userEvent.click(screen.getByRole("checkbox", { name: "id_token" }));
  await userEvent.click(screen.getByRole("button"));
  expect(onSubmit.mock.calls[0][0]).toMatchObject({
    ...values,
    grant_types: ["authorization_code"],
    response_types: ["id_token"],
  });
});
