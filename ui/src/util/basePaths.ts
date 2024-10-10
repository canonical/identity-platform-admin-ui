import { removeTrailingSlash } from "util/removeTrailingSlash";
import { getFullPath } from "./getFullPath";
type BasePath = `/${string}`;

export const calculateBasePath = (): BasePath => {
  const basePath = document.querySelector("base")?.href;
  const path = basePath ? getFullPath(basePath) : null;
  if (path) {
    return `${removeTrailingSlash(path)}/` as BasePath;
  }
  return "/ui/";
};

export const basePath: BasePath = calculateBasePath();
export const apiBasePath: BasePath = "/api/v0/";

export const appendBasePath = (path: string) =>
  `${removeTrailingSlash(basePath)}/${path.replace(/^\//, "")}`;

export const appendAPIBasePath = (path: string) =>
  `${removeTrailingSlash(apiBasePath)}/${path.replace(/^\//, "")}`;
