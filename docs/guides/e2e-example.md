---
page_title: "End-to-end example: AWS Application"
---

# End-to-end example: AWS Application

The guide illustrates how to provision infrastructure for an application deployed in AWS to communicate with the Neon
database.

## Prerequisites

- terraform ~> v1.1
- Neon SaaS [API key](https://neon.tech/docs/manage/api-keys/)
- AWS account and role with sufficient privileges

## Code Snippet

The code below documents the following resources:

- the Neon stack:
    - project;
    - dedicated branch with
        - role;
        - database;
- the AWS Secretsmanager secret to store database access details securely;
- the AWS IAM policy to be attached to the AWS IAM role assumed by an application intended to communicate with the Neon
  database.

```terraform
terraform {
  required_providers {
    neon = {
      source  = "kislerdm/neon"
      version = ">= 0.2.1"
    }

    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.49"
    }
  }
}

provider "aws" {
  region = "us-east-2"
}

provider "neon" {}

resource "neon_project" "this" {
  name = "myproject"
}

resource "neon_endpoint" "this" {
  project_id = neon_project.this.id
  branch_id  = neon_branch.this.id
}

resource "neon_branch" "this" {
  project_id = neon_project.this.id
  parent_id  = neon_project.this.default_branch_id
  name       = "mybranch"
}

resource "neon_role" "this" {
  project_id = neon_project.this.id
  branch_id  = neon_branch.this.id
  name       = "myrole"
}

resource "neon_database" "this" {
  project_id = neon_project.this.id
  branch_id  = neon_branch.this.id
  owner_name = neon_role.this.name
  name       = "mydb"
}

resource "aws_secretsmanager_secret" "this" {
  name                    = "neon/mybranch/mydb/myrole"
  description             = "Neon SaaS access details for mydb, myrole @ mybranch"
  recovery_window_in_days = 0

  tags = {
    project  = "demo"
    platform = "neon"
  }
}

resource "aws_secretsmanager_secret_version" "this" {
  secret_id = aws_secretsmanager_secret.this.id
  secret_string = jsonencode({
    host     = neon_branch.this.host
    user     = neon_role.this.name
    password = neon_role.this.password
    dbname   = neon_database.this.name
  })
}

data "aws_iam_policy_document" "neon_access_secret" {
  statement {
    effect    = "Allow"
    actions   = ["secretsmanager:ListSecrets"]
    resources = ["*"]
  }

  statement {
    effect = "Allow"
    actions = [
      "secretsmanager:GetResourcePolicy",
      "secretsmanager:GetSecretValue",
      "secretsmanager:DescribeSecret",
      "secretsmanager:ListSecretVersionIds",
    ]
    resources = [
      aws_secretsmanager_secret_version.this.arn,
    ]
  }
}

resource "aws_iam_policy" "neon_access_secret" {
  name   = "mybranch-mydb-myrole"
  path   = "/neon/read-only/"
  policy = data.aws_iam_policy_document.neon_access_secret.json
}
```

## Execution

1. Export API keys as environment variables:
   ```commandline
   export NEON_API_KEY=##neon api key##
   export AWS_ACCESS_KEY_ID=##AWS access key ID##
   export AWS_SECRET_ACCESS_KEY=##AWS access secret##
   ```
2. Create the file `main.tf` and copy the code [snippet](#code-snippet) there.
3. Initialize the terraform to download required providers:
   ```commandline
   terraform init
   ```
   ~> Note that the terraform state will be stored locally. Find more
   details [here](https://developer.hashicorp.com/terraform/language/settings/backends/configuration).
4. Validate the syntax correctness:
   ```commandline
   terraform validate
   ```
5. Run terraform plan:
   ```commandline
   terraform plan
   ```
   Expected output in stdout:
   ```commandline
    Terraform used the selected providers to generate the following execution plan. Resource actions are indicated with the following symbols:
      + create
     <= read (data resources)
    
    Terraform will perform the following actions:
    
      # data.aws_iam_policy_document.neon_access_secret will be read during apply
      # (config refers to values not yet known)
     <= data "aws_iam_policy_document" "neon_access_secret" {
          + id   = (known after apply)
          + json = (known after apply)
    
          + statement {
              + actions   = [
                  + "secretsmanager:ListSecrets",
                ]
              + effect    = "Allow"
              + resources = [
                  + "*",
                ]
            }
          + statement {
              + actions   = [
                  + "secretsmanager:DescribeSecret",
                  + "secretsmanager:GetResourcePolicy",
                  + "secretsmanager:GetSecretValue",
                  + "secretsmanager:ListSecretVersionIds",
                ]
              + effect    = "Allow"
              + resources = [
                  + (known after apply),
                ]
            }
        }
    
      # aws_iam_policy.neon_access_secret will be created
      + resource "aws_iam_policy" "neon_access_secret" {
          + arn       = (known after apply)
          + id        = (known after apply)
          + name      = "mybranch-mydb-myrole"
          + path      = "/neon/read-only"
          + policy    = (known after apply)
          + policy_id = (known after apply)
          + tags_all  = (known after apply)
        }
    
      # aws_secretsmanager_secret.this will be created
      + resource "aws_secretsmanager_secret" "this" {
          + arn                            = (known after apply)
          + description                    = "Neon SaaS access details for mydb, myrole @ mybranch"
          + force_overwrite_replica_secret = false
          + id                             = (known after apply)
          + name                           = "neon/mybranch/mydb/myrole"
          + name_prefix                    = (known after apply)
          + policy                         = (known after apply)
          + recovery_window_in_days        = 0
          + rotation_enabled               = (known after apply)
          + rotation_lambda_arn            = (known after apply)
          + tags                           = {
              + "platform" = "neon"
              + "project"  = "demo"
            }
          + tags_all                       = {
              + "platform" = "neon"
              + "project"  = "demo"
            }
    
          + replica {
              + kms_key_id         = (known after apply)
              + last_accessed_date = (known after apply)
              + region             = (known after apply)
              + status             = (known after apply)
              + status_message     = (known after apply)
            }
    
          + rotation_rules {
              + automatically_after_days = (known after apply)
            }
        }
    
      # aws_secretsmanager_secret_version.this will be created
      + resource "aws_secretsmanager_secret_version" "this" {
          + arn            = (known after apply)
          + id             = (known after apply)
          + secret_id      = (known after apply)
          + secret_string  = (sensitive value)
          + version_id     = (known after apply)
          + version_stages = (known after apply)
        }
    
      # neon_branch.this will be created
      + resource "neon_branch" "this" {
          + created_at         = (known after apply)
          + current_state      = (known after apply)
          + endpoint           = (known after apply)
          + host               = (known after apply)
          + id                 = (known after apply)
          + logical_size       = (known after apply)
          + name               = "mybranch"
          + parent_id          = (known after apply)
          + parent_lsn         = (known after apply)
          + parent_timestamp   = (known after apply)
          + pending_state      = (known after apply)
          + physical_size_size = (known after apply)
          + project_id         = (known after apply)
          + updated_at         = (known after apply)
        }
    
      # neon_database.this will be created
      + resource "neon_database" "this" {
          + branch_id  = (known after apply)
          + created_at = (known after apply)
          + id         = (known after apply)
          + name       = "mydb"
          + owner_name = "myrole"
          + project_id = (known after apply)
          + updated_at = (known after apply)
        }
    
      # neon_project.this will be created
      + resource "neon_project" "this" {
          + autoscaling_limit_max_cu  = (known after apply)
          + autoscaling_limit_min_cu  = (known after apply)
          + branch_logical_size_limit = (known after apply)
          + connection_uri            = (sensitive value)
          + cpu_quota_sec             = (known after apply)
          + created_at                = (known after apply)
          + database_host             = (known after apply)
          + database_name             = (known after apply)
          + database_password         = (sensitive value)
          + database_user             = (known after apply)
          + id                        = (known after apply)
          + name                      = "myproject"
          + pg_settings               = (known after apply)
          + pg_version                = (known after apply)
          + region_id                 = (known after apply)
          + updated_at                = (known after apply)
        }
    
      # neon_role.this will be created
      + resource "neon_role" "this" {
          + branch_id  = (known after apply)
          + created_at = (known after apply)
          + id         = (known after apply)
          + name       = "myrole"
          + password   = (sensitive value)
          + project_id = (known after apply)
          + protected  = (known after apply)
          + updated_at = (known after apply)
        }
    
    Plan: 7 to add, 0 to change, 0 to destroy.
   ```
6. Run terraform apply:
   ```commandline
   terraform apply -auto-approve
   ```
   Expected output in stdout:
   ```commandline
   Apply complete! Resources: 7 added, 0 changed, 0 destroyed.
   ```
7. Done! The database can be accessed using the connection details from the AWS secretsmanager.
8. Clean the demo infrastructure:
   ```commandline
   terraform destroy -auto-approve
   ```
   Expected output in stdout:
   ```commandline
   Destroy complete! Resources: 7 destroyed.
   ```
9. Unset environment variables:
   ```commandline
   unset NEON_API_KEY
   unset AWS_ACCESS_KEY_ID
   unset AWS_SECRET_ACCESS_KEY
   ```
