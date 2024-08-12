export const getFullPath = () => location.href.match(/(?<!\/)\/(?!\/).+$/)?.[0];
