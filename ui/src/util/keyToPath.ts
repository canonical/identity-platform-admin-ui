import { urls } from "urls";
import { URLS } from "./getURLKey";
import { ValueOf } from "@canonical/react-components";

export const keyToPath = (path: string) => {
  // Restrict the key to safe values.
  if (!path.match(/^[a-z|.]+$/)) {
    return null;
  }
  let currentSection = urls as ValueOf<URLS>;
  const keys = path.split(".");
  for (const [index, section] of keys.entries()) {
    const isLast = index === keys.length - 1;
    if (typeof currentSection === "object") {
      if (section in currentSection) {
        currentSection = currentSection[section];
      } else if (isLast) {
        // If this is the last item and the object doesn't contain the key then
        // exit and return null.
        break;
      }
    }
    // If this is the last item in the key path then try to get the URL.
    if (isLast) {
      // The trailing .index is removed from the key path so if the path leads
      // to a object then get the index.
      if (typeof currentSection === "object" && "index" in currentSection) {
        currentSection = currentSection.index;
      }
      if (typeof currentSection === "string") {
        return currentSection;
      } else if (typeof currentSection === "function") {
        return currentSection(null);
      }
      break;
    }
  }
  return null;
};
