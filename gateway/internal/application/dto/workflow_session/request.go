package workflowsession

type StartWorkflowRequest struct {
	WorkflowConfigID uint64 `json:"workflow_config_id" binding:"required"`
}