import { getDomain } from "./getDomain";
import { getFullPath } from "./getFullPath";
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
  if (
    path.replace(/\/$/, "") === window.location.pathname.replace(/\/$/, "") ||
    (domain && domain !== window.location.host)
  ) {
    // Remove the query string from the URL as we don't need to do anything with
    //the 'next' param.
    window.location.replace(window.location.pathname);
    return;
  }
  window.location.replace(path);
};
