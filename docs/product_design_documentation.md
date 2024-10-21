# Product Design Document - Heureka 

## Vision

**Heureka is a Security Posture Management tool designed to manage security issues in a cloud operating system.**

Heureka is committed to empowering cloud operators with a central platform for _proactive_ and _automated_ **assessment** and **remediation** of security issues while easing **compliance** fulfillment.​

- **Issue Assessment** - automated identification, classification, and prioritization of security issues.​ Example: assigning a severity level based on a service's classification of high availability, integrity, or confidentiality.​


- **Remediation Tracking** — streamlined and automated tracking of all operations involved in addressing security issues. The current scope includes tracking security patches/updates to address known vulnerabilities; tracking services to ensure they are configured according to security standards; triaging detected threats and providing a means to respond to/address threats.


- **Compliance Management** - Ensuring state change documentation​ and adherence to relevant security regulations and industry standards. e.g., PCI DSS,

>>> _A Security Issue​ refers to any problem that can compromise confidentiality, integrity, and availability e.g.: Vulnerabilities; threats; exposures (due to misconfigurations)_


## Problem Statements

### Complexity and Visibility
Maintaining security in a cloud operations platform landscape is a monumental task. These landscapes often consist of numerous services, each comprising multiple components like images, databases, libraries, and configurations. 
The challenge is compounded by the fact that these components have varying versions and can be shared across multiple services, creating a critical need to pinpoint a component's specific instance (version) as the security baseline. Therefore, vulnerabilities of a single component can impact multiple services within the landscape.

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
  - Provide a central platform to monitor and assess the overall security posture of the cloud operating system
  - Streamline the remediation process to close vulnerabilities and reduce attack surfaces.
  - Enforce consistent security configurations across all systems to minimize misconfigurations.
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
| Service | Services are the main entities in Heureka. A service consists of one or multiple components. With an owner and at least one delegate |
| Component | A component is the baseline for confirming a service is affected by an issue. There are different component types such as (Keppel)Image of Github Repository. |
| Component Instance | A component that is an instance of a component |
| Component Version | A version of a specific component at a certain point in time |
| Package | Packages are a technical detail for Keppel Images. Keppel uses Clair as a vulnerability scanner and performs package indexing. This package list gives information about what is used in the Keppel Image. |
| Issue | An issue is the absolute base unit of a weakness which could be a vulnerability, security event, or policy violation |
| Issue Match | This is the association of a found weakness to the deficient resource (component instance). It therefore represents the issue to be fixed|
| Service Owner | A service owner is the responsible person for a service. They represent the service in an audit and are responsible for the fulfillment of security controls |
| Delegate | A person that is a delegate for the Service Owner |
| Activity | The collects all remediation-related information and associated changes for each issue match|
| Change | A change captures a change event in a specific component |
| Evidence | Evidence consists of all audit-relevant information. |
| Target Remediation Timeline | The timeline in which a specific issue match has to get remediated |
| Remediation | A change that eliminates an issue match |
| Irrelevance Statement | A reasoned statement that a specific "Issue Match" is irrelevant for a component or set of components |
| Rollback | The rollback of a change for a component(s)/issue(s) combination |
| Process Facilitator | The person responsible for ensuring that all established processes are followed including the patch management process and the Security Information & Event Management (SIEM) process |
        
<br/>

## Personas

### The Engineer/Operator	

The Engineer belongs to any of the teams responsible for maintaining the resources of a cloud operating system, such as individual applications, OpenStack services, etc.
>Example Role: service owner, service support group engineer

- #### Goals

 	* Keeping all services available and in pristine condition - no vulnerabilities, no suspicious activity unaddressed, no misconfiguration
  	* Monitors found issues, and ensure remediation timelines are met.

- #### Challenges

 	* Manually tracking (and documenting) all components consumed by owned service for open issues can be time-consuming and error-prone.
  	* Manually tracking  remediation timelines is challenging due to varying issue complexities.

 - #### Needs

 	* An automated tracking of all components consumed by owned service for open issues.
  	* A centralized and automated documentation system that tracks the management of remediation activities.
  

### The Manager	

A management professional responsible for the overall well-being of security posture and overseeing the implementation and maintenance of security systems and protocols
>Example Role: Team Manager, Manager of Managers

- #### Goals

 	* Maintain a comprehensive overview of the organization's security posture.
  	* Ensure effective implementation and adherence to security protocols within the team.

- #### Challenges

 	* The lack of a centralized security management system makes getting a holistic view of the organization's security posture difficult.
  	* Lack of visibility of the effect of specific issues on the cloud operating system.

 - #### Needs

 	* A centralized security and comprehensive view of the organization's security posture.
  	* Drill down on specific issues and fetch associating information such as the component/services impacted, the remediation efforts/status, and user affiliations.



