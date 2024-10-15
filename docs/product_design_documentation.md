# Product Design Document - Heureka 

## Vision

**Heureka is a Security Posture Management tool designed to manage security issues in a cloud operating system.**

Heureka is committed to empowering cloud operators with a central platform for _proactive_ and _automated_ **assessment** and **remediation** of security issues while easing **compliance** fulfillment.​

- **Issue Assessment** - automated identification, classification, and prioritization of security issues.​ Example: assigning a severity level based on a service's classification of high availability, integrity, or confidentiality.​


- **Remediation Tracking** - streamlined and automated tracking of all operations involved in addressing security issues. The current scope includes tracking security patches/updates to address known vulnerabilities; tracking services to ensure they are configured according to security standards; and triaging detected threats and providing a means to respond/address threats.


- **Compliance Management** - Ensuring state change documentation​ and adherence to relevant security regulations and industry standards. e.g., PCI DSS,

>>> _A Security Issue​ refers to any problem that can compromise confidentiality, integrity, and availability e.g.: Vulnerabilities; threats; exposures (due to misconfigurations)_


## Problem Statements

### Complexity and Visibility
Maintaining security in a complex cloud operations platform landscape is a monumental task. These landscapes often consist of numerous services, each made up of multiple components like images, databases, libraries, and configurations. 
The challenge is compounded by the fact that these components have varying versions and can be shared across multiple services creating a critical need to pinpoint the specific instance (version) of a component as the security baseline. Therefore, vulnerabilities of a single component can impact multiple services within the landscape.

![image](https://github.com/user-attachments/assets/91e7507e-dd86-40d8-8a32-c35825d5ff03)


### Compliance and Efficiency
Meeting compliance requirements and maintaining robust security standards is time-consuming and resource-intensive due to the lack of centralized visibility into the intricate relationships and dependencies between services and their underlying components and configurations.
This makes tracking remediation, documenting evidence, and managing security configurations difficult leading to inefficient security operations and delayed remediation efforts.

![image](https://github.com/user-attachments/assets/d1248c66-d3df-4e58-aa08-12e0115669e9)

 
## Business Goals 
 
- Enhance Visibility and Security Posture:
  - Track the overall state of the cloud operating system 
  - Track security issues associated with specific components of the cloud operating system
 
- Streamline Security Operations:
  - Provide a central platform to monitor and assess the overall security posture of the technology landscape
  - Streamline the remediation process to close vulnerabilities and reduce attack surfaces.
  - Enforce consistent security configurations across all systems to minimize misconfigurations that could create security gaps.
  - Improve threat detection and response with Security Incident and Event Management (SIEM) Integration.

- Enhance Compliance, and Auditability:
  - Assists in meeting security compliance requirements by tracking/documenting assessments, and remediation status/progress thereby providing evidence of adherence.
  - Document all state changes resulting from the tracked remediation activities.

<br/>      

## Terms

| Term | Description |
| --- | --- |
| Support Group | A support group consists of multiple users working on a defined scope of services |
| Support Group Owner | Owners/ administrators of Support Group |
| Support Group Engineer | Individual users that are members of a support group |
| Service | Services are the main entities used by Heureka to represent our in. A service consists of one or multiple components. With an owner and at least one delegate |
| Component | A service consists of one or more components and is the basis for confirming a service is affected by an issue. There are different component types such as (Keppel)Image of Github Repository. |
| Component Instance | A component that is an instance of a component |
| Component Version | A version of a specific component at a certain point in time |
| Package | Packages are a technical detail for Keppel Images. Keppel uses Clair as a vulnerability scanner and performs package indexing. This package list gives information about what is used in the Keppel Image. |
| Issue | An issue is the absolute base unit of a weakness which could be a vulnerability, security event, or policy violation |
| Issue Match | This is the association of a found weakness to the deficient resource (component instance). It therefore represents the issue to be fixed|
| Service Owner | A service owner is the responsible person for a service. They represent the service in an audit and are responsible for the fulfillment of security controls |
| Delegate | A person that is a delegate for the Service Owner |
| Activity | The collects all remediation-related information and related changes for each issue match|
| Change | A change captures a change event in a specific component |
| Evidence | Evidence consists of all audit-relevant information. |
| Target Remediation Timeline | Target Remediation Timeline - the timeline in which a specific issue match has to get remediated |
| Remediation | A change that eliminates an issue match |
| Irrelevance Statement | A reasoned statement that a specific "Issue Match" is irrelevant for a component or set of components |
| Rollback | The rollback of a change for a component(s)/issue(s) combination |
| Process Facilitator | The person responsible for ensuring that all established processes are followed including the patch management process and the Security Information & Event Management (SIEM) process |
        
<br/>

## User Profiles 

### Auditor	
Responsible for validating the compliance of the platform to Industry Standards.

- #### Goals

	* Validate if all requirements of industry standards are met or not.	

- #### Tasks 

	* Review logs and historical remediation activities to verify that the process is followed
	* Review remediation activities; this includes, patching, Security Event Alert resolution, and policy violations
	* Review audit artifacts

### Service Owner	

The responsible person for a Service - manages a service’s complete lifecycle	

- #### Goals

	* Ensure availability and quality of a specific service and that remediation TRT are met	

- #### Tasks 

	* Define what components are belonging to the service
	* Owns risk definitions
	* Owns service 
	* Coordinate activities and participate in audit sessions as a service rep
	* Review vulnerability and patching activities 
	* Handles security event alerts as well as policy violations
	* Review evidence
	* Monitor issue statuses

### Support Group Engineer	

A group of experts dedicated to managing component instances and is responsible for performing actual issue-remediation activities	

- #### Goals

	* Seamless activity tracking	

- #### Tasks 

	* Plan activities
	* Monitor issue matches and issue match states
	* Perform manual patches, issue remediations, and respond to SIEM alerts. 

### Process Facilitator	

Responsible for ensuring that all established processes are followed including the patch management process and the Security Information & Event Management (SIEM) process

- #### Goals

	* Overview of activities 
	* Detect deficiencies timely	

- #### Tasks 

	* Perform monthly self- assessment 
	* Ensure that the audit-relevant information required as evidence is correct, complete, and maintained.

<br/>  


## High-Level Features

**TBD**





