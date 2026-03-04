# Database users

## Initialization problem

Our current database schema mandates that every entity must have a `CreatedBy` and `UpdatedBy` user associated with it. As a result, it has been decided that a `systemuser` must be present during the database initialization process. This `systemuser` should have an ID set to `1`, what is currently established in initialization sql schema file. Without the `systemuser`, no new user can be created, as the creation of new users would fail due to the `CreatedBy` field requiring a valid user (i.e., one that already exists in the database).


## User context and identification

The Heureka API operations are designed to be used by either a scanner or a UI user. Two distinct authentication methods are provided for these use cases.


### Scanner Authentication

This method relies on a custom JWT token. To enable scanner authentication, the `AUTH_TOKEN_SECRET` variable must be configured. This secret must be shared with the token generator. The user associated with the scanner stored in context is identified in the JWT token under the `subject` claim.


### User Authentication

This method utilizes an external OIDC (OpenID Connect) provider for authentication of UI users. To enable OIDC-based authentication, you must configure the following:

- `AUTH_OIDC_URL`: The URL of a valid OIDC provider.
- `AUTH_OIDC_CLIENT_ID`: The client ID used for authentication.

The user is identified from the `sub` claim of the OIDC token.
