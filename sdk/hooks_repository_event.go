package sdk

import (
	"fmt"
	"net/url"
	"strconv"

	"crypto/sha512"
	"encoding/base64"

	"golang.org/x/crypto/pbkdf2"
)

type WorkflowHookEventName string

type WorkflowHookEventType string

func (t WorkflowHookEventType) IsValidForEventName(n WorkflowHookEventName) bool {
	switch n {
	case WorkflowHookEventNamePullRequest:
		switch t {
		case WorkflowHookEventTypePullRequestOpened, WorkflowHookEventTypePullRequestReopened, WorkflowHookEventTypePullRequestClosed, WorkflowHookEventTypePullRequestEdited:
			return true
		}
	case WorkflowHookEventNamePullRequestComment:
		switch t {
		case WorkflowHookEventTypePullRequestCommentCreated, WorkflowHookEventTypePullRequestCommentDeleted, WorkflowHookEventTypePullRequestCommentEdited:
			return true
		}
	default:
		return t == ""
	}
	return false
}

const (
	SignHeaderVCSName   = "X-Cds-Hooks-Vcs-Name"
	SignHeaderRepoName  = "X-Cds-Hooks-Repo-Name"
	SignHeaderVCSType   = "X-Cds-Hooks-Vcs-Type"
	SignHeaderEventName = "X-Cds-Hooks-Event-Name"

	WorkflowHookEventNameWorkflowUpdate WorkflowHookEventName = "workflow-update"
	WorkflowHookEventNameModelUpdate    WorkflowHookEventName = "model-update"
	WorkflowHookEventNamePush           WorkflowHookEventName = "push"
	WorkflowHookEventNameManual         WorkflowHookEventName = "manual"
	WorkflowHookEventNameWebHook        WorkflowHookEventName = "webhook"
	WorkflowHookEventNameWorkflowRun    WorkflowHookEventName = "workflow-run"
	WorkflowHookEventNameScheduler      WorkflowHookEventName = "scheduler"

	WorkflowHookEventNamePullRequest         WorkflowHookEventName = "pull-request"
	WorkflowHookEventTypePullRequestOpened   WorkflowHookEventType = "opened"
	WorkflowHookEventTypePullRequestReopened WorkflowHookEventType = "reopened"
	WorkflowHookEventTypePullRequestClosed   WorkflowHookEventType = "closed"
	WorkflowHookEventTypePullRequestEdited   WorkflowHookEventType = "edited"

	WorkflowHookEventNamePullRequestComment        WorkflowHookEventName = "pull-request-comment"
	WorkflowHookEventTypePullRequestCommentCreated WorkflowHookEventType = "created"
	WorkflowHookEventTypePullRequestCommentDeleted WorkflowHookEventType = "deleted"
	WorkflowHookEventTypePullRequestCommentEdited  WorkflowHookEventType = "edited"

	RepoEventPush = "push"

	HookEventStatusScheduled     = "Scheduled"
	HookEventStatusAnalysis      = "Analyzing"
	HookEventStatusCheckAnalysis = "CheckAnalyzing"
	HookEventStatusWorkflowHooks = "WorkflowHooks"
	HookEventStatusSignKey       = "SignKey"
	HookEventStatusGitInfo       = "GitInfo"
	HookEventStatusWorkflow      = "Workflow"
	HookEventStatusDone          = "Done"
	HookEventStatusError         = "Error"
	HookEventStatusSkipped       = "Skipped"

	HookEventWorkflowStatusScheduled = "Scheduled"
	HookEventWorkflowStatusSkipped   = "Skipped"
	HookEventWorkflowStatusError     = "Error"
	HookEventWorkflowStatusDone      = "Done"
)

type HookEventCallback struct {
	AnalysisCallback   *HookAnalysisCallback `json:"analysis_callback"`
	SigningKeyCallback *Operation            `json:"signing_key_callback"`
	HookEventUUID      string                `json:"hook_event_uuid"`
	HookEventKey       string                `json:"hook_event_key"`
	VCSServerName      string                `json:"vcs_server_name"`
	RepositoryName     string                `json:"repository_name"`
}

