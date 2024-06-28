# Technical Design Document - Heureka 
 
 
### Roles and Responsibilities 
 
|                           | Name               | SAP User ID | 
| ------------------------- | ------------------ | ----------- | 
| Document Owner            | Michael Reimsbach  | I545293     | 
|                           | Lola Apenna        | I562822     | 
| Reviewer                  | David Rochow       | I577502     | 
|                           | Christian Schleich | D063376     | 
|                           | Fabian Kroenner    | D042704     | 
|                           | Esther Schmitz     | D038721     | 
|                           | Arturo Reuschenbac Puncernau | D063222| 
|                           | Abishak Koul       | I335040     | 
| Implementation Specialist | Michael Reimsbach  | I545293     | 
 
<br/>

## General Overview 
 
#### Business Goals 
 
At Converged Cloud, we have a vast, complex landscape of components where the responsibility of those components is split across multiple teams. Each of those components can potentially be affected by Vulnerabilities of the underlying technologies, libraries, misconfigurations, and custom implementations. 
 
To align with regulatory requirements, SAP corporate requirements, as well as with industry standards such as PCI DSS, ISO 27001, and others, we need a software-aided process to be able to: 
 
- track the overall state of our technology landscape 
- establish a unified & complete Patch Management Process: 
  - maintain & track vulnerabilities and affected components 
  - document the remediation, classification, and impact of vulnerabilities 
  - document the changes corresponding with Patching of Vulnerabilities, as well as Updating components 


<br/>      

## Terms


| Term | Description |
| --- | --- |
| Support Group | A support group consists of multiple users working on a defined scope of services |
| Support Group Owner | owners/ administrators of Support Group |
| Support Group Engineer | individual users that are member of a support group |
| Service | Services are the main entities used by Heureka to represent our in. A service consists of one or multiple components. With an owner and at least one delegate |
| Component | A service consists of one or more components and is used to see if a service is affected by a Vulnerability. There are different component types such as (Keppel)Image of Github Repository. |
| Component Instance | A component that is an instance of an component |
| Component Version | A version of a specific component at a certain point in time |
| Package | Packages are a technical detail for Keppel Images. Keppel uses clair as a vulnerability scanner and performs package indexing. This package list gives information about what is used in the Keppel Image. |
| Issue | The SCN (SAP Cert Notification) is one example for a security issue repository. In the future, there might be additional sources for advisories such as vmware security advisors or github security advisories. The there published vulnerabilities or weaknesses are named Issue |
| Issue Match | An found match of a Issue on a component |
| Service Owner | A service owner is the responsible person for a service which represents the service in an audit and is responsible for fullfillment of controls |
| Delegate | A person that is a delegate for the Service Owner |
| Activity | The collects all patching-related information and related changes. |
| Change | A change captures a change event in a specific component. |
| Evidence | Evidence consists of the audit relevant information. |
| Target Remediation Timeline | Target Remediation Timeline - the timeline in which a specific vulnerability match have to get remediated |
| Remediation | A change that eliminates a vulnerabilty match |
| Irrelevance Statement | A reasoned statement that a specific "Issue Match" is irrelevant for a component or set of components |
| Rollback | The rollback of a change for a component(s)/vulnerbilty(ies) combination |
| Process Facilitator | The person responible for ensuring that the patch management process is followed |


<br/>          
<br/>

## User Profiles
 

### Auditor	
Responsible for validating the compliance of the platform to Industry Standards.

- #### Goals

	* Validate if all requirements of a industry-standard are met or not.	

- #### Tasks 

	* Review logs to validate that process is followed
	* Review historical patching activities to verify that process is followed
	* Review Vulnerability and patching activities 
	* Review evidences


### Service Owner	

The responsible person for a Service, manages a service’s complete lifecycle	

- #### Goals

	* Ensure availability and quality of a specific service and that patch TRT are met	

- #### Tasks 

	* Define what components are belonging to the service
	* Owns risk definitions
	* Owns service 
	* Coordinating activities Participate in audit sessions as a service rep
	* Review Vulnerability and patching activities Review evidences
	* Monitor vulnerability/ patch statuses


### Support Group Engineer	

Manages component instances and is responsible for performing actual patch activities	

- #### Goals

	* Seemless patching activity tracking	

- #### Tasks 

	* Plan activities
	* Perform manual patches/ remediations 
	* Monitor vulnerability/ patch statuses


### Process Facilitator	

Resposnible for ensuring that the patch management process is followed	

