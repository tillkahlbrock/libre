workflow "libre Workflow" {
  on = "push"
  resolves = "Test"
}

action "Test" {
  needs = "Build"
  uses = "actions/action-builder/shell@master"
  runs = "make"
  args = "test"
}

action "Build" {
  uses = "actions/action-builder/shell@master"
  runs = "make"
  args = "build"
}
