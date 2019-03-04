workflow "libre Workflow" {
  on = "push"
}

action "Test" {
  uses = "actions/action-builder/shell@master"
  runs = "make"
  args = "test"
}

action "Build" {
  uses = "actions/action-builder/shell@master"
  runs = "make"
  args = "build"
}