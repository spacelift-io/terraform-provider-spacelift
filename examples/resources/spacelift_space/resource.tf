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
# All callers should compute identical values for description / labels /
# inherit_entities from shared inputs; the resource has standard "last apply
# wins" semantics, and competing stacks would otherwise drift the space on
# every run.
resource "spacelift_space" "team_a" {
  name            = "team-a"
  parent_space_id = "root"
  description     = "Space for team A"
  labels          = ["team:a", "owner:platform"]
  inherit_entities = true

  adopt_existing = true
}