### The Process Owner	

Any central security/compliance team member, responsible for the compliance-related processes within the cloud operating system. This includes the Patch Management process, the SIEM process, etc. This is usually a single person ensuring that all activities relating to a specific compliance process align with the security control guiding it toward meeting regulatory requirements.
>Example Role: Patch Management Process Owner, SIEM Process Owner


- #### Goals

	* Maintain the integrity of a compliance process
 	* Ensure the audit-relevant information required as evidence is correct, complete, and maintained.

- #### Challenges 

	* Ensuring all process stakeholders participate in their respective activities
 	* Difficulty in coordinating and managing multiple stakeholders

 - #### Needs

 	* A Centralised system that triages and categorizes all issues and provides relevant information such as type of issue, remediation timeline, remediation activities
  	* A centralized system to facilitate the coordination and management of multiple stakeholders.


### The Third Party Integrator
	
An engineer responsible for the technical implementation of security measures. They triage  existing security elements via the Heureka plugin and ensure its seamless integration into the cloud operating system.
>Example Role: A Security Engineer

- #### Goals

 	* Efficiently integrate and enable the Heureka plugin in the organization's technology landscape.
  	* Continuously monitor and optimize the performance of the Heureka plugin.

- #### Challenges

 	* Streamlining various security processes and tooling can be complex and time-consuming.

 - #### Needs

 	* A comprehensive understanding of the Heureka plugin and its technical requirements for efficient integration.
  	* To be able to validate the proper configuration of Heureka



### The Auditor	
A representative of a specific business unit, responsible for confirming the integrity of the security posture management system in terms of compliance. This individual is particularly interested in audit artifacts and plays a crucial role in maintaining the organization's adherence to regulatory standards.
>Example Role: A Business Unit Auditor

- #### Goals

 	* Ensures adherence to remediation timelines.
 	* Ensures remediation activities are carried out
 	* Maintain and review audit trails and artifacts to ensure all processes are compliant.	

- #### Challenges

 	* The absence of a centralized tooling system makes it difficult to monitor compliance across different services
 	* Ensuring consistent compliance across all teams of the LOB

 - #### Needs

 	* A centralized view providing a comprehensive view of compliance across different services.
 	* Clear communication channels for collaboration and information sharing within the business unit.

<br/>  


## High-Level Features

### Security Issue Overview

This provides a comprehensive and interactive overview of all security issues in the cloud operating system including real-time data on  vulnerabilities, policy violations, and active threats. 
It allows for easy navigation and drill-down into specific issues for detailed information facilitating prompt and effective response.
Powered by scanner engines gathering new issues from established repositories. It includes:

- CVE Scanners: Collects data from CVE advisories, providing up-to-date information on potential system vulnerabilities.
- Policy Violations Scanners: Interacts with the Policy Gatekeeper tool to identify actions that violate predefined security policies.
- SIEM Alerts Scanners: Interfaces with a SIEM system, triaging alerts of security threats and potential breaches.


### Inventory Overview
This provides a comprehensive view of all components within the cloud operating system from a Service point of view.
It provides detailed information about existing component instances and versions of a service along with the owning team.
This helps assess the impact of security issues, plan for improvements, and ensure oversight of the cloud operating system's security posture.

### Issue Classification/Documentation
This feature ensures a systematic and compliant classification of all security issues. It involves documenting essential attributes such as a unique ID (e.g., CVE ID), severity level, target remediation time, and affected service resolution efforts.
This comprehensive documentation aids in prioritizing and managing security issues effectively.

### Issue Remediation Management
This feature empowers cloud operators to track the entire lifecycle of each issue, providing complete visibility from identification to remediation. It enables more efficient issue management and resolution, thereby enhancing the overall security posture.

### Compliance Artifact Management
This feature simplifies the fulfillment of compliance requirements. by providing real-time, comprehensive audit artifacts and evidence. It ensures that all necessary compliance documentation is readily available and up-to-date, making audits smoother and more efficient.

### Alerting and Notifications
This feature would provide real-time alerts and notifications about new and emerging security issues. This could include email notifications, SMS alerts, or integration with communication platforms like Slack. This would ensure that teams are immediately aware of any issues and can respond promptly.



## High-Level Architecture

![](https://objectstore-3.eu-nl-1.cloud.sap/v1/AUTH_8eba81a5654c4bb2a86fde93ccc33cab/codimd-images/uploads/48a8cca4-6fad-448e-b721-c9fe562f8e8e.png)



