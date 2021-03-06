package jira

import (
	"context"
	"fmt"
	"net/http"
)

// EpicsService handles communication with the epic related
// methods of the Jira Agile API
//
// Jira Agile API docs: https://docs.atlassian.com/jira-software/REST/7.3.1/#agile/1.0/epic
type EpicsService service

// EpicWrap represents the data returned by the API,
// in addition to the board information, paging data is returned
type EpicWrap struct {
	Pagination
	Values []*Epic `json:"values,omitempty"`
}

// Epic represents a Jira Agile Epic
type Epic struct {
	ID       int               `json:"id,omitempty"`
	Key      string            `json:"key,omitempty"`
	Name     string            `json:"name,omitempty"`
	Summary  string            `json:"summary,omitempty"`
	SelfLink string            `json:"self,omitempty"`
	Done     bool              `json:"done,omitempty"`
	Color    map[string]string `json:"color,omitempty"`
}

// EpicRank contains the fields for ranking epics
type EpicRank struct {
	RankAfter         string `json:"rankAfterEpic,omitempty"`
	RankBefore        string `json:"rankBeforeEpic,omitempty"`
	RankCustomFieldID string `json:"rankCustomFieldId,omitempty"`
}

// EpicsOptions contains all options to list all epics from the board
type EpicsOptions struct {
	//The starting index of the returned epics. Base index: 0. See the 'Pagination' section at the top of this page for more details.
	StartAt int `query:"startAt"`
	//The maximum number of epics to return per page. Default: 50. See the 'Pagination' section at the top of this page for more details.
	MaxResults int `query:"maxResults"`
	//Filters results to epics that are either done or not done. Valid values: true, false.
	Done bool `query:"done"`
}

// Get returns the epic for a given epic Id.
// This epic will only be returned if the user has permission to view it.
//
// GET /rest/agile/1.0/epic/{epicIdOrKey}
func (e *EpicsService) Get(ctx context.Context, idOrKey string) (*Epic, *Response, error) {

	req, err := e.client.NewRequest("GET", fmt.Sprintf("epic/%s", idOrKey), nil)
	if err != nil {
		return nil, nil, err
	}

	var epic = &Epic{}
	resp, err := e.client.Do(ctx, req, epic)
	if err != nil {
		return nil, resp, err
	}

	return epic, resp, nil
}

// ListIssues returns all issues that belong to the epic, for the given epic Id. This only includes
// issues that the user has permission to view. Issues returned from this resource include Agile
// fields, like sprint, closedSprints, flagged, and epic. By default, the returned issues are
// ordered by rank.
//
// GET /rest/agile/1.0/epic/{epicIdOrKey}/issue
func (e *EpicsService) ListIssues(ctx context.Context, idOrKey string, opts *IssuesOptions) ([]*Issue, *Response, error) {

	q := QueryParameters(opts)

	req, err := e.client.NewRequest("GET", fmt.Sprintf("epic/%s/issue%s", idOrKey, q), nil)
	if err != nil {
		return nil, nil, err
	}

	var wrap = &IssueWrap{}
	resp, err := e.client.Do(ctx, req, wrap)
	if err != nil {
		return nil, resp, err
	}

	resp.MaxResults = wrap.MaxResults
	resp.StartAt = wrap.StartAt
	resp.IsLast = wrap.IsLast

	return wrap.Values, resp, nil
}

// PartiallyUpdate performs a partial update of the epic. A partial update means that fields not present
// in the request JSON will not be updated. Valid values for color are color_1 to color_9.
//
// POST /rest/agile/1.0/epic/{epicIdOrKey}
func (e *EpicsService) PartiallyUpdate(ctx context.Context, idOrKey string, epic *Epic) (*Epic, *Response, error) {
	req, err := e.client.NewRequest("POST", fmt.Sprintf("epic/%s", idOrKey), epic)
	if err != nil {
		return nil, nil, err
	}

	var updatedEpic = &Epic{}
	resp, err := e.client.Do(ctx, req, updatedEpic)
	if err != nil {
		return nil, resp, err
	}

	return updatedEpic, resp, nil
}

// MoveIssuesTo moves issues to an epic, for a given epic id. Issues can be only in a single epic
// at the same time. That means that already assigned issues to an epic, will not be assigned to
// the previous epic anymore. The user needs to have the edit issue permission for all issue
// they want to move and to the epic. The maximum number of issues that can be moved in one
// operation is 50.
//
// POST /rest/agile/1.0/epic/{epicIdOrKey}/issue
func (e *EpicsService) MoveIssuesTo(ctx context.Context, idOrKey string, issueKeys *IssueKeys) (bool, *Response, error) {
	req, err := e.client.NewRequest("POST", fmt.Sprintf("epic/%s/issue", idOrKey), issueKeys)
	if err != nil {
		return false, nil, err
	}

	resp, err := e.client.Do(ctx, req, nil)
	if err != nil {
		return false, resp, err
	}

	if resp.StatusCode == http.StatusNoContent {
		return true, resp, nil
	}

	return false, resp, nil
}

// ListIssuesWithoutEpic returns all issues that do not belong to any epic. This only includes issues
// that the user has permission to view. Issues returned from this resource include Agile fields,
// like sprint, closedSprints, flagged, and epic. By default, the returned issues are ordered by rank.
//
// GET /rest/agile/1.0/epic/none/issue
func (e *EpicsService) ListIssuesWithoutEpic(ctx context.Context, opts *IssuesOptions) ([]*Issue, *Response, error) {

	q := QueryParameters(opts)

	req, err := e.client.NewRequest("GET", "epic/none/issue"+q, nil)
	if err != nil {
		return nil, nil, err
	}

	var wrap = &IssueWrap{}
	resp, err := e.client.Do(ctx, req, wrap)
	if err != nil {
		return nil, resp, err
	}

	resp.MaxResults = wrap.MaxResults
	resp.StartAt = wrap.StartAt
	resp.IsLast = wrap.IsLast

	return wrap.Values, resp, nil
}

// RemoveIssuesFrom removes issues from epics. The user needs to have the edit issue permission for
// all issue they want to remove from epics. The maximum number of issues that can be moved in one
// operation is 50.
//
// POST /rest/agile/1.0/epic/none/issue
func (e *EpicsService) RemoveIssuesFrom(ctx context.Context, issueKeys *IssueKeys) (bool, *Response, error) {
	return e.MoveIssuesTo(ctx, "none", issueKeys)
}

// Rank moves (ranks) an epic before or after a given epic.
// If rankCustomFieldId is not defined, the default rank field will be used.
//
// PUT /rest/agile/1.0/epic/{epicIdOrKey}/rank
func (e *EpicsService) Rank(ctx context.Context, idOrKey string, rank *EpicRank) (bool, *Response, error) {

	req, err := e.client.NewRequest("PUT", fmt.Sprintf("epic/%s/rank", idOrKey), rank)
	if err != nil {
		return false, nil, err
	}

	resp, err := e.client.Do(ctx, req, nil)
	if err != nil {
		return false, resp, err
	}

	if resp.StatusCode == http.StatusNoContent {
		return true, resp, nil
	}

	return false, resp, nil
}
