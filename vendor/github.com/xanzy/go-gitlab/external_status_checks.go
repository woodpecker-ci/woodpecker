package gitlab

import (
	"fmt"
	"net/http"
	"time"
)

// ExternalStatusChecksService handles communication with the external
// status check related methods of the GitLab API.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/status_checks.html
type ExternalStatusChecksService struct {
	client *Client
}

type MergeStatusCheck struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	ExternalURL string `json:"external_url"`
	Status      string `json:"status"`
}

type ProjectStatusCheck struct {
	ID                int                          `json:"id"`
	Name              string                       `json:"name"`
	ProjectID         int                          `json:"project_id"`
	ExternalURL       string                       `json:"external_url"`
	ProtectedBranches []StatusCheckProtectedBranch `json:"protected_branches"`
}

type StatusCheckProtectedBranch struct {
	ID                        int        `json:"id"`
	ProjectID                 int        `json:"project_id"`
	Name                      string     `json:"name"`
	CreatedAt                 *time.Time `json:"created_at"`
	UpdatedAt                 *time.Time `json:"updated_at"`
	CodeOwnerApprovalRequired bool       `json:"code_owner_approval_required"`
}

// ListMergeStatusChecks lists the external status checks that apply to it
// and their status for a single merge request.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/status_checks.html#list-status-checks-for-a-merge-request
func (s *ExternalStatusChecksService) ListMergeStatusChecks(pid interface{}, mr int, opt *ListOptions, options ...RequestOptionFunc) ([]*MergeStatusCheck, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/merge_requests/%d/status_checks", PathEscape(project), mr)

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var mscs []*MergeStatusCheck
	resp, err := s.client.Do(req, &mscs)
	if err != nil {
		return nil, resp, err
	}

	return mscs, resp, err
}

// ListProjectStatusChecks lists the project external status checks.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/status_checks.html#get-project-external-status-checks
func (s *ExternalStatusChecksService) ListProjectStatusChecks(pid interface{}, opt *ListOptions, options ...RequestOptionFunc) ([]*ProjectStatusCheck, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/external_status_checks", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var pscs []*ProjectStatusCheck
	resp, err := s.client.Do(req, &pscs)
	if err != nil {
		return nil, resp, err
	}

	return pscs, resp, err
}
