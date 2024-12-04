import { screen } from "@testing-library/dom";
import { faker } from "@faker-js/faker";
import userEvent from "@testing-library/user-event";
import { Formik } from "formik";

import IdentityForm from "./IdentityForm";
import { IdentityFormTypes } from "./IdentityForm";
import { Label } from "./types";
import MockAdapter from "axios-mock-adapter";
import { axiosInstance } from "api/axios";
import { mockSchema } from "test/mocks/schemas";
import { renderComponent } from "test/utils";

const mock = new MockAdapter(axiosInstance);

const initialValues = {
  email: "",
  schemaId: "",
};

beforeEach(() => {
  mock.reset();
});

test("submits field data", async () => {
  const schemas = [mockSchema(), mockSchema()];
  mock.onGet("/schemas?page_token=&page_size=50").reply(200, { data: schemas });
  const values = {
    email: faker.word.sample(),
    schemaId: schemas[0].id,
  };
  const onSubmit = vi.fn();
  renderComponent(
    <Formik<IdentityFormTypes>
      initialValues={initialValues}
      onSubmit={onSubmit}
    >
      {(formik) => (
        <>
          <IdentityForm formik={formik} />
          <button onClick={() => void formik.submitForm()}>Submit</button>
        </>
      )}
    </Formik>,
  );
  await userEvent.type(
    screen.getByRole("textbox", { name: Label.EMAIL }),
    values.email,
  );
  await screen.findByRole("option", { name: schemas[0].id });
  await userEvent.selectOptions(
    screen.getByRole("combobox", { name: Label.SCHEMA }),
    values.schemaId,
  );
  await userEvent.click(screen.getByRole("button"));
  expect(onSubmit.mock.calls[0][0]).toMatchObject(values);
});
