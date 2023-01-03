data "spacelift_contexts" "contexts" {
  labels {
    any_of = ["foo", "bar"]
  }

  labels {
    any_of = ["baz"]
  }
}
