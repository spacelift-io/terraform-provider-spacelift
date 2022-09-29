data "spacelift_scheduled_delete_stack" "ireland-kubeconfig-delete" {
  scheduled_delete_stack_id = "$STACK_ID/$SCHEDULED_DELETE_STACK_ID" // id of the scheduled delete stack
}