import axios from "axios";

import type { UserPrincipal } from "types/auth";
import { ErrorResponse } from "types/api";
import { AxiosError } from "axios";

const BASE = "auth";

export const authURLs = {
  login: BASE,
  me: `${BASE}/me`,
};

export const fetchMe = (): Promise<UserPrincipal | null> => {
  return new Promise((resolve, reject) => {
    axios
      .get<UserPrincipal>(authURLs.me)
      .then(({ data }) => resolve(data))
      .catch(({ response }: AxiosError<ErrorResponse>) => {
        // If the user is not authenticated then return null instead of throwing an
        // error. This is necessary so that a login screen can be displayed instead of displaying
        // the error.
        if (response?.status && [401, 403].includes(response.status)) {
          resolve(null);
        }
        return reject(response?.data?.error ?? response?.data?.message);
      });
  });
};
