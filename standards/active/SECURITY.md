<!-- Generated from registry/standards/security/no-secrets@1.1.0. Update the registry standard, then regenerate. -->

# No Secrets Security Standard

This standard keeps generated projects safe for public repositories.

## May 2026 Baseline

1. OWASP ASVS 5 and OWASP API Security Top 10 2023 are the security reference baseline.
2. SLSA v1.2 concepts guide supply-chain maturity.
3. Public starter repos must be safe before first push.
4. Generated projects must default to placeholder values only.
5. CI should fail on likely secret patterns where the project enables the check.

## Rules

1. Real secrets are never committed.
2. `.env` files are ignored.
3. Only `.env.example` files may be committed.
4. Tokens, private keys, passwords and production hostnames must not appear in templates.
5. Generated projects must include a hygiene check before publish.
6. Incident notes must explain how to rotate leaked credentials.
7. Default auth/session/payment examples must be non-production placeholders.
8. Public docs must not include private IPs, private hostnames or personal server paths.

## Secret Shapes

The hygiene check must scan for common secret shapes:

1. private key headers;
2. GitHub PATs;
3. OpenAI/API-style `sk-` keys;
4. AWS access keys;
5. Google API keys;
6. Slack tokens;
7. obvious `client_secret` or `api_key` assignments.
8. Stripe secret keys;
9. JWT signing secrets;
10. database URLs with real passwords.

## Template Hygiene

Templates must not include:

1. private company names;
2. private hostnames;
3. personal server paths;
4. production URLs;
5. real account IDs;
6. dependency vendor folders.
7. generated lockfiles that were created with private registries unless intentionally documented.

## Dependency And Supply Chain

1. Prefer lockfiles for runnable generated projects.
2. Use official registries or document private registry setup outside public templates.
3. Pin CI actions by version or SHA according to project maturity.
4. Use OIDC federation for cloud deploys when possible.
5. Keep SBOM/provenance optional until real release automation exists.
6. Do not claim SLSA compliance without provenance-producing CI.

## API Security Basics

Generated APIs should plan for:

1. object-level authorization;
2. function-level authorization;
3. object property authorization;
4. resource consumption limits;
5. SSRF protection when accepting URLs;
6. safe third-party API consumption;
7. endpoint inventory.

## Logging And Redaction

1. Secrets are redacted before logs leave the process.
2. Tokens are logged only as hashes or last-four style debug hints.
3. Payment, OAuth and Telegram signed payloads are not logged raw.
4. Error responses do not reveal stack traces.

## Incident Response

If a real credential is found:

1. stop publishing;
2. revoke or rotate the credential;
3. remove it from the working tree;
4. rewrite public history only with explicit owner approval;
5. document the incident and prevention check.
6. add a regression check if the leak shape was not detected.

## Enforcement Reality

Secret-shape and private-residue scanning is linted and CI-blocking through `check_template_hygiene` and `vibe verify`.

Broader ASVS, SLSA, dependency and runtime security practices are documented until a project adds real security scanning and release automation.
