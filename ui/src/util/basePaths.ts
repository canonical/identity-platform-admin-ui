// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

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
export const apiBasePath =
  `${basePath.replace(/ui\/$/, "")}api/v0/` as BasePath;

export const appendBasePath = (path: string) =>
  `${removeTrailingSlash(basePath)}/${path.replace(/^\//, "")}`;

export const appendAPIBasePath = (path: string) =>
  `${removeTrailingSlash(apiBasePath)}/${path.replace(/^\//, "")}`;
