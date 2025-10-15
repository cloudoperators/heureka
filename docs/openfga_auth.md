# OpenFGA Authentication & Authorization

Heureka utilizes OpenFGA for authentication and authorization.
It works using an OpenFGA server loaded with a custom authorization model for Heureka's entities which the main Heureka app communicates to via API calls.

## Setup Details

The OpenFGA server is setup in the [docker-compose](/docker-compose.yaml) file with each service tied to the `openfga` profile.
It is setup using [API token auth](https://github.com/openfga/go-sdk?tab=readme-ov-file#api-token), with the `AUTHZ_FGA_API_TOKEN` environment variable being used as the api token.

The Heureka app's OpenFGA interface is implemented using the [openfga/go-sdk](https://github.com/openfga/go-sdk), and is in the /internal/openfga directory.
Upon starting up, Heureka checks if the `AUTHZ_FGA_API_URL` is empty or not to determine if auth should be setup or not.
If the variable is not set, a template `noauthz.go` implementation of the interface is used and no authn/authz is done.
If the variable is set, the full `authz.go` implementation of the interface is used and a client is created using the value of the variable as the url to the running OpenFGA server.

After creating the client, the app then checks if a store with a name equal to the `AUTHZ_FGA_STORE_NAME` env variable already exists, and creates one if it does not.
It then checks if a model has already been created in the store or not, and if one has not then it creates a new model using the file pointed to by the `AUTHZ_MODEL_FILE_PATH` env variable as a source.

## Interface

The interface consists of four main functions

- CheckPermission(p PermissionInput)
    - Checks if a given user has a given level of permission on a given resource (based on relation between user and resource)
- AddRelation(r RelationInput)
    - Adds a specified relation between a given user and a given resource
- RemoveRelationBulk (r []RelationInput)
    - Remove all relations that match any given RelationInput as filters
- RemoveRelation(r RelationInput)
    - Removes a single relation between a given user and a given resource (if such a relation exists)
- ListRelations(filters []RelationInput)
    - Returns a list of relations that match any given RelationInput as filters
- ListAccessibleResources(p PermissionInput)
    - Returns a list of all objects of a specified type that a given user has a given relation with
- GetCurrentUser() 
    - Placeholder function to be implemented for future user context functionality

PermissionInput and RelationInput are structs defined in the interface that contain all the parameters for the above functions.

For more info on how OpenFGA handles users, objects, and relations: https://openfga.dev/docs/concepts

## Handlers

Using the event handling system, openfga tuples are modified based on the event handled. Based on the [Auth Model](https://github.com/cloudoperators/heureka/pull/829/files) defined, the OpenFGA tuples are implemented as follows here. 

| Event                | Event Handler                        | Tuples Implemented                                                                                   |
|-------------------------|-------------------------------|--------------------------------------------------------------------------------------------------|
| AddOwnerToService       | OnAddOwnerToService           | add service - user                                                                               |
| RemoveOwnerFromService  | OnRemoveOwnerFromService      | remove service - user                                                                            |
| CreateService           | OnServiceCreateAuthz           | add role - service                                                                               |
| DeleteService           | OnServiceDeleteAuthz           | remove user - service<br>remove role - service<br>remove support_group - service<br>remove service - component_instance |
| UpdateComponentVersion  | OnComponentVersionUpdateAuthz  | update component_version - component                                                             |
| CreateComponentVersion  | OnComponentVersionCreateAuthz  | add role - component_version                                                                     |
| DeleteComponentVersion  | OnComponentVersionDeleteAuthz  | remove user - component_version<br>remove component_instance - component_version<br>remove role - component_version<br>remove component - component_version |
| UpdateIssueMatch        | OnIssueMatchUpdateAuthz        | update issue_match - component_instance                                                                 |
| CreateIssueMatch        | OnIssueMatchCreateAuthz        | add role - issue_match                                                                           |
| DeleteIssueMatch        | OnIssueMatchDeleteAuthz        | delete user - issue_match<br>delete issue_match - component_instance<br>delete role - issue_match |
| UpdateComponentInstance | OnComponentInstanceUpdateAuthz | update component_instance - component_version_id<br>update component_instance - service           |
| CreateComponentInstance | OnComponentInstanceCreateAuthz | add service - component_instance<br>add role - component_instance<br>add component_instance - component_version |
| DeleteComponentInstance | OnComponentInstanceDeleteAuthz | delete user - component_instance<br>delete service - component_instance<br>delete role - component_instance<br>delete component_instance - component_version<br>delete component_instance - issue_match |
| DeleteSupportGroup      | OnSupportGroupDeleteAuthz      | delete user - support_group<br>delete support_group - support_group<br>delete role - support_group<br>delete support_group - service |
| CreateSupportGroup      | OnSupportGroupCreateAuthz      | add role - support_group                                                                         |
| AddServicetoSupportGroup| OnAddServiceToSupportGroup     | add support_group - service                                                                      |
| RemoveServiceFromSupportGroup | OnRemoveServiceFromSupportGroup | remove support_group - service                                                              |
| AddUsertoSupportGroup   | OnAddUserToSupportGroup        | add support_group - user                                                                         |
| RemoveUserFromSupportGroup | OnRemoveUserFromSupportGroup | remove support_group - user                                                                  |
| CreateComponent         | OnComponentCreateAuthz         | add role - component                                                                             |
| DeleteComponent         | OnComponentDeleteAuthz         | delete user - component<br>delete component_version - component<br>delete role - component        |

## Usage

The following four environment variables must be set to use OpenFGA

- AUTHZ_FGA_API_URL
    - The URL to the running OpenFGA server
- AUTHZ_FGA_STORE_NAME
    - A name for Heureka's store within the OpenFGA server
- AUTHZ_MODEL_FILE_PATH
    - The file path to Heureka's authorization model definition (internal/openfga/model/model.fga)
- AUTHZ_FGA_API_TOKEN
    - An api token to be used for communication between the OpenFGA server and Heureka

With the above variables set, use the following command to run the OpenFGA server

```
make openfga-up
```

Then run Heureka and a connection will be made to the server.
To verify it ran correctly, go to http://localhost:3000/playground to view the OpenFGA playground and check that the store and model were created as expected.