- #### Goals

	* Overview of activities 
	* Detect deficiencies timely	

- #### Tasks 

	* Perform monthly self- assessment 
	* Ensure that the audit relevant information required as evidence is correct, complete and maintained.

<br/>  



## User Stories 
 
 <br/>

------
### Auditor 
------

#### A01 
 
##### User Story 
As an Auditor, I want to see who did which action and when to verify that the Vulnerability and Patch management process is followed according to company policies and that the platform is functioning as expected.



- ##### Acceptance Criteria
	1. Every state-changing action is logged into an immutable log collection, including:

	    - What action was performed
	    - Who performed the action
	    - When was the action performed
	    - Why was the action performed


	2. Every authentication to the platform is logged into an immutable log collection, including:

	    - Who logged in
	    - When was the logon


	3. Every component/vulnerability discovery is logged into an immutable log collection, including:


	    - Which tool did report the vulnerability
	    - What got discovered (all stored details)
	    - When was the report submitted



#### A02 
 
##### User Story 
As an Auditor, I want Heureka Monitoring to be able to verify that the Platform is monitored appropriately.

- ##### Acceptance Criteria
	1. The Monitoring includes the Heureka and all scanners
	2. The Monitoring includes alerts to stakeholders in case of outages
	3. The Monitoring retains 1 Year of Metric collection for health checks and availability


------

### Auditor or Service Owner 
 
------
 
#### ASO01 
 
##### User Story 
As an Auditor, I want to be able to list all in-scope issues (Issue List View) by filtering, showing the number of affected components, the number of affected activities, the earliest not yet met remediation target timeline, the earliest vulnerability match discovery date, and the severity rating to be able to validate the functioning of Converged Clouds Vulnerability & Patch Management.

- ##### Acceptance Criteria 
	1. Assuming I am on the Issue List view:

	    - I can filter / sort by
		    - issue
		    - service
		    - issue match discovery date
		    - the issue match remediation target date
	    - I only see relevant Issues, which are:
		    - issues with issue matches that are not manually marked as irrelevant with a reasoning
		    - issues with issue matches that are currently present in components
	    - I can select to display previously relevant (but not relevant anymore) issues
	    - I can click issue to navigate to the issue detail view (ASO02)  
		- I can select to display previously relevant (but not relevant anymore) vulnerabilities


#### ASO02 
 
##### User Story 
As an Auditor or Service Owner, I want to be able to view the details of a issue (Issue Detail View) and list down the affected components, including their severity rating, discovery date, and remediation target timeline, grouped by services & affected activities, including the current activity status to be able to validate the patch status of individual issues.


- ##### Acceptance Criteria 

	1. Assuming I am on the Issue detail view, I can see:

	    - all related activities grouped by the owning service

	    - all affected components grouped by the affected activities

	   - all affected components that are not acknowledged through a created activity

	    - a section, which is treated like a service named "Unassigned,"  which includes all Components that are not assigned to any service.

	    - I can see the following for each activity: the creation date

	    - I can see for each component:

		- the individual severity rating
		- the individual issue match discovery date
		- the individual remediation target date
		- the individual "remediation date", "acceptance date", "in progress" depending on the status of the individual issue match


#### ASO03 
 
##### User Story 
As an Auditor or Service Owner, I want to be able to view the details of an activity (Activity Detail View) and see all Updates (changes), including the fixed vulnerabilities and the affected components per change, all evidence for mitigations or irrelevance statements to be able to verify the status and correctness of individual patch procedures.

- ##### Acceptance Criteria 
	1. Assuming I am on the Activity Detail View:

	      - I can see all components in the scope
	      - I can see all issues in the scope
	      - I can see all Changes, including Before and After component status and vulnerability delta for the component
	    - I can see, if present, all Evidence for remediations and irrelevance statements, including remediated/irrelevant components, remediation description, verification description, verification evidence files (mandatory),  and remediated vulnerabilities tracked in a change

------

### Service Owner 

------

#### SO01 
 
##### User Story
As a Service Owner, I want to have an overview of my services (Service Owner Dashboard) and be able to assign components  to my Service to maintain them and have an  overview of my Services and the required Patch Management activities related to them.


- ##### Acceptance Criteria 

	1. Assuming I am on the Service Owner Dashboard, I can:

	    - See all my Services
	    - navigate to the service detail views by clicking the service
	    - show number of all vulnerability matches (divided by severity)
	    - show number of all activities


#### SO02
 
