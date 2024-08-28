import { urls as rebacURLs } from "@canonical/rebac-admin";

const rebacAdminURLS = rebacURLs("/");

export const urls = {
  clients: {
    index: "/client",
  },
  groups: rebacAdminURLS.groups,
  identities: {
    index: "/identity",
  },
  index: "/",
  providers: {
    index: "/provider",
  },
  roles: rebacAdminURLS.roles,
  schemas: {
    index: "/schema",
  },
};
