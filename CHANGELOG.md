# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.0.2] - unreleased

### Added

- Branch resource:

```terraform
resource "neon_project" "this" {
  name = "foo"
}

resource "neon_branch" "this" {
  project_id = neon_project.foo.id
  name       = "bar"
}
```

### Changed

- Bumped [Neon Go SDK](https://pkg.go.dev/github.com/kislerdm/neon-sdk-go) v0.1.3

## [0.0.1] - 2022-01-05

### Added

- `Neon` Provider:

```terraform
terraform {
  required_providers {
    neon = {
      source = "kislerdm/neon"
    }
  }
}

provider "neon" {}
```

- Project resource:

```terraform
resource "neon_project" "this" {
  name = "foo"
}
```
