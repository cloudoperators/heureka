# Component Versioning for Openstack Entities

## Context and Problem Statement

How to accomodate Openstack entities like Security Groups in the existing Entity Relationship of Heureka? 

The current Entity Relationships in Heureka are defined for vulnerabilities in objects like containers, where it is intuitive to have ComponentVersion and ComponentInstances. For non-compliance (not exactly vulnerabilities) in entities like Security Group, how to adapt to existing data model?

## Decision Drivers

* Fits into existing data model 
* Make use of Heureka Issue Matching through Component Version

## Considered Options

* Opt 1: Using the configuration as ComponentVersion
* Opt 2: Adding ComponentContext to ComponentInstance
* Opt 3: Do not store any configuration context data

## Decision Outcome

Chosen option: "Opt 1: Using Configuration as ComponentVersion", because neither option 2 or 3 fully utilize existing entity relationships in Heureka

### Consequences

* Good, because we can decrypt the hash to get the configuration back to do processing
* Good, because we are using the same Entity Relationship logic for "Vulnerabilities" as well, so Heureka now handles issue matching, not the scanner
* Bad, because the naming of ComponentVersion may be confusing, as we only store the hash of the configuration, not the config itself


## Pros and Cons of the Options

### Opt 1: Using the configuration as ComponentVersion
![](../images/components-for-policy-scanning-opt1.png)
We use the configuration of the Openstack entity and hash it to create a unique ComponentVersion.

* Good, because we can decrypt the hash to get the configuration back to do processing
* Good, because we are using the same Entity Relationship logic for "Vulnerabilities" as well, so Heureka now handles issue matching, not the scanner
* Bad, because the naming of ComponentVersion may be confusing, as we only store the hash of the configuration, not the config itself

### Opt 2: Adding ComponentContext to ComponentInstance
![](../images/components-for-policy-scanning-opt2.png)
We add a new entity called ComponentContext to the Entity Relationship for Component Instance

* Good, because having ComponentContext makes it easier to differentiate between Component Instances
* Bad, because there is no matching happening in Heureka, and scanner handles it without ComponentVersion
* Bad, because we are NOT using the same Entity Relationship logic for "Vulnerabilities"

### Opt 3: Do not store any configuration context data
![](../images/components-for-policy-scanning-opt3.png)
We do not store ComponentVersion or add ComponentContext for Openstack ComponentInstances

* Good, because it simplies handling the configuration and versioning
* Bad, because there is no matching happening in Heureka, and scanner handles it without ComponentVersion
* Bad, because we are NOT using the same Entity Relationship logic for "Vulnerabilities"
