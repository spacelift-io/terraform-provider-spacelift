package structs

type Action string

const (
	ActionSpaceRead                         Action = "SPACE_READ"
	ActionSpaceWrite                        Action = "SPACE_WRITE"
	ActionSpaceAdmin                        Action = "SPACE_ADMIN"
	ActionRunTrigger                        Action = "RUN_TRIGGER"
	ActionRunTriggerWithCustomRuntimeConfig Action = "RUN_TRIGGER_WITH_CUSTOM_RUNTIME_CONFIG"
	ActionRunConfirm                        Action = "RUN_CONFIRM"
	ActionRunDiscard                        Action = "RUN_DISCARD"
	ActionRunReview                         Action = "RUN_REVIEW"
	ActionRunComment                        Action = "RUN_COMMENT"
	ActionRunTargetedReplan                 Action = "RUN_TARGETED_REPLAN"
	ActionRunPromote                        Action = "RUN_PROMOTE"
	ActionRunPrioritizeSet                  Action = "RUN_PRIORITIZE_SET"
	ActionRunRetry                          Action = "RUN_RETRY"
	ActionRunRetryBlocking                  Action = "RUN_RETRY_BLOCKING"
	ActionRunCancel                         Action = "RUN_CANCEL"
	ActionRunCancelBlocking                 Action = "RUN_CANCEL_BLOCKING"
	ActionRunProposeLocalWorkspace          Action = "RUN_PROPOSE_LOCAL_WORKSPACE"
	ActionRunProposeWithOverrides           Action = "RUN_PROPOSE_WITH_OVERRIDES"
	ActionRunStop                           Action = "RUN_STOP"
	ActionRunStopBlocking                   Action = "RUN_STOP_BLOCKING"
	ActionTaskCreate                        Action = "TASK_CREATE"
	ActionStackCreate                       Action = "STACK_CREATE"
	ActionStackDelete                       Action = "STACK_DELETE"
	ActionStackDisable                      Action = "STACK_DISABLE"
	ActionStackLock                         Action = "STACK_LOCK"
	ActionStackUnlock                       Action = "STACK_UNLOCK"
	ActionStackUnlockForce                  Action = "STACK_UNLOCK_FORCE"
	ActionStackSetCurrentCommit             Action = "STACK_SET_CURRENT_COMMIT"
	ActionStackSyncCommit                   Action = "STACK_SYNC_COMMIT"
	ActionStackDeleteConfig                 Action = "STACK_DELETE_CONFIG"
	ActionStackUpdate                       Action = "STACK_UPDATE"
	ActionStackSetStar                      Action = "STACK_SET_STAR"
	ActionStackAddConfig                    Action = "STACK_ADD_CONFIG"
	ActionStackUploadLocalWorkspace         Action = "STACK_UPLOAD_LOCAL_WORKSPACE"
	ActionStackManagedStateImport           Action = "STACK_MANAGED_STATE_IMPORT"
	ActionStackReslug                       Action = "STACK_RESLUG"
	ActionStackManagedStateRollback         Action = "STACK_MANAGED_STATE_ROLLBACK"
	ActionStackEnable                       Action = "STACK_ENABLE"
	ActionModuleDisable                     Action = "MODULE_DISABLE"
	ActionModuleEnable                      Action = "MODULE_ENABLE"
	ActionModulePublish                     Action = "MODULE_PUBLISH"
	ActionStackManage                       Action = "STACK_MANAGE"
	ActionContextCreate                     Action = "CONTEXT_CREATE"
	ActionContextUpdate                     Action = "CONTEXT_UPDATE"
	ActionContextDelete                     Action = "CONTEXT_DELETE"
	ActionWorkerDrainSet                    Action = "WORKER_DRAIN_SET"
)

var ActionList = []Action{
	ActionSpaceRead,
	ActionSpaceWrite,
	ActionSpaceAdmin,
	ActionRunTrigger,
	ActionRunTriggerWithCustomRuntimeConfig,
	ActionRunConfirm,
	ActionRunDiscard,
	ActionRunReview,
	ActionRunComment,
	ActionRunTargetedReplan,
	ActionRunPromote,
	ActionRunPrioritizeSet,
	ActionRunRetry,
	ActionRunRetryBlocking,
	ActionRunCancel,
	ActionRunCancelBlocking,
	ActionRunProposeLocalWorkspace,
	ActionRunProposeWithOverrides,
	ActionRunStop,
	ActionRunStopBlocking,
	ActionTaskCreate,
	ActionStackCreate,
	ActionStackDelete,
	ActionStackDisable,
	ActionStackLock,
	ActionStackUnlock,
	ActionStackUnlockForce,
	ActionStackSetCurrentCommit,
	ActionStackSyncCommit,
	ActionStackDeleteConfig,
	ActionStackUpdate,
	ActionStackSetStar,
	ActionStackAddConfig,
	ActionStackUploadLocalWorkspace,
	ActionStackManagedStateImport,
	ActionStackReslug,
	ActionStackManagedStateRollback,
	ActionStackEnable,
	ActionModuleDisable,
	ActionModuleEnable,
	ActionModulePublish,
	ActionStackManage,
	ActionContextCreate,
	ActionContextUpdate,
	ActionContextDelete,
	ActionWorkerDrainSet,
}

type Role struct {
	ID                string   `graphql:"id"`
	Slug              string   `graphql:"slug"`
	IsSystem          bool     `graphql:"isSystem"`
	Name              string   `graphql:"name"`
	Description       string   `graphql:"description"`
	Actions           []Action `graphql:"actions"`
	RoleBindingsCount int      `graphql:"roleBindingsCount"`
}
