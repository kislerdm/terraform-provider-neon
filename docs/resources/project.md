---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "neon_project Resource - terraform-provider-neon"
subcategory: ""
description: |-
  Neon Project. See details: https://neon.tech/docs/get-started-with-neon/setting-up-a-project/
---

# neon_project (Resource)

Neon Project. See details: https://neon.tech/docs/get-started-with-neon/setting-up-a-project/

## Example Usage

```terraform
resource "neon_project" "example" {
  name = "foo"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `instance_handle` (String) Instance type name.
- `name` (String) Project name.
- `platform_id` (String) Platform type id.
- `region_id` (String) AWS Region.
- `settings` (Map of String) Project custom settings.

### Read-Only

- `created_at` (String) Project creation timestamp.
- `current_state` (String) Project current state.
- `databases` (List of Object) List of the project databases. (see [below for nested schema](#nestedatt--databases))
- `deleted` (Boolean) Flag is the project is deleted.
- `id` (String) Project ID.
- `instance_type_id` (String) Instance type ID.
- `max_project_size` (Number) Project max size.
- `parent_id` (String) Project parent.
- `pending_state` (String) Project pending state.
- `platform_name` (String) Platform type name.
- `pooler_enabled` (Boolean) Flag if pooler is enabled.
- `region_name` (String) AWS Region name.
- `roles` (List of Object) List of roles for the project. (see [below for nested schema](#nestedatt--roles))
- `size` (Number) Project size.
- `updated_at` (String) Project last update timestamp.

<a id="nestedatt--databases"></a>
### Nested Schema for `databases`

Read-Only:

- `id` (Number)
- `name` (String)
- `owner_id` (Number)


<a id="nestedatt--roles"></a>
### Nested Schema for `roles`

Read-Only:

- `id` (Number)
- `name` (String)
- `password` (String)


