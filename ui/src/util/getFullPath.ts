import { basePath } from "./basePaths";
import { removeTrailingSlash } from "./removeTrailingSlash";

// Extract the path from the URL including query params, hash etc.
export const getFullPath = (url: string, removeBase = false) => {
  const path = url.match(/(?<!\/)\/(?!\/).+$/)?.[0];
  return removeBase
    ? path?.replace(new RegExp(`^${removeTrailingSlash(basePath)}`), "")
    : path;
};
