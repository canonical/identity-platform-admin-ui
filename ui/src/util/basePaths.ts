type BasePath = `/${string}`;

export const calculateBasePath = (): BasePath => {
  const path = window.location.pathname;
  // find first occurrence of /ui/ and return the string before it
  const basePath = path.match(/(.*\/ui\/)/);
  if (basePath) {
    return basePath[0] as BasePath;
  }
  return "/";
};

export const basePath: BasePath = calculateBasePath();
export const apiBasePath: BasePath = `${basePath}../api/v0/`;

export const appendBasePath = (path: string) =>
  `${basePath.replace(/\/$/, "")}/${path.replace(/^\//, "")}`;

export const appendAPIBasePath = (path: string) =>
  `${apiBasePath.replace(/\/$/, "")}/${path.replace(/^\//, "")}`;
