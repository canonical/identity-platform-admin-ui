import { removeTrailingSlash } from "util/removeTrailingSlash";
type BasePath = `/${string}`;

declare global {
  interface Window {
    base?: string;
  }
}

export const calculateBasePath = (): BasePath => {
  let basePath = "";
  if ("base" in window && typeof window.base === "string") {
    basePath = window.base;
  }
  if (basePath) {
    return `${removeTrailingSlash(basePath)}/` as BasePath;
  }
  return "/";
};

export const basePath: BasePath = `${calculateBasePath()}ui`;
export const apiBasePath: BasePath = `${calculateBasePath()}api/v0/`;

export const appendBasePath = (path: string) =>
  `${removeTrailingSlash(basePath)}/${path.replace(/^\//, "")}`;

export const appendAPIBasePath = (path: string) =>
  `${removeTrailingSlash(apiBasePath)}/${path.replace(/^\//, "")}`;
