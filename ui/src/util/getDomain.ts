// Extract the domain from the URL.
export const getDomain = (url: string) => url.match(/(?<=.+:\/\/)[^/]+/)?.[0];
