export const getParentsBottomSpacing = (element: Element): number => {
  let sum = 0;
  while (element.parentElement) {
    element = element.parentElement;
    const style = window.getComputedStyle(element);
    const margin = parseInt(style.marginBottom) || 0;
    const padding = parseInt(style.paddingBottom) || 0;
    sum += margin + padding;
  }
  return sum;
};

export const getAbsoluteHeightBelow = (belowId: string): number => {
  const element = belowId ? document.getElementById(belowId) : undefined;
  if (!element) {
    return 0;
  }
  const style = window.getComputedStyle(element);
  const margin =
    parseFloat(style.marginTop) + parseFloat(style.marginBottom) || 0;
  const padding =
    parseFloat(style.paddingTop) + parseFloat(style.paddingBottom) || 0;
  return element.offsetHeight + margin + padding + 1;
};
