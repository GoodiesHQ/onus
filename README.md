# onus
A simple self-hosted todo application.

All users with the same organization email address will be able to assign tasks to one another.

## What is it?
Onus is a dead-simple To-Do task manager built for organizations. Authentication is done via OIDC and the domain name extracted from the email address will create the organization within Onus. The first account within an organization will become the "owner" and the owner can assign additional "admin" users once they sign up.

Once joined, any user can assign a task to any other user within their organization. Tasks are simple and hae the following options:

 - Title (required): A brief, high level description of the task.
 - Assignee (required): A user to whom you'd like to assign the task.

All other attributes are optional:

 - Description: An optional long description of, or additional context for, the task.
 - Priority: It will default to "Medium," but can be set to "Low" or "Medium" or "High" or "Urgent"
 - Target Due Date: An optional due date that will make the due date turn red when it is at or after this date.

## How to Run:

1) Download `docker-compose.yml`, `.env.example`, and `providers.env.example`
2) Modify `docker-compose.yml` as needed for your reverse proxy setup. A traefik example (commented) is included.
3) Rename `.env.example` -> `.env` and `providers.env.example` -> `providers.env`
4) Modify the values of `.env` as needed. Choose a database password, provide the public URL.
5) Modify the values of `providers.env` as needed. Create an OIDC-capable application and populate the file.
6) Run `docker compose up -d`


## Server configuration

Environment variables for the container.

| Variable | Purpose | Example |
| -------- | ------- | ------- |
| **ONUS_SERVER_HOST** | Host that the Onus service will listen on (in docker) | `0.0.0.0` (or empty) |
| **ONUS_SERVER_PORT** | Port that the Onus service will listen on (in docker) | `8888` |
| **ONUS_SERVER_URL**| Publicly accessible server URL | `https://onus.example.com` |
| **ONUS_SERVER_STATIC_DIR** | The directory that holds the Onus web application. | `/app/static` |
| **ONUS_DB_TYPE** | The database driver to use. Only "postgres" is currently supported. | `postgres` |
| **ONUS_DB_HOST** | The target hostname for the database host | `db` |
| **ONUS_DB_USER** | The username for the database account | `onus` |
| **ONUS_DB_PASS** | The password for the database account | `Passw0rd!` |
| **ONUS_DB_SSL** | Whether to use SSL for the database connection or not | `true` |
| **ONUS_AUTH_OIDC_ENABLED** | Whether the OIDC provider engine is enabled. OIDC is the only supported auth type, so this should always be true. | `true` |


#### Providers
Providers can be provided in environment variables through pattern matching.

| Variable | Purpose | Example |
| -------- | ------- | ------- |
| **ONUS_AUTH_OIDC_PROVIDER_\[ProviderName\]_ISSUER_URL** | The OIDC provider named "ProviderName" Issuer URL | `https://login.microsoftonline.com/fc69f5fc-e5c1-4954-8a21-282ddb2bbb44/v2.0` |
| **ONUS_AUTH_OIDC_PROVIDER_\[ProviderName\]_CLIENT_ID** | The OIDC provider named "ProviderName" Client ID | `435c2954-a4ad-4711-89f4-c25964de587e` |
| **ONUS_AUTH_OIDC_PROVIDER_\[ProviderName\]_CLIENT_SECRET** | The OIDC provider named "ProviderName" Client Secret | `YourClientSecret` |


**NOTE:** Once you have a provider name defined, you will need the callback URL for your OAuth provider application. The URL will be in the following pattern:

    /auth/[ProviderName]/callback

For example, if your provider name is "ENTRA" then the callback URL will be: `https://onus.example.com/auth/entra/callback`