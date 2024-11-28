import React, { FC } from "react";
import { PaginatedResponse } from "types/api";
import { Button, Icon } from "@canonical/react-components";
import { usePagination } from "util/usePagination";
import { Label } from "./types";

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
        <Button onClick={() => setPageToken("")} title={Label.FIRST_PAGE}>
          <Icon name="chevron-left" />
          <Icon name="chevron-left" />
        </Button>
      )}
      {prev && (
        <Button onClick={() => setPageToken(prev)} title={Label.PREVIOUS_PAGE}>
          <Icon name="chevron-left" />
        </Button>
      )}
      {next && (
        <Button onClick={() => setPageToken(next)} title={Label.NEXT_PAGE}>
          <Icon name="chevron-right" />
        </Button>
      )}
    </>
  );
};

export default Pagination;
