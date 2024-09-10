## Known Issues

This file lists known issues with Litmus.

- **Authentication:** When the password contains special characters, the password might not be stored correctly in the secret manager, resulting in the API failing to start. As a workaround use a password that contains only alphanumeric characters and no special characters.
- **Analytics:** Currently, analytics for the worker service are not implemented.
