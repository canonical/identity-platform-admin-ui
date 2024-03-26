export const getParentsBottomSpacing = (element: Element): number => {
  let sum = 0;
  while (element.parentElement) {
    element = element.parentElement;
    const style = window.getComputedStyle(element);
    const margin = parseInt(style.marginBottom);
    const padding = parseInt(style.paddingBottom);
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
  const margin = parseFloat(style.marginTop) + parseFloat(style.marginBottom);
  const padding =
    parseFloat(style.paddingTop) + parseFloat(style.paddingBottom);
  return element.offsetHeight + margin + padding + 1;
};
