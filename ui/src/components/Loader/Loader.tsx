import { FC } from "react";
import { Spinner } from "@canonical/react-components";

import { LoaderTestId } from "./index";
import { testId } from "test/utils";

interface Props {
  text?: string;
}

const Loader: FC<Props> = ({ text = "Loading..." }) => {
  return (
    <div className="u-loader" {...testId(LoaderTestId.COMPONENT)}>
      <Spinner text={text} />
    </div>
  );
};

export default Loader;