##### User Story 
As a Service Owner, I want to have a detailed overview of my services (Service Owner Dashboard) and be able to assign components  to my Service to maintain them and have an  overview of my Services and the required Patch Management activities related to them.

- ##### Acceptance Criteria 
	1. Assuming I am on the Service Owner Dashboard, I can:

	    -  see all components & component versions assigned to my service
	    - see all activities that are not finished for my service including a state
	    - see all vulnerability matches for my service(s) that are not covered by an activity
	    - Assign components to my service with a Filterable/Searchable select form


#### SO03 

##### User Story 
As a Service Owner i want to be able to create a new service

- ##### Acceptance Criteria 
	1. Assuming I am on the Service List View

	    - I can create a new service
	    - Add a description to the service
	    - Assign components and component versions to the service with a filterable/searchable multi select form


#### SO04 - Add Issue

##### User Story
As a service owner, I want to be able to add a new issue in the event that it isn’t already automatically discovered.

- ##### Acceptance Criteria 
	1. Assuming I am on the issue, I can:
	    - create a new issue
	    - Add a description to the VD
	    - Add an Issue match

------

### Facilitator

------

#### F02

##### User Story
As a Facilitator, I want to have an overview of Components (Components View) and be able to assign them to a Service to ensure no ownerless components


- ##### Acceptance Criteria 
	1. Assuming I am on the Components View, I can:
	    - See all components
	    - Assign  component versions to a service
	    - Filter components by Assignment (Unassigned or Service):
		- If a component does not have a version or is not associated with any service, it is considered unassigned.
	    - Filter components by labels/tags
	    - Filter components by name


------

### Service Owner or Support Group Engineer 

------

#### SOSGE01 
 
##### User Story 
As a Service Owner or Support Group Engineer, I want to get notifications about newly discovered vulnerability matches related to my service, as well as activities & vulnerabiliy matches that are close to the target remediation date and require immediate attention to know when I have to take a look into the Platform.

- ##### Acceptance Criteria 

	1. Assuming at least one Issue match or activity has less than 10 Days to resolve, I get a notification

	2. Notifications do not get sent out more often than once a Day

	3. Notifications are sent as a Summary and not as individual Notifications

	4. Notifications are sent to the respective support group channels through Prometheus metrics

 

#### SOSGE02 - merge of S02 and S03

##### User Story
As a Service owner or Support Group Engineer, I want to be able to track the acceptance of a vulnerability match and the irrelevance of a vulnerability match to justify why we did not remediate a vulnerability match in time.


- ##### Acceptance Criteria 
	1. Assuming I am on the “Activity Detail View”. I can click a button to be called "Add evidence" which opens a Form "Evidence Creation View" Where I see:


	    - A free text Texbox called  "Information".

	    - A multi-select field called "Type" with the options:
		- Mitigated
		- Risk Accepted
		- Marked as Irrelevant



	2. On submission of the Evidence Creation :

	    - The Evidence is attached to the activity

	    - The status of the Issue Match did not change  but the requested transition is displayed at the evidence entry which is highlighted in yellow/orange:
		- Mitigated
		- Risk Accepted
		- Marked as Irrelevant
	    - The vulnerability match gets the corresponding state attached after submission

	    - Changes on evidence type or deletion of evidence also reflects in vulnerability match status

	    - Changes to evidences are visible on the activity



#### SOSGE03

##### User Story
As a Service Owner or Support Group Engineer, I want to be able to adjust the Severity of a Issue Match to align it with the environmental cirumstances.


- ##### Acceptance Criteria 
	1. Assuming I am on the Issue Match Detail view, i have a form with 2 fields

	    - a multi select form with the severity levels to change to

	    - A free text field called reasoning

	    - On submission, the Severity of the Issue match. is changed and the reasoning is attached to the Issue Match



#### SOSGE04

##### User Story
As a Support Group Engineer I want components, component versions and component instances to automatically be attached to my service if they are labeled with **"service:(ServiceName)"** for any supported scanner.

- ##### Acceptance Criteria
    

	1. In case of not matchable entries the label is ignored

	2. Only labels starting with "service:" are evaluated and Everything After is evaluated as the Service Name.


------


### Support Group Engineer 
 
------
 
 
#### SGE01 
 
##### User Story 
    
As a Support Group Engineer, I want updates of components in scope to be automatically tracked as Activities & Changes to reduce the manual documentation effort.

