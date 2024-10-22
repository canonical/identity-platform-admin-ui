# Security Overview in Identity Platform Admin UI

The Identity Platform Admin UI is a tool designed to manage user identities,
permissions, and various configurations within the broader scope of the Identity
Platform. As security is a primary concern in identity management, the tool
includes built-in protection mechanisms and best practices to ensure secure
interaction with sensitive data and services.

This document provides an overview of security concerns related to the Identity
Platform Admin UI, highlighting common risks and outlining best practices to
mitigate them.

## Common Risks

Similar to other web-based services, the Identity Platform Admin UI faces the
following typical potential security risk challenges.

- **Authentication and Authorization Vulnerabilities**: Proper handling of
  authentication tokens, session management, and role-based access control
  (RBAC) are critical.
- **Injection Attacks**: Inputs provided by users, especially those directly
  interacting with backend services, must be sanitised to avoid injection
  vulnerabilities such as SQL injection or command injection.
- **Sensitive Data Exposure**: Personally identifiable information (PII), access
  tokens, and passwords should be handled with encryption and strong access
  controls to avoid leaks.
- **Cross-Site Scripting (XSS) and Cross-Site Request Forgery (CSRF)**: Frontend
  security is a concern to avoid XSS, CSRF, and other web-related
  vulnerabilities.

## Built-in Protection Mechanisms

The Identity Platform Admin UI offers several built-in security features to
protect from potential security issues.

### Authentication

The service leverages Ory Hydra for OAuth2-based authentication and Ory Kratos
for identity management, creating a robust system for managing authentication
securely. This integration ensures that sensitive user data are managed through
industry-standard protocols and best practices. More information can be found
[here](https://www.ory.sh/docs/hydra/security-architecture).

### Fine-Grained Authorization

Permissions within the service are restricted through OpenFGA integration,
ensuring that access is granted only to authorized users.

### Encrypted Communication

All communications between the UI and all backend services are expected to use
SSL/TLS encryption to protect against man-in-the-middle attacks.

### Web Application Security

The service implements comprehensive web application security measures by
adhering to industry standards. This includes input validation and sanitization,
secure cookie management, and browser-based defences. These measures are
designed to mitigate common web vulnerabilities, such as cross-site scripting
(XSS) and other attack vectors that target web applications.

## Cryptography

### Packages Providing Cryptographic Functionality

- [Go Cryptography](https://pkg.go.dev/golang.org/x/crypto@v0.28.0)

### Cryptography Technology Used by Identity Platform Admin UI

The Identity Platform Admin UI utilizes
Golang’s [crypto package](https://pkg.go.dev/golang.org/x/crypto@v0.28.0) to
encrypt and decrypt authentication cookies, employing the `AES-GCM` symmetric
encryption algorithm. This ensures both the confidentiality and integrity of the
cookies, providing robust protection against tampering and unauthorised access.

### Cryptography Technology Exposed to the User

The Identity Platform Admin UI allows the user to provide a 32-bit secret key
for encrypting and decrypting cookies. This key should be supplied through an
environment variable and securely managed using appropriate management practices
to ensure the protection of sensitive data.

## Best Practices

The service comes with multiple security features by default, but implementing
best practices can further safeguard against new threats. Here are a few
important suggestions.

### Keep Service Updated

Regularly update the service and its dependencies to address known security
vulnerabilities. Canonical provides frequent security patches and updates to
ensure systems remain secure.

### Restrict Privileged Access

Follow the principle of least privilege by giving users only the level of access
necessary for their tasks, minimizing potential security risks.

### Rate Limiting

Combine both IP-based and user-based rate limits to defend against attackers
using multiple IP addresses or shared networks.

### Enable Secret Management

Using dedicated secret management solutions helps automate the process of
storing, rotating, and accessing secrets securely. Limit the access to the
secrets by enforcing strict access control policies. In addition, periodically
rotating secrets reduces the risk of long-term exposure in the event of a leak.
For example, rotating the cookie secret key helps mitigate the risk in
`AES-GCM`, reducing the chance of attackers exploiting such vulnerabilities.

### Enable Data Encryption

Use encryption to protect data both during transmission (via SSL/TLS) and
while it is stored, ensuring sensitive information remains secure.

### Set Up Monitoring and Alerts

Constantly monitor service logs for unusual activities, and configure alerts for
critical events like failed login attempts, privilege escalation, or
configuration changes to quickly address potential threats.
