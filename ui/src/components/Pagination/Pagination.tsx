import React, { FC } from "react";
import { PaginatedResponse } from "types/api";
import { Button } from "@canonical/react-components";
import { usePagination } from "util/usePagination";
import IconLeft from "components/IconLeft";
import IconRight from "components/IconRight";

interface Props {
  response?: PaginatedResponse<unknown[]>;
}

const Pagination: FC<Props> = ({ response }) => {
  const { pageToken, setPageToken } = usePagination();
  const showFirstLink = pageToken !== "";
  const isMissingNext = !response || !response._meta.next;
  const isEmptyPage = response?.data.length === 0;

  if (!showFirstLink && (isMissingNext || isEmptyPage)) {
    return null;
  }

  const next = response?._meta.next ?? "";
  const prev = response?._meta.prev ?? "";

  return (
    <>
      {showFirstLink && (
        <Button onClick={() => setPageToken("")} title="First page">
          <IconLeft />
          <IconLeft />
        </Button>
      )}
      {prev && (
        <Button onClick={() => setPageToken(prev)} title="Previous page">
          <IconLeft />
        </Button>
      )}
      {next && (
        <Button onClick={() => setPageToken(next)} title="Next page">
          <IconRight />
        </Button>
      )}
    </>
  );
};

export default Pagination;
