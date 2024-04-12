import { useSearchParams } from "react-router-dom";

export const usePagination = () => {
  const [searchParams, setSearchParams] = useSearchParams();
  const pageToken = searchParams.get("page_token") ?? "";
  const setPageToken = (newPage: string) => {
    setSearchParams({ ...searchParams, page_token: newPage });
  };

  return { pageToken, setPageToken };
};
