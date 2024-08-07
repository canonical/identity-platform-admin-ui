// https://github.com/canonical/identity-platform-admin-ui/blob/7781bcb13d2ea71c0b898c8ff47131334fd7df9e/pkg/authentication/principal.go#L12
export type UserPrincipal = {
  email: string;
  name: string;
  nonce: string;
  sid: string;
  sub: string;
};
