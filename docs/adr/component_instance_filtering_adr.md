# Identification of Component Instances without support group or service

## Context and Problem Statement

Our current database schema enforces having a service id for each component instance, which could potentially present problems if we decide to change this requirement in the future. The relationship between component instances, services, and support groups is complex, with a many-to-many relationship between service and support group. This setup may lead to challenges in managing and maintaining the data.

## Decision Drivers

* Reasonable impelementation effort
* Performance is reasonable
* Filtering for UI is reasonable

## Considered Options Support Group

* **Don't change anything:**
    We can check in the SupportGroupService table if there is at least one entry for the service (SELECT service_id FROM SupportGroupService WHERE service_id = ?). If the row count is 0 we know that there is no support group. This would require implementing a custom filter for this scenario. 

* **Default Support Group Unknown**
    When no support group is linked to the service we create a relationship to the "Unknown" Support Group. A "dummy" Support Group that exist for this purpose. This require additional checks for managing Support Groups. We have a "Unknown" Support Group as deafault and if a Support Group is added we remove "Unknown" and link to the new Support Group. On the other hand, when a Support Group is removed and service is without Support Group we link it to the "Unknown" Support Group

## Considered Options Service

- **Don't change anything**
    We could be able to filter for a null value for service_id in component instance table. This would require a change in the database schema since this field can't be null at the moment. 

- **Unknown Service**
    This would be a similar approch as to the default "Unknown" Support Group. As Component Instance that has no Service would be linked to the "Unknown" Service. And again, when a Service is removed and the Component Instance is without the Service we link it to the "Unknown" Service


## Decision Outcome

The decision is to have the "unknown" Service and SupportGroup that can be attached by default to a ComponentInstance that does not have a Service.
In addition we need to add an Event Handler (in the App Layer) that checks and removes the "unknown" relationship if any other Group/Service is present or has been added.

### Consequences

* Neutral, because of additional implementation of business logic and requires managing these relations
* Good, because its easier to fiter for Group "Unknown"



## Pros and Cons of the Options

### Options Support Group

#### {Option Support Group - "Don't Change Anything"}


* Good, because it won't need a "dummy" SG
* Bad, because we need to implement custom filtering logic since the current filter wouldn't work for this scenario

#### {Option Support Group - "Deafault Support Group Unknown"}


* Good, because we can easily filter for "unknown" in the front-end
* Neutral, because this might not be the most elegant solution
* Bad, because we require additional checks for managing support groups
* Bad, because of implementation effort


### Options Service

#### {Option Service - "Don't Change Anything"}


* Good, because the easy filtering
* Good, because no big additional implementation effort
* Bad, because it requires a change in the DB Schema. Null-Values would be a option now

#### {Option Service - "Deafault Service Unknown"}


* Neutral, because it would have the same logic as in SG (if we would implement it for SG)
* Bad, because it requires managing relations to unknown Service object as in the Support Group example