type HookAnalysisCallback struct {
	AnalysisID         string           `json:"analysis_id"`
	AnalysisStatus     string           `json:"analysis_status"`
	Error              string           `json:"error"`
	Models             []EntityFullName `json:"models"`
	Workflows          []EntityFullName `json:"workflows"`
	SkippedWorkflows   []EntityFullName `json:"skipped_workflows"`
	SkippedHooks       []V2WorkflowHook `json:"skipped_hooks"`
	DeprecatedUsername string           `json:"username"`
	DeprecatedUserID   string           `json:"user_id"`
	Initiator          *V2Initiator     `json:"initiator"`
}

type HookRepository struct {
	VCSServerType  string `json:"vcs_server_type"`
	VCSServerName  string `json:"vcs_server_name" cli:"vcs_server_name"`
	RepositoryName string `json:"repository_name" cli:"repository_name"`
	Stopped        bool   `json:"stopped" cli:"stopped"`
}

type HookWorkflowRunOutgoingEvent struct {
	UUID                string               `json:"uuid"`
	Created             int64                `json:"created"`
	ProcessingTimestamp int64                `json:"processing_timestamps"`
	LastUpdate          int64                `json:"last_update"`
	Event               HookWorkflowRunEvent `json:"event"`
	Status              string               `json:"status"`
	LastError           string               `json:"last_error"`
	NbErrors            int64                `json:"nb_errors"`
	HooksToTriggers     []HookWorkflowRunOutgoingEventHooks
}

type HookWorkflowRunOutgoingEventHooks struct {
	V2WorkflowHook
	Status                string `json:"status"`
	Error                 string `json:"error"`
	HookRepositoryEventID string `json:"repository_event_id"`
}

func (h *HookWorkflowRunOutgoingEvent) GetFullName() string {
	return fmt.Sprintf("%s/%s/%s/%s/%s", h.Event.WorkflowProject, h.Event.WorkflowVCSServer, h.Event.WorkflowRepository, h.Event.WorkflowName, h.UUID)
}

type HookRepositoryEvent struct {
	UUID                      string                         `json:"uuid"`
	Created                   int64                          `json:"created"`
	EventName                 WorkflowHookEventName          `json:"event_name"` // WorkflowHookEventPush, sdk.WorkflowHookEventPullRequest
	EventType                 WorkflowHookEventType          `json:"event_type"` // created, deleted, edited, opened
	VCSServerName             string                         `json:"vcs_server_name"`
	RepositoryName            string                         `json:"repository_name"`
	Body                      []byte                         `json:"body"`
	ExtractData               HookRepositoryEventExtractData `json:"extracted_data"`
	Status                    string                         `json:"status"`
	ProcessingTimestamp       int64                          `json:"processing_timestamp"`
	LastUpdate                int64                          `json:"last_update"`
	LastError                 string                         `json:"last_error"`
	NbErrors                  int64                          `json:"nb_errors"`
	Analyses                  []HookRepositoryEventAnalysis  `json:"analyses"`
	ModelUpdated              []EntityFullName               `json:"model_updated"`
	WorkflowUpdated           []EntityFullName               `json:"workflow_updated"`
	SkippedWorkflows          []EntityFullName               `json:"skipped_workflows"`
	SkippedHooks              []V2WorkflowHook               `json:"skipped_hooks"`
	WorkflowHooks             []HookRepositoryEventWorkflow  `json:"workflows"`
	DeprecatedUserID          string                         `json:"user_id"`  // Deprecated
	DeprecatedUsername        string                         `json:"username"` // Deprecated
	SignKey                   string                         `json:"sign_key"`
	SigningKeyOperation       string                         `json:"signing_key_operation"`
	SigningKeyOperationStatus OperationStatus                `json:"signing_key_operation_status"`
	Initiator                 *V2Initiator                   `json:"initiator"`
}

