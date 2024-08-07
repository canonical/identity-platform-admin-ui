import { apiBasePath } from "util/basePaths";
import type { UserPrincipal } from "types/auth";
import { handleResponse } from "util/api";

const BASE = `${apiBasePath}auth`;

export const authURLs = {
  login: BASE,
  me: `${BASE}/me`,
};

export const fetchMe = (): Promise<UserPrincipal> => {
  return new Promise((resolve, reject) => {
    fetch(authURLs.me)
      .then((response: Response) =>
        // If the user is not authenticated then return null instead of throwing an
        // error. This is necessary so that a login screen can be displayed instead of displaying
        // the error.
        [401, 403].includes(response.status) ? null : handleResponse(response),
      )
      .then((result: UserPrincipal) => resolve(result))
      .catch(reject);
  });
};
