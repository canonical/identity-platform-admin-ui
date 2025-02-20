import { matchPath } from "react-router";
import { urls } from "urls";

export type URLS = {
  [key: string]:
    | string
    | URLS
    | ((args: unknown, relativeTo?: string) => string);
};

const findPath = (
  sections: URLS,
  pathname: string,
  keyPath = "",
): string | null => {
  const entries = Object.entries(sections);
  let path: string | null = null;
  for (const entry of entries) {
    const [key, section] = entry;
    const thisPath = [keyPath, key].filter(Boolean).join(".");
    if (typeof section === "string" && section === pathname) {
      path = thisPath;
      break;
    } else if (
      typeof section === "function" &&
      !!matchPath(section(null), pathname)
    ) {
      path = thisPath;
      break;
    } else if (typeof section === "object") {
      const matchingPath = findPath(section, pathname, thisPath);
      if (matchingPath) {
        path = matchingPath;
        break;
      }
    }
  }
  // Don't expose the index key. The index is handled when the path is returned
  // in the useNext hook.
  return path ? path.replace(/\.index$/, "") : null;
};

export const getURLKey = (pathname: string) => findPath(urls as URLS, pathname);
