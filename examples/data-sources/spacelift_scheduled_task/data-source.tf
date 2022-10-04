data "spacelift_scheduled_task" "ireland-kubeconfig-destroy" {
  scheduled_task_id = "$STACK_ID/$SCHEDULED_TASK_ID" // id of the scheduled task
}
