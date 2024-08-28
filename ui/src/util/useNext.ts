import { removeTrailingSlash } from "util/removeTrailingSlash";
import { keyToPath } from "./keyToPath";
import { useLocation, useNavigate, useSearchParams } from "react-router-dom";
import { useEffect } from "react";

export const useNext = () => {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const location = useLocation();

  useEffect(() => {
    const next = searchParams.get("next");
    if (!next) {
      return;
    }
    const path = keyToPath(next);
    if (
      // Don't redirect if a path matching the next key wasn't found.
      !path ||
      // Don't redirect if the 'next' param is the same as the current path.
      removeTrailingSlash(path) === removeTrailingSlash(location.pathname)
    ) {
      // Remove the query string from the URL as we don't need to do anything with
      //the 'next' param.
      searchParams.delete("next");
      setSearchParams(searchParams);
      return;
    }
    navigate(path, { replace: true });
  }, [location, searchParams]);
};
