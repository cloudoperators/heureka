# Remediation Management


## Introduction/Overview:
This feature empowers cloud operators to track each vulnerability's lifecycle, providing complete visibility from identification to remediation. It enables more efficient vulnerability management and resolution, enhancing the overall security posture.


## Objectives and Goals:
...


## Personas
Operator - 
Auditor - 
Process Owner - 


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


## Details of Actions

### **False Positive:**

- A description or reason (mandatory) must be entered by the operator after selecting an issue and choosing the **False Positive** action.
- No batch actions are allowed at this stage; only single actions per issue.

### **Risk Acceptance:**

- A description or reason (mandatory) must be entered by the operator after selecting an issue and choosing the **Risk Acceptance** action.
- No batch actions are allowed at this stage; only single actions per issue.
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


### **Manual Mitigation:**

- A description of the action(s) taken (mandatory) must be provided for the selected issue.

### **Change Severity** *(this action will only be offered upon request)*:

- An operator should be provided with a list of severity levels that can be changed to.
- A description or justification for the severity change must be included.


### Non-Functional Requirements:
*```Consider including performance, security, usability, and reliability requirements```*.
...


## Remediation Management User Stories
...


## Dependencies and Assumptions:
*```Any dependencies on other systems, teams, or technologies, as well as assumptions made during the planning process.```*
...


## Risks and Mitigations:
...


## Potential risks associated with the feature and strategies to mitigate them.
...


## Timeline and Milestones:
*```An estimated timeline for development, testing, and deployment, along with key milestones.```*
...


## Glossary:
*```Definitions of any technical terms or acronyms used in the document.*```
...