func (h *HookRepositoryEvent) IsTerminated() bool {
	return h.Status == HookEventStatusDone || h.Status == HookEventStatusError || h.Status == HookEventStatusSkipped
}

func (h *HookRepositoryEvent) ToInsightReport(uiURL string) VCSInsight {
	report := VCSInsight{
		Title:  "CDS",
		Detail: fmt.Sprintf("Event %q (%s): %s", h.EventName, h.UUID, h.Status),
		Datas:  make([]VCSInsightData, 0),
	}
	if h.Status != HookEventStatusDone {
		report.Detail += "\n\n" + h.LastError
	} else {
		for _, a := range h.Analyses {
			if a.Error != "" {
				report.Detail += fmt.Sprintf("\n\nOn project %s: %s", a.ProjectKey, a.Error)
			}
		}
	}

	// Check if there is max 5 analysis to display all of them
	if len(h.Analyses) < 6 {
		for _, a := range h.Analyses {
			report.Datas = append(report.Datas, VCSInsightData{
				Title: "Analysis on " + a.ProjectKey,
				Type:  "LINK",
				Text:  a.Status,
				Href:  fmt.Sprintf("%s/project/%s/explore/vcs/%s/repository/%s/settings", uiURL, a.ProjectKey, h.VCSServerName, url.PathEscape(h.RepositoryName)),
			})
		}

		// Else only create 1 data for the first one
	} else if len(h.Analyses) > 0 {
		a := h.Analyses[0]
		report.Datas = append(report.Datas, VCSInsightData{
			Title: "Analysis on " + a.ProjectKey,
			Type:  "LINK",
			Text:  a.Status,
			Href:  fmt.Sprintf("%s/project/%s/explore/vcs/%s/repository/%s/settings", uiURL, a.ProjectKey, h.VCSServerName, url.PathEscape(h.RepositoryName)),
		})
		// If no analysis
	} else {
		report.Datas = append(report.Datas, VCSInsightData{
			Title: "Nb of Analysis triggered",
			Type:  "TEXT",
			Text:  strconv.Itoa(0),
		})
	}

	// Check if workflows have been triggered
	if len(h.WorkflowHooks)+len(report.Datas) < 7 {
		for _, w := range h.WorkflowHooks {
			da := VCSInsightData{
				Title: fmt.Sprintf("%s #%d", w.WorkflowName, w.RunNumber),
				Type:  "LINK",
				Text:  w.Status,
			}

			if w.Status == string(V2WorkflowRunStatusSuccess) {
				da.Href = fmt.Sprintf("%s/project/%s/run/%s", uiURL, w.ProjectKey, w.RunID)
			} else {
				da.Href = fmt.Sprintf("%s/project/%s/explore/vcs/%s/repository/%s/settings", uiURL, w.ProjectKey, h.VCSServerName, url.PathEscape(h.RepositoryName))
			}
			report.Datas = append(report.Datas, da)
		}
		// Else display 1 data for all workflows + link to the run view with filter
	} else {
		repoFilter := fmt.Sprintf("%s/%s", h.VCSServerName, h.RepositoryName)

		w := h.WorkflowHooks[0]
		da := VCSInsightData{
			Title: "Nb of workflows triggered",
			Type:  "LINK",
			Text:  strconv.Itoa(len(h.WorkflowHooks)),
			Href:  fmt.Sprintf("%s/project/%s/run?workflow_ref=%s&repository=%s", uiURL, w.ProjectKey, url.PathEscape(w.Ref), url.PathEscape(repoFilter)),
		}
		report.Datas = append(report.Datas, da)
	}
	return report
}

