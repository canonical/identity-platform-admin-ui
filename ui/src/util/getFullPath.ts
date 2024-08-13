// Extract the path from the URL including query params, hash etc.
export const getFullPath = (url: string) =>
  url.match(/(?<!\/)\/(?!\/).+$/)?.[0];
