# Semantic Release in the Heureka Project

[Semantic Release](https://github.com/semantic-release/semantic-release) is an automated versioning and package publishing tool that helps maintain consistent versioning and
changelogs in a project. Heureka uses the [Semantic Release Action](https://github.com/codfish/semantic-release-action) to pass the output to the next Workflow steps.

Semantic Release requires a `.releaserc.json` configuration file in the root directory to define how it should operate.

Example:

```
{
  "branches": ["main"],
  "plugins": [
    "@semantic-release/commit-analyzer",
    "@semantic-release/release-notes-generator",
    [
      "@semantic-release/github",
      {
        "assets": [
          {
            "path": "dist/*.tar.gz",
            "label": "Binary distribution"
          }
        ],
        "successComment": false
      }
    ]
  ]
}
```

### Explanation of Configuration

- branches: Specifies the branches on which releases should be made. Typically, this is the main branch.
- plugins: A list of plugins that Semantic Release uses to analyze commits, generate release notes, update changelogs, publish to npm, and push changes back to the repository.


## Workflow

Heureka uses GitHub Actions to automate the release process. The workflow is defined in `.github/workflows/release.yaml`.

Steps:
- Semantic Release: Determines if a new version should be released based on commit messages
- Build Docker Images: Builds and pushes Docker images with appropriate tags
- Vulnerability Scan: Scans Docker images for vulnerabilities using Trivy

## Commit Message Guidelines

Semantic Release relies on commit messages to determine the type of version bump. Heureka follows the 
[Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) specification:

- feat: A new feature
- fix: A bug fix
- docs: Documentation changes
- style: Code style changes (formatting, missing semi-colons, etc.)
- refactor: Code changes that neither fix a bug nor add a feature
- perf: Performance improvements
- test: Adding or updating tests
- chore: Changes to the build process or auxiliary tools

Example commit message:

```
feat: add new user authentication module
```

### Commit Message Types

Commit Message Types

1. Patch Release:
  - Triggered by commit messages that indicate bug fixes. These messages use the fix type.
  - Example: fix: correct typo in user authentication module
2. Minor Release:
  - Triggered by commit messages that introduce new features without breaking existing functionality. These messages use the feat type.
  - Example: feat: add new user authentication module
3. Major Release:
  - Triggered by commit messages that introduce breaking changes. These messages can use any type, but must include a BREAKING CHANGE section in the commit message body.
  - Example:
    ```
    feat: refactor authentication module
    BREAKING CHANGE: The authentication module API has changed and is not backward compatible.
    ```

## Release Process

- Commit Changes: Developers commit changes following the Conventional Commits guidelines.
- Pull Request: Changes are reviewed before merge
- Merge to Main: Changes are merged to the main branch.
- CI/CD Pipeline: The GitHub Actions workflow is triggered.
- Semantic Release Execution: Semantic Release analyzes the commit messages, determines the version bump, updates the changelog, and publishes the package.
