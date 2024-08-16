import { basePath } from "./basePaths";
import { getDomain } from "./getDomain";
import { getFullPath } from "./getFullPath";
import { removeTrailingSlash } from "util/removeTrailingSlash";

export const handleNext = () => {
  const next = new URLSearchParams(window.location.search).get("next");
  if (!next) {
    return;
  }
  const decoded = decodeURIComponent(next);
  const path = getFullPath(decoded);
  const domain = getDomain(decoded);
  if (!path) {
    return;
  }
  // Don't redirect if the 'next' param is the same as the current path.
  // Ignore redirects to external domains. This may be a malicious attempt to
  // redirect after login.
  // Ignore redirects that aren't from the UI basename.
  if (
    removeTrailingSlash(path) ===
      removeTrailingSlash(window.location.pathname) ||
    (domain && domain !== window.location.host) ||
    !path.startsWith(removeTrailingSlash(basePath))
  ) {
    // Remove the query string from the URL as we don't need to do anything with
    //the 'next' param.
    window.location.replace(window.location.pathname);
    return;
  }
  window.location.replace(path);
};
