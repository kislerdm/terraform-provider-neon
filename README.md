# Terraform Provider Neon

Terraform provider to manage the [Neon](https://neon.tech/) Postgres projects.

## Using the provider

```terraform
terraform {
    required_providers {
        neon = {
            source = "kislerdm/terraform-provider-neon"
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

-	[Terraform](https://www.terraform.io/downloads.html) >= 0.13.x
-	[Go](https://golang.org/doc/install) >= 1.17

## Building The Provider

1. Clone the repository
2. Enter the repository directory
3. Build the provider using the Go `install` command: 
```sh
$ go install
```
4. Run to install the provider to be used locally:
```sh
make install
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `make install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```sh
$ make testacc
```
