package api

import (
	"path"
	"time"
)

// CreateGroupResponse is a service response returned when a new group is
// created. For now it's just the group ID, but may be expanded in the future.
type CreateGroupResponse struct {
	ID string `json:"id"`
}

// GroupSpec is a specification for creating a new Group.
type GroupSpec struct {
	// (required) Unique name to assign the group.
	Name string `json:"name"`

	// (optional) Text description for the dataset.
	Description string `json:"description,omitempty"`

	// (optional) Initial set of experiments to add to the group.
	Experiments []string `json:"experiments,omitempty"`

	// (optional) A token representing the user to which the object should be attributed.
	// If omitted attribution will be given to the user issuing request.
	AuthorToken string `json:"author_token,omitempty"`
}

// Group is a collection of experiments.
type Group struct {
	ID          string    `json:"id"`
	User        User      `json:"user"`
	Name        string    `json:"name,omitempty"`
	Description string    `json:"description,omitempty"`
	Created     time.Time `json:"created"`
	Modified    time.Time `json:"modified"`
}

// DisplayID returns the most human-friendly name available for a group
// while guaranteeing that it's unique and non-empty.
func (e *Group) DisplayID() string {
	if e.Name != "" {
		return path.Join(e.User.Name, e.Name)
	}
	return e.ID
}

// GroupExperimentTask identifies an (experiment, task) pair within a group.
type GroupExperimentTask struct {
	Experiment GroupExperiment `json:"experiment"`
	Task       GroupTask       `json:"task"`
}

// GroupExperiment is a minimal experiment summary for aggregated views.
type GroupExperiment struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
	User User   `json:"user"`
}

// GroupTask is a minimal task summary for aggregated views.
type GroupTask struct {
	ID      string                 `json:"id"`
	Status  TaskStatus             `json:"status"`
	Metrics map[string]interface{} `json:"metrics,omitempty"`
	Env     map[string]string      `json:"env,omitempty"`
	Name    string                 `json:"name,omitempty"`
}

// GroupPatchSpec describes a patch to apply to a group's editable fields.
type GroupPatchSpec struct {
	// (optional) Unqualified name to assign to the group. It is considered
	// a collision error if another group has the same creator and name.
	Name *string `json:"name,omitempty"`

	// (optional) Description to assign to the group or empty string to
	// delete an existing description.
	Description *string `json:"description,omitempty"`

	// (optional) Experiment IDs to add to the group.
	// It is an error to add and remove the same experiment in one patch.
	AddExperiments []string `json:"add_experiments,omitempty"`

	// (optional) Experiment IDs to remove from the group.
	// It is an error to add and remove the same experiment in one patch.
	RemoveExperiments []string `json:"remove_experiments,omitempty"`

	// (optional) New selected environment variables and metrics.
	Parameters *[]GroupParameter `json:"parameters,omitempty"`
}

// GroupParameter is a measurable value for use in group analyses.
type GroupParameter struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

// GroupParameterCount summarizes how often a parameter is observed among a group's tasks.
type GroupParameterCount struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Count int64  `json:"count"`
}
