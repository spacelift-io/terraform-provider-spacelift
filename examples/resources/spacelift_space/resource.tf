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