# GasTrack Security Architecture & Threat Model

This document outlines the security architecture, threat model, and mitigations applied in the GasTrack application, satisfying the OWASP Top 10 mitigation requirements.

## OWASP Top 10 Mitigations

### A01: Broken Access Control (GT-022)
- **Mitigation:** Hierarchical REST routing combined with explicit ownership checks (`IsVehicleOwner`) prevents horizontal privilege escalation. All vehicle-specific data (maintenance, history, odometer, expense) strictly validates the authenticated user's ownership of the vehicle.

### A02: Cryptographic Failures (GT-023)
- **Mitigation:** User passwords are securely hashed using `bcrypt` with a default cost factor. Passwords are never stored or logged in plaintext. Transport encryption (HTTPS/TLS) is mandated for production environments. JWT tokens use strong HMAC-SHA256 signatures.

### A03: Injection (GT-024)
- **Mitigation:** All PostgreSQL queries use parameterized execution via the `pgx` library. There is no string concatenation used for SQL queries, rendering SQL injection impossible.

### A04: Insecure Design (GT-025)
- **Threat Model:** 
  - *Threat:* Unauthorized data access. *Mitigation:* Strict Ownership checks.
  - *Threat:* Account takeover. *Mitigation:* Secure session tokens and password hashing.
  - *Threat:* Malicious payloads. *Mitigation:* Server-side validation regardless of client input.

### A05: Security Misconfiguration (GT-026)
- **Mitigation:** Production builds disable verbose error outputs and stack traces from leaking in HTTP responses. Environment variables securely manage secrets without exposing them in source code.

### A06: Vulnerable and Outdated Components (GT-027)
- **Mitigation:** `go.mod` and `go.sum` are utilized to lock dependencies, ensuring reproducible builds. Regular audits should be performed prior to releases.

### A07: Identification and Authentication Failures (GT-028)
- **Mitigation:** Sensitive credentials are not logged. (Rate limiting and session invalidation will be implemented as part of specific feature updates).

### A08: Software and Data Integrity Failures (GT-029)
- **Mitigation:** Database migrations are strictly versioned using `goose`. Server-side validation handles data integrity irrespective of client checks.

### A09: Security Logging and Monitoring Failures (GT-030)
- **Mitigation:** (Logging for authentication and authorization failures will be systematically added in specific feature updates).

### A10: Server-Side Request Forgery (SSRF) (GT-031)
- **Mitigation:** The application does not fetch arbitrary external URLs based on user input. Only predefined internal or approved endpoints are reachable by the backend services.
