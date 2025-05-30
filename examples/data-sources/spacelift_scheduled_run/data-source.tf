data "spacelift_scheduled_run" "example" {
  scheduled_run_id = "$STACK_ID/$SCHEDULED_RUN_ID" // id of the scheduled run
}