- ##### Acceptance Criteria 
	1. Assuming I am Patching a  Component Version (such as a container), I want all related components versions ( and, through this update activity, remediated vulnerabilities to be added to an Activity, if they are assigned to the same service.

	2. For component instances not assigned to the same service activities per affected service are created that include all updates for the component instances of the respective service


	3. Assuming I am Updating a Component already in scope for a single or multiple Patches and the Update is affecting at least one vulnerability match in those patches, the Update is attached to the relevant activity as an immutable change.


	4. Assuming I have Updated all components of a activity and this resulted in remediation of all vulnerability matches, a activity should be automatically be completed/finished without the requirement of review through an facilitator


	5. The automated completion of an activity marks all related vulnerability matches as patched


	6. The automatically tracked changes include the change date, before component, after component, vulnerability delta, information source, date of entry


#### SGE04 
 
##### User Story 
As a Support Group Engineer, I want to see the target remediation time for a activity based on the in-scope Issue Matches to understand when I need to finish the complete activity.


- ##### Acceptance Criteria 
	1. Assuming I have created a activity

	    - I get redirected to the Activity Detail View (ASO03)

	    - The Activity Detail View as well contains the target remediation time for the individual vulnerability matches



#### SGE05 
 
##### User Story 
As a Support Group Engineer, I want to be able to see all my action items, which include not finished activities and Not (through activities) acknowledged vulnerability matches, including respective remediation target date, in one place (SGE Inbox View) to understand what activities and vulnerability matches I have to work on.

- ##### Acceptance Criteria 
	1. Assuming I am on the SGE Inbox View and I have 3 activities associated with my Support Group, 1 finished, 1 and 2/3 done, and one 0/3 done, and 10 vulnerability matches  that I need to mitigate that are not included in the patches. I see:

	    - A Section showing 2 activities:

		- one with 0/3 components fixed, including the number of remaining vulnerability matches, earliest target remediation time

		- one with 2/3 components fixed, including the number of remaining vulnerability matches, earliest target remediation time


	    - A Section with the 10 vulnerability matches:

		    - including for each vulnerability the count of affected components associated with my support group

		    - including for each vulnerability match the severity, target remediation DateTime, discovery DateTime


	2. Assuming I am on the SGE Inbox View and I can click a  Act button next to Issue Match and  I will get redirected to a Activity Creation View

<br/>      
<br/>  


## MVP

As per the workshop results, below is a list of features that have been recognized to be included in the MVP ( similar features have been clustered together).

<br/> 

**Cluster 1**
* Automatically detect component versions and map to service.
* Discover Components Instances of Service
* Detect and match new vulnerabilities to components/services
* Version match with component
* Populate vulnerabilities to the service objects


**Cluster 2**
* Discover Vulnerabilities
* Status update on vulnerability
* Provide overview of all unresolved vulnerabilities
* Overview of all vulnerabilities


**Cluster 3**
* Manage activities
* Add Evidences
* Capturing evidence details
* Automatically capture implemented patches to reduce manual interaction
    * 	clarify if we need to automatically link Change Management ticket


**Cluster 4**
* Manage Service Components
* Service List
* Name of Services???
* Service List


**Cluster 5**
* Automatically detect support group per service


**Cluster 6**
* Docker Images 
* Containers 


**Cluster 7**
* Filtering for all kind of Objects


**Cluster 8**
* User Dashboard

<br/> 

## **MVP-User Story Mapping**

<br/> 

| No | Feature Cluster | Mapping | Mapped User Story |
| --- | --- | --- | --- |
| 1 | Automated detection, mapping and mapping of objects | Automatic detection of Component instances, Component Versions and Vulnerabilties | SOSGE04 |
| ... | ... | Automatic mapping/matching of component version to component and vulnerability to service objects | SGE01 |
| 2 | All about Vulnerabilities: Discovering, detailed overview and Status information | Discovering, detailed overview | ASO01 |
| ... | ... | Status information | ASO02 |
| 3 | About Activities: actions and action evidences | Activities management | SGE01 |
| ... | ... | Evidence management | A01 |
| 4 | All about Services: Overview and management | Service Overview | SO01 |
| ... | ... | Service Management | SO02 |
| 5 | Automatically detect support group per service. | No U.S., but SOSGE01 covers notifying support groups. | ... |
| 6 | Docker Images | ... | ... |
| ... | Containers | ... | ... |
| 7 | Filtering for all kind of Objects | ... | ... |
| 8 | User Dashboard | ... | ... |


<br/>      
<br/>          

