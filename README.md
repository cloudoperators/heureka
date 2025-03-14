# Heureka

[![REUSE status](https://api.reuse.software/badge/github.com/cloudoperators/heureka)](https://api.reuse.software/info/github.com/cloudoperators/heureka)


**Heureka** is a Security Posture Management tool designed to manage security issues in a cloud operating system. 

Its primary focus is remediation management of security issues such as vulnerabilities, security events, and policy violations while ensuring compliance and auditability.


## Value Propositions

**1. Enhanced Visibility and Security Posture**

A holistic view of the technology landscape, enabling proactive identification and tracking of security issues.

**2. Streamlined Security Operations**

Centrally manage security posture, automate patch management, enforce consistent configurations, and improve threat detection with SIEM integration.

**3. Enhanced Compliance, and Auditability**

Facilitate compliance by tracking remediation progress and providing a complete audit trail (evidence) with detailed documentation of state changes and actions taken.


## Architecture & Design

For a detailed understanding of Heureka's architecture and design, refer to the following resources:

- [Heureka Product Design Document](docs/product_design_documentation.md): This document provides a general overview, a glossary of terms, and user personas.
- [Entity Relationship Documentation](docs/entity_relationships.md): This document outlines the core entities within Heureka and how they interact.
- [High-Level Architecture Diagram](https://github.com/cloudoperators/heureka/blob/main/docs/product_design_documentation.md#high-level-features): This provides a visual representation of the overall system architecture.
- [High-Level Features](https://github.com/cloudoperators/heureka/blob/main/docs/product_design_documentation.md#high-level-features): A high-level overview of Heureka's functionalities.

## Requirements and Setup

The application can be configured using environment variables. These variables are stored in a `.env` file at the root of the project.
For configuring tests, there is a separate `.test.env` file.

Here's a basic example of what the .env file could look like:

```
DB_USER=my_username
DB_PASSWORD=my_password
DB_ROOT_PASSWORD=my_password
DB_NAME=heureka
DB_ADDRESS=localhost
DB_PORT=3306
DB_SCHEMA=internal/database/mariadb/init/schema.sql

DB_CONTAINER_IMAGE=mariadb:latest

DOCKER_IMAGE_REGISTRY=hub.docker.com

DOCKER_CREDENTIAL_STORE=docker-credential-desktop

LOG_PRETTY_PRINT=true

LOCAL_TEST_DB=true

SEED_MODE=false
```

To enable JWT token authentication, define `AUTH_TOKEN_SECRET` environment variable. Those variable is read by application on startup to start token validation middleware.

### Docker

The `docker-compose.yml` file defines two profiles: `db` for the `heureka-db` service and `heureka` for the `heureka-app` service.
To start a specific service with its profile, use the --profile option followed by the profile name.

For example, to start the heureka-db service, run:
```
docker-compose --profile db up
```

And to start the heureka-app service, run:
```
docker-compose --profile heureka up
```

To start both services at the same time, run:
```
docker-compose --profile db --profile heureka up
```

### Makefile

The application can be started by using the provided Makefile:

```
make start-all-heureka
```

### Devcontainers

Devcontainers is a new standard for development environments based on (docker)
[devcontainer](./.devcontainer).

At the moment devcontainers are supported by Visual Studio Code and IDEA IDEs.

For Microsoft Visual Studio code, install the remote container extension via
Ctrl-P and this command:

        ext install ms-vscode-remote.remote-containers

When opening the root folder in Visual Studio code a prompt will ask you to
open the project in a dev container, which you should.

Once inside the devcontainer the provided launch.json is configured to allow
launching heureka and running the unit and integration tests.

At the moment there is a known issue with the permissions of the .mariadb-dev
folder. This folder has to be deleted every time after using the devcontainers.
Use the following command in the root folder of heureka:

    sudo rm -rf .mariadb-dev


### Tests

#### Mockery

Heureka uses [Mockery](https://vektra.github.io/mockery/) for building Mocks based on defined interfaces for the purpose of Unit-Testing.

The Makefile/Dockerfile take care of installing mockery via

    go install github.com/vektra/mockery/v2@v2.46.3

#### Using Ginkgo

Heureka uses [Ginkgo](https://onsi.github.io/ginkgo/) for behavior-driven development (BDD) style tests. In the current project setup, tests are organized into three different directories, each serving a specific purpose:

- End-to-End Tests: These tests are located in the ./internal/e2e directory. End-to-end tests are designed to test the flow of an application from start to finish and ensure all integrated pieces of an application function as expected together.

- Application Layer Tests: These tests resides in the ./internal/app directory. Application layer tests focus on testing the application's behavior, such as handling requests and responses, executing appropriate logic, and more.

- Database Tests: These tests are found in the ./internal/database/db directory. Database tests ensure that the application correctly interacts with the database. They test database queries, updates, deletions, and other related operations.

In the `.test.env` file, the `LOCAL_TEST_DB` variable controls the database used for testing:

- If `LOCAL_TEST_DB=true`, tests will interact with a **local database**. Please ensure your local database server is running before executing the tests.
- If `LOCAL_TEST_DB=false`, tests will run against a **containerized database**.

Run all tests:
```
ginkgo -r
```

Run end-to-end tests:
```
ginkgo ./internal/e2e
```

Run application tests:
```
ginkgo ./internal/app
```

Run database tests:
```
ginkgo ./internal/database/mariadb
```

The ginkgo `-focus` allows using a regular expression to run a specific test:
```
ginkgo -focus="Getting Services" ./internal/database/mariadb
```
If the test block you're trying to run depends on `BeforeEach`, `JustBeforeEach`, or `Describe` blocks that aren't being run when you use the `-focus` flag, this could cause the test to fail.

## Support, Feedback, Contributing

This project is open to feature requests/suggestions, bug reports etc. via [GitHub issues](https://github.com/SAP/<your-project>/issues). Contribution and feedback are encouraged and always welcome. For more information about how to contribute, the project structure, as well as additional contribution information, see our [Contribution Guidelines](CONTRIBUTING.md).

## Security / Disclosure
If you find any bug that may be a security problem, please follow our instructions at [in our security policy](https://github.com/SAP/<your-project>/security/policy) on how to report it. Please do not create GitHub issues for security-related doubts or problems.

## Code of Conduct

We as members, contributors, and leaders pledge to make participation in our community a harassment-free experience for everyone. By participating in this project, you agree to abide by its [Code of Conduct](https://github.com/SAP/.github/blob/main/CODE_OF_CONDUCT.md) at all times.

## Licensing

Copyright 2004 SAP SE or an SAP affiliate company and heureka contributors. Please see our [LICENSE](LICENSE) for copyright and license information. Detailed information including third-party components and their licensing/copyright information is available [via the REUSE tool](https://api.reuse.software/info/github.com/SAP/<your-project>).
