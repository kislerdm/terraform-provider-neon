terraform {
  required_providers {
    neon = {
      source  = "kislerdm/neon"
      version = ">= 0.0.9"
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

resource "neon_branch" "this" {
  project_id = neon_project.this.id
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
  name       = "mydb"
  owner_name = neon_role.this.name
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
