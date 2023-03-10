package dashboardevents

import "github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"

type LeafNodeError struct {
	LeafNode    dashboardtypes.DashboardTreeRun
	Session     string
	Error       error
	ExecutionId string
}

// IsDashboardEvent implements DashboardEvent interface
func (*LeafNodeError) IsDashboardEvent() {}
