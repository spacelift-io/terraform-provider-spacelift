resource "spacelift_stacks" "filtered" {
  administrative {}

  branch {
    any_of = ["main", "master"]
  }

  commit {
    any_of = [
      "4b099200beefdb821c75fdcae2e2cd3e43b75675",
      "6854e36a9cee2f716cc29a34a5067344189a9581",
    ]
  }

  locked {
    equals = false
  }

  labels {
    any_of = ["k8s", "kubernetes"]
  }

  labels {
    all_of = ["core", "production"]
  }

  name {
    any_of = ["k8s-core", "k8s-core-prod"]
  }

  project_root {
    any_of = ["k8s/core", "k8s/core-prod"]
  }

  repo {
    any_of = ["acme/k8s-core", "acme/k8s-core-prod"]
  }

  state {
    any_of = ["FINISHED"]
  }

  vendor {
    any_of = ["Terraform"]
  }
}
