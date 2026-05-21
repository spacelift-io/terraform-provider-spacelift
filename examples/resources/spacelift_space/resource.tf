resource "spacelift_space" "development" {
  name = "development"

  # Every account has a root space that serves as the root for the space tree.
  # Except for the root space, all the other spaces must define their parents.
  parent_space_id = "root"

  # An optional description of a space.
  description = "This a child of the root space. It contains all the resources common to the development infrastructure."
}


resource "spacelift_space" "development-frontend" {
  name = "development-frontend"

  # This space will be a child of the development space.
  parent_space_id = spacelift_space.development.id

  # An optional value, that gives this space a read access to all the entities that it's parent has access to.
  inherit_entities = true
}

# Shared spaces across multiple stacks.
#
# When the same logical space is declared from several Terraform configurations
# (for example, a shared module instantiated by many admin stacks), set
# `adopt_existing = true` so that the second and subsequent applies pick up the
# row created by the first one instead of erroring or creating duplicates.
#
# Two important semantics to keep in mind:
#
#   1. All callers should compute identical values for description / labels /
#      inherit_entities from shared inputs. Updates do propagate to the backend
#      (last apply wins), so competing stacks with different inputs will drift
#      the space on every run.
#
#   2. `terraform destroy` on an adopting resource only removes it from the
#      local Terraform state — the backend space stays. This is intentional:
#      with many stacks adopting the same space, a single destroy would either
#      fail with "space has dependants" or strand every other stack's state. To
#      actually delete the space, take ownership from a single configuration
#      with `adopt_existing = false`, then destroy.
resource "spacelift_space" "team_a" {
  name             = "team-a"
  parent_space_id  = "root"
  description      = "Space for team A"
  labels           = ["team:a", "owner:platform"]
  inherit_entities = true

  adopt_existing = true
}