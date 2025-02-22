# Terraform Provider Neon

-----

<div align="center">
    ⭐ The project needs your support! Please leave a star and become a GitHub sponsor! ⭐
</div>

-----

<div align="center">
    💖 Thank you <a href="https://github.com/neondatabase">@neondatabase</a> for sponsoring the project! 💖
</div>

-----

Terraform provider to manage the [Neon](https://neon.tech/) Postgres projects.

## Using the provider

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

### Authentication and Configuration

Configuration for the Neon Provider can be derived from several sources, which are applied in the following order:

1. Parameters in the provider configuration

```terraform
provider "neon" {
  api_key = "<neon-api_key>"
}
```

2. Environment variables:
- Api key specified as `NEON_API_KEY`

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.14.x

## Building The Provider

1. Clone the repository
2. Enter the repository directory
3. Build the provider using the Go `install` command: 
```sh
make build
```
4. Run to install the provider to be used locally:
```sh
make install
```

## Updating the Provider

When updating your provider version, avoid using `terraform init -upgrade` in CI pipelines and auto-approved pull requests, as this can lead to unintended resource replacements and data loss. Instead, use `terraform init` in your automated workflows. Running `terraform init -upgrade` should always be done manually, followed by plan reviews. For additional guidance, see [Important usage notes](https://neon.tech/docs/reference/terraform#important-usage-notes) in the Neon documentation.

## Local development

### Requirements

- [Go](https://golang.org/doc/install) (find required version in [go.mod](go.mod))
- gnuMake / cMake

### Commands for local development

Run to see the full list of commands:

```commandline
make help
```

Run to compile the provider:

```commandline
make build
```

It will yield the binary `terraform-provider-neon_vdev`.

Run to install the provider locally:

```commandline
make install
```

It will yield the path to the built binary.

To generate or update documentation, run `go docu`.

In order to run the full suite of Unit tests, run `make test`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.
