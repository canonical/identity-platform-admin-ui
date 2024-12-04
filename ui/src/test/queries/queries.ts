import { queries } from "@testing-library/react";

import * as tableQueries from "./tables";

const customQueries = {
  ...queries,
  ...tableQueries,
};

export default customQueries;
