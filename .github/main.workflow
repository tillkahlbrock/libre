workflow "libre Workflow" {
  on = "push"
  resolves = "Test"
}

action "Test" {
  needs = "Build"
  uses = "./.github/actions/go-build"
  args = "make test"
}

action "Build" {
  uses = "./.github/actions/go-build"
  args = "make build"
}
