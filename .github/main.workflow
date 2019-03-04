workflow "libre Workflow" {
  on = "push"
  resolves = "Test"
}

action "Test" {
  needs = "Build"
  uses = "./.github/actions/go-build"
  runs = "make"
  args = "test"
}

action "Build" {
  uses = "./.github/actions/go-build"
  runs = "make"
  args = "build"
}
