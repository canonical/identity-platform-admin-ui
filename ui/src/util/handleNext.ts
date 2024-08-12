export const handleNext = () => {
  const next = new URLSearchParams(window.location.search).get("next");
  if (!next) {
    return;
  }
  const decoded = decodeURIComponent(next);
  // Don't redirect if the 'next' param is the same as the current path.
  if (
    decoded.replace(/\/$/, "") === window.location.pathname.replace(/\/$/, "")
  ) {
    // Remove the query string from the URL as we don't need to do anything with
    //the 'next' param.
    window.location.replace(window.location.pathname);
    return;
  }
  window.location.replace(decoded);
};
