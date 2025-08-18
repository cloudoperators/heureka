# Remediation Management


## Introduction/Overview:
This remediation management feature 0f Heureka allows cloud operators to track each vulnerability's lifecycle, providing complete visibility from identification to remediation. 


## Objectives:
- Facilitate efficient remediation processes by offering clear, actionable steps for addressing vulnerabilities, including marking false positives, accepting risks, and applying manual mitigations.
- Support compliance processes by integrating compliant action sequences in the Heureka.
- Improve informed decision-making by providing detailed information and options for each vulnerability..
- Ensure that all actions taken are logged and traceable, providing a clear audit trail for security and compliance purposes.


## Relevant Personas
- **Operator** - responsible for maintaining individual resources (Services) of the cloud operating system
- **Auditor** - responsible for confirming the integrity of the security posture management system in terms of compliance. 
- **Process Owner** - responsible for any of the compliance processes pertaining to the cloud operating system. E.g. Patch Management process, the SIEM process, etc


## Functional Requirements

- The remediation path will be introduced from the image list view of a service. This means we display images per service, along with an aggregated issues list for each image. Additionally, we include the actual image versions affected by each issue in the list.

   > **Proposed Navigation Flow:**
   >
   > - Service → Service Detail → Image(s) → Vulnerabilities (Issues)

- Remediation actions will be applied at the image level and should be applicable to all existing or upcoming versions it.

- For each issue in the issue list (on the image view), remediation actions will be presented. The remediation actions, listed below by order of implementation priority, are:

    1. False Positive
    2. Risk Acceptance
    3. Manual Mitigation
    4. Change Severity

   > **Note:** The Change Severity action may not be implemented.

- For each remediation action, users will have the option to apply it either:

    - On the service level for all existing active versions and all upcoming image versions (with the possibility to set an expiry date)
    - To all active versions
    - Only to specific active versions (by default, the affected versions can be pre-selected)

- The four remediation actions — **Risk Acceptance**, **False Positive**, **Change Severity**, and **Manual Mitigation** — can be applied to each issue and should offer the above-mentioned options to apply them to all or specific services.

- A view should be added to display the issues that have already been remediated per service, including the actions taken. The activity stream is organized per service, functioning as a changelog or history.

- No batch actions are allowed at this stage; only single actions per issue.


### **Remediation actions**

- **False Positive:**
    - A description or reason (mandatory) must be entered by the operator after selecting an issue and choosing the **False Positive** action.
  

- **Risk Acceptance:**
    - A description or reason (mandatory) must be entered by the operator after selecting an issue and choosing the **Risk Acceptance** action.
    - An operator should be able to accept an issue for a defined period as "Risk Accepted."
    - The duration must align with the risk acceptance guidelines defined in the SAP Exception Management Process (SEMP) provided by the SRC team to ensure compliance.

       > **The SEMP Outline:**
       > 1. SGE defines Risk Acceptance Details: Vulnerability Description, Justification for Accepting the Risk, Affected Service, Vulnerability Severity Level.
       > 
       > 2. The Risk Acceptance Details are aligned with Ying (SCI Risk Management Process Owner).
       > 
       > 3. Ying coordinates with the Risk Coordinator via email as per the official SEMP (Security Exception Management Process).
       > 
       > 4. Upon approval, a Jira ticket is available, and the vulnerability can be marked as ‘Risk Accepted’ in Heureka.
       > 
       > 5. The Jira ticket link along with the risk acceptance statement should be added 

    > **The SEMP outline is specific to SAP, how will this be handled for Apeiro?**


- **Manual Mitigation:**
    - A description of the action(s) taken (mandatory) must be provided for the selected issue.


- **Change Severity** *(this action will only be offered upon request)*:
    - An operator should be provided with a list of severity levels that can be changed to.
    - A description or justification for the severity change must be included.



## Non-Functional Requirements:

Performance: Views load without performance degradation.
Usability: The interface should be intuitive, allowing operators to perform tasks seamlessly


## Remediation Management User Stories

### U.S. 1 - View Remediation Attributes
As an operator, Process Owner, or Manager, I want to see the following remediation attributes for a vulnerability on every view with vulnerability information so that I can get a quick overview of my affected services and images:
- Remediation status
- Target Remediation Date (TRD)
- Discovery date
- Remediation action date


> Relevant views include: 
> - Service detail view (see remediation attributes for each image) 
> - Image detail view (see remediation attributes for each vulnerability)
> - Vulnerability List View

### U.S. 2 - Mark as False Positive 
As an operator, I want to be able to mark a vulnerability as false positive and provide the (Mandatory) reason for it being a false positive 

### U.S. 3 - Accept Risk
As an operator, I want to be able to accept the risk associated with a vulnerability and provide all risk information (Mandatory), including the duration of acceptance and a link to the official exception case

### U.S. 4 - Mark as Manually Mitigated
As an operator, I want to be able to mark a vulnerability as manually mitigated and provide the (Mandatory) mitigation details 

### U.S. 5 - Change Severity
As an operator, I want to be able to change the severity of a vulnerability which would result in a different Target Remediation Date (TRD).

### U.S. 6 - View Automatic Mitigation Progress
As an operator, i want to be able to see the mitigation progress while rolling out the patch in multiple steps

**Additional Information:**
>For scenarios where a rolled-out patch takes a while to be deployed in all the regions (bronze, silver, and gold regions), there is a need to display the status of the vulnerability it fixes accordingly and automatically in Heureka. Therefore, we need to show its mitigation status as "in progress" until the patch is deployed in all regions.

> User Stories 6 and 7 cover this edge case.

### U.S. 7 - View Patched and Un-patched Containers
As an operator, i want to be able to see the patched and un-patched containers for a specific vulnerability. 



## Dependencies and Assumptions:
**Dependencies**
- Vulnerabiltiy List View(s)
- Image Version centric to image centric switch


## Timeline and Milestones:
*```Milestone and Roadmap URLs' to be added here```*
