import { useSearchParams } from "react-router";

export const usePagination = () => {
  const [searchParams, setSearchParams] = useSearchParams();
  const pageToken = searchParams.get("page_token") ?? "";
  const setPageToken = (newPage: string) => {
    setSearchParams({
      ...Object.fromEntries(searchParams),
      page_token: newPage,
    });
  };

  return { pageToken, setPageToken };
};