type HookRepositoryEventWorkflow struct {
	ProjectKey           string             `json:"project_key"`
	VCSIdentifier        string             `json:"vcs_identifier"`
	RepositoryIdentifier string             `json:"repository_identifier"`
	WorkflowName         string             `json:"workflow_name"`
	EntityID             string             `json:"entity_id"`
	Ref                  string             `json:"ref"`
	Commit               string             `json:"commit"`
	Type                 string             `json:"type"`
	Status               string             `json:"status"`
	Error                string             `json:"error"`
	TargetCommit         string             `json:"target_commit,omitempty"`
	ModelFullName        string             `json:"model,omitempty"`
	PathFilters          []string           `json:"path_filters,omitempty"`
	Data                 V2WorkflowHookData `json:"data,omitempty"`
	Initiator            *V2Initiator       `json:"initiator,omitempty"`

	// Workflow run result
	RunID     string `json:"run_id,omitempty"`
	RunNumber int64  `json:"run_number,omitempty"`

	// Git info to be able to start a new workflow run
	SemverCurrent string   `json:"semver_current"`
	SemverNext    string   `json:"semver_next"`
	UpdatedFiles  []string `json:"updated_files"`

	// Operation data to get gitInfo
	OperationUUID   string          `json:"operation_uuid"`
	OperationStatus OperationStatus `json:"operation_status"`
	OperationError  string          `json:"operation_error"`
	LastCheck       int64           `json:"last_check"`
}

func (wh *HookRepositoryEventWorkflow) IsTerminated() bool {
	return wh.Status == HookEventWorkflowStatusError || wh.Status == HookEventWorkflowStatusSkipped || wh.Status == HookEventWorkflowStatusDone
}

type HookRepositoryEventExtractData struct {
	CDSEventName       WorkflowHookEventName                       `json:"cds_event_name"`
	CDSEventType       WorkflowHookEventType                       `json:"cds_event_type"`
	Commit             string                                      `json:"commit"`
	CommitFrom         string                                      `json:"commit_from"`
	CommitMessage      string                                      `json:"commit_message"`
	CommitAuthor       string                                      `json:"commit_author,omitempty"`
	CommitAuthorEmail  string                                      `json:"commit_author_email,omitempty"`
	Paths              []string                                    `json:"paths"`
	Ref                string                                      `json:"ref"`
	PullRequestID      int64                                       `json:"pullrequest_id"`
	PullRequestRefTo   string                                      `json:"pullrequest_ref_to"`
	Manual             HookRepositoryEventExtractedDataManual      `json:"manual"`
	DeprecatedAdminMFA bool                                        `json:"admin_mfa"` // Deprecated
	Scheduler          HookRepositoryEventExtractedDataScheduler   `json:"scheduler"`
	WorkflowRun        HookRepositoryEventExtractedDataWorkflowRun `json:"workflow_run"`
	WebHook            HookRepositoryEventExtractedDataWebHook     `json:"workflow_hook,omitempty"`
	HookProjectKey     string                                      `json:"hook_project_key"` // force the hook to only trigger from the given CDS project
}

type HookRepositoryEventExtractedDataWebHook struct {
	Project    string `json:"project"`
	VCS        string `json:"vcs"`
	Repository string `json:"repository"`
	Workflow   string `json:"workflow"`
	ID         string `json:"id"`
}

type HookRepositoryEventExtractedDataManual struct {
	Project      string `json:"project"`
	Workflow     string `json:"workflow"`
	TargetCommit string `json:"target_commit"`
	TargetBranch string `json:"target_branch"`
	TargetTag    string `json:"target_tag"`
}

type HookRepositoryEventExtractedDataWorkflowRun struct {
	Project               string `json:"project"`
	Workflow              string `json:"workflow"`
	WorkflowRunID         string `json:"workflow_run_id"`
	TargetVCS             string `json:"target_vcs"`
	TargetRepository      string `json:"target_repository"`
	OutgoingHookEventUUID string `json:"outgoing_hook_event_uuid"`
}

