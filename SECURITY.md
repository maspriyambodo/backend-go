# Security Policy

This document describes the security policy for the **AdminBE** backend API built with Go and Gin, including how to report vulnerabilities and recommended hardening guidelines for production deployments.

---

## Supported Versions

Security fixes are generally applied to the latest code in the `master` branch and the most recent tagged release.

| Version / Branch | Status          |
|------------------|-----------------|
| `master`         | Actively supported |
| Latest tag       | Actively supported |
| Older tags       | Not guaranteed to receive security fixes |

If you are running an older tagged version, you are encouraged to regularly pull the latest changes from `master` or upgrade to the latest tagged release.

---

## Reporting a Vulnerability

If you discover a security vulnerability in this project, please follow these steps:

1. **Do not** create a public GitHub issue describing the vulnerability in detail.
2. Contact the maintainer privately using one of the following options:
   - GitHub Security Advisories (preferred):  
     - Go to the repository’s **Security** tab  
     - Click **“Report a vulnerability”**
   - Alternatively, contact the repository owner via the email listed on their GitHub profile or another private channel you may already use.

When reporting, please include (as far as possible):

- A description of the vulnerability and potential impact
- Steps to reproduce the issue
- Any proof-of-concept code or requests
- Affected configuration or environment details (e.g. specific `GIN_MODE`, reverse proxy, etc.)

You will receive an acknowledgement of your report and, if confirmed, we will work on a fix and coordinate a responsible disclosure timeline when appropriate.

---

## Handling Secrets and Sensitive Data

This project uses environment variables and/or configuration files such as `.env` and `configs/config.yaml` for sensitive information:

- Database credentials (`DB_USER`, `DB_PASSWORD`, `DB_HOST`, `DB_NAME`, etc.)
- Redis configuration (`REDIS_HOST`, `REDIS_PASSWORD`, etc.)
- JWT settings (`JWT_SECRET`, `JWT_EXPIRATION`)
- JasperServer credentials (`JASPER_USERNAME`, `JASPER_PASSWORD`, `JASPER_BASE_URL`)
- Other service credentials or tokens

**Guidelines:**

- Never commit `.env`, `.env.prod`, `.env.production` or any file containing real secrets to version control.
- Use strong, unique values for `JWT_SECRET` (at least 32 characters).  
  - You can generate one with `go run ./cmd/secret`.
- Use different secrets and credentials for development, staging, and production.
- Store secrets in a secure secret manager (e.g. environment variables, Vault, SSM, etc.) instead of plain-text files where possible.
- Rotate secrets periodically and immediately after any suspected compromise.

---

## Default Credentials

The project README may provide **sample** login credentials such as:

- `admin` / `admin123`  
- JasperServer default credentials

These are **for development/testing only** and must **never** be used in production.

In production environments:

- Change all default usernames and passwords before deployment.
- Enforce strong password policies for all administrative accounts.
- Restrict admin routes and dashboards to trusted networks/users wherever possible.

---

## Authentication and Authorization

The backend uses JWT-based authentication and role-based access control (RBAC), including:

- Users
- Roles
- Role inheritances
- User–Role, Role–Menu, and User–Menu associations

**Recommendations:**

- Always send JWT tokens via the `Authorization: Bearer <token>` header over **HTTPS** only.
- Set reasonable token expiration using `JWT_EXPIRATION` and refresh tokens if needed.
- Revoke tokens as part of user deactivation or password reset processes.
- Restrict high-privilege roles (e.g. system administrators) to a minimal set of accounts.
- Regularly review role and permission assignments to ensure least privilege.

---

## Transport Security (HTTPS / TLS)

For any non-local environment (staging, production, etc.):

- Terminate all HTTP traffic behind a reverse proxy such as Nginx or a cloud load balancer that enforces **HTTPS**.
- Redirect all HTTP requests to HTTPS.
- Use modern TLS configuration and up-to-date certificates (e.g. from Let’s Encrypt).
- Avoid exposing the Gin server directly to the public internet without a secure, hardened proxy layer.

---

## Database and Redis Security

This project uses MySQL/MariaDB and Redis (optionally) as dependencies.

**Recommendations:**

- Do not expose MySQL or Redis directly to the public internet.
- Bind database and Redis ports to private networks only (e.g. `127.0.0.1` or internal Docker networks).
- Use strong passwords for all database and Redis users.
- Grant the application user only the minimal required privileges on the database.
- Regularly back up database data and verify restore procedures.
- Keep MySQL/MariaDB and Redis versions up to date with security patches.

---

## JasperReports / JasperServer Security

If you enable JasperServer integration:

- Do not use default JasperServer credentials in production.
- Restrict access to JasperServer with appropriate authentication and network rules.
- Ensure that report paths and parameters are validated to prevent injection or unauthorized access to sensitive reports.
- Treat any report output that includes sensitive data as confidential and protect it accordingly.

---

## CORS, Rate Limiting, and API Hardening

- Configure `cors.allow_origins` to a strict, known list of allowed front-end origins in production instead of `"*"`.
- Consider implementing:
  - Rate limiting or throttling to mitigate brute-force login attempts and abuse.
  - IP allowlists/denylists for sensitive endpoints (e.g. admin operations).
  - Additional application-level checks on critical operations (e.g. delete, update with high impact).

---

## Logging and Audit Logs

The project provides audit logging endpoints and general API logging.

Security-related guidelines:

- Avoid logging full credentials, JWT tokens, or other secrets.
- Be cautious when logging request bodies that may contain PII or confidential information.
- Protect log files with appropriate access controls.
- Rotate logs regularly and store them securely.
- Use audit logs to monitor security-relevant events (logins, role changes, permission modifications, etc.).

---

## Dependency and Build Security

- Keep Go, Gin, and all third-party modules up to date with security patches:
  - Run `go list -m -u all` or similar tools regularly.
  - Consider using `govulncheck` or other scanners.
- Use reproducible builds and pinned versions in `go.mod`.
- If you use Docker:
  - Start from a minimal, supported Go base image.
  - Regularly rebuild images to include upstream security patches.
  - Avoid running the application as `root` inside containers where possible.

---

## Responsible Disclosure

We ask all users and researchers to practice responsible disclosure:

- Give the maintainer a reasonable amount of time to investigate and patch the issue before any public disclosure.
- Avoid exploiting vulnerabilities beyond what is necessary to demonstrate the issue.
- Do not attempt to access, modify, or delete data that does not belong to you.

Thank you for helping keep this project and its users secure.