type HookRepositoryEventExtractedDataScheduler struct {
	TargetVCS      string `json:"target_vcs"`
	TargetRepo     string `json:"target_repo"`
	TargetWorkflow string `json:"target_workflow"`
	TargetProject  string `json:"target_project"`
	Cron           string `json:"cron"`
	Timezone       string `json:"timezone"`
}

type GeneratedWebhook struct {
	Key           string `json:"key"`
	UUID          string `json:"uuid"`
	HookPublicURL string `json:"url"`
}

func (h *HookRepositoryEvent) GetFullName() string {
	return fmt.Sprintf("%s/%s/%s", h.VCSServerName, h.RepositoryName, h.UUID)
}

type HookRepositoryEventAnalysis struct {
	AnalyzeID      string `json:"analyze_id"`
	Status         string `json:"status"`
	ProjectKey     string `json:"project_key"`
	Error          string `json:"error"`
	FindRetryCount int64  `json:"find_retry_count"`
}

type HookRetrieveSignKeyRequest struct {
	ProjectKey            string `json:"projectKey"`
	VCSServerName         string `json:"vcs_server_name"`
	RepositoryName        string `json:"repository_name"`
	Commit                string `json:"commit"`
	Ref                   string `json:"ref"`
	HookEventUUID         string `json:"hook_event_uuid"`
	HookEventKey          string `json:"hook_event_key"`
	GetSigninKey          bool   `json:"get_signin_key"`
	ChangesetsCommitSince string `json:"changesets_commit"`
	GetChangesets         bool   `json:"get_changesets"`
	GetSemver             bool   `json:"get_semver"`
	GetCommitMessage      bool   `json:"commit_message"`
}

type HookRetrieveUserRequest struct {
	ProjectKey     string `json:"projectKey"`
	VCSServerName  string `json:"vcs_server_name"`
	RepositoryName string `json:"repository_name"`
	Commit         string `json:"commit"`
	SignKey        string `json:"sign_key"`
	HookEventUUID  string `json:"hook_event_uuid"`
}

type HookRetrieveUserResponse struct {
	DeprecatedUserID   string       `json:"user_id"`  // Deprecated
	DeprecatedUsername string       `json:"username"` // Deprecated
	Initiator          *V2Initiator `json:"initiator"`
}

type AnalysisRequest struct {
	ProjectKey         string       `json:"projectKey"`
	VcsName            string       `json:"vcsName"`
	RepoName           string       `json:"repoName"`
	Ref                string       `json:"ref"`
	Commit             string       `json:"commit"`
	HookEventUUID      string       `json:"hook_event_uuid"`
	HookEventKey       string       `json:"hook_event_key"`
	DeprecatedUserID   string       `json:"user_id"`
	DeprecatedAdminMFA bool         `json:"admin_mfa"`
	Initiator          *V2Initiator `json:"initiator"`
}

type AnalysisResponse struct {
	AnalysisID string `json:"analysis_id" cli:"analysis_id"`
	Status     string `json:"status" cli:"status"`
}

func GenerateRepositoryWebHookSecret(hookSecretKey, pkey, vcsName, repoName, uuid string) string {
	pass := fmt.Sprintf("%s-%s-%s-%s", pkey, vcsName, repoName, uuid)
	keyBytes := pbkdf2.Key([]byte(pass), []byte(hookSecretKey), 4096, 128, sha512.New)
	key64 := base64.StdEncoding.EncodeToString(keyBytes)
	return key64
}

func GenerateWorkflowWebHookSecret(hookSecretKey, pkey, vcsName, repoName, workflowName, uuid string) string {
	pass := fmt.Sprintf("%s-%s-%s-%s-%s", pkey, vcsName, repoName, workflowName, uuid)
	keyBytes := pbkdf2.Key([]byte(pass), []byte(hookSecretKey), 4096, 128, sha512.New)
	key64 := base64.StdEncoding.EncodeToString(keyBytes)
	return key64
}
