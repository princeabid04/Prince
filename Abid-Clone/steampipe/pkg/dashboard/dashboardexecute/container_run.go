package dashboardexecute

import (
	"context"
	"fmt"

	"github.com/turbot/steampipe/pkg/dashboard/dashboardevents"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

// DashboardContainerRun is a struct representing a container run
type DashboardContainerRun struct {
	DashboardParentImpl

	dashboardNode *modconfig.DashboardContainer
}

func (r *DashboardContainerRun) AsTreeNode() *dashboardtypes.SnapshotTreeNode {
	res := &dashboardtypes.SnapshotTreeNode{
		Name:     r.Name,
		NodeType: r.NodeType,
		Children: make([]*dashboardtypes.SnapshotTreeNode, len(r.children)),
	}
	for i, c := range r.children {
		res.Children[i] = c.AsTreeNode()
	}
	return res
}

func NewDashboardContainerRun(container *modconfig.DashboardContainer, parent dashboardtypes.DashboardParent, executionTree *DashboardExecutionTree) (*DashboardContainerRun, error) {
	children := container.GetChildren()

	r := &DashboardContainerRun{
		DashboardParentImpl: DashboardParentImpl{
			DashboardTreeRunImpl: NewDashboardTreeRunImpl(container, parent, executionTree),
		},

		dashboardNode: container,
	}
	if container.Title != nil {
		r.Title = *container.Title
	}

	if container.Width != nil {
		r.Width = *container.Width
	}
	r.childComplete = make(chan dashboardtypes.DashboardTreeRun, len(children))
	for _, child := range children {
		var childRun dashboardtypes.DashboardTreeRun
		var err error
		switch i := child.(type) {
		case *modconfig.DashboardContainer:
			childRun, err = NewDashboardContainerRun(i, r, executionTree)
			if err != nil {
				return nil, err
			}
		case *modconfig.Dashboard:
			childRun, err = NewDashboardRun(i, r, executionTree)
			if err != nil {
				return nil, err
			}
		case *modconfig.Benchmark, *modconfig.Control:
			childRun, err = NewCheckRun(i.(modconfig.DashboardLeafNode), r, executionTree)
			if err != nil {
				return nil, err
			}

		default:
			// ensure this item is a DashboardLeafNode
			leafNode, ok := i.(modconfig.DashboardLeafNode)
			if !ok {
				return nil, fmt.Errorf("child %s does not implement DashboardLeafNode", i.Name())
			}

			childRun, err = NewLeafRun(leafNode, r, executionTree)
			if err != nil {
				return nil, err
			}
		}

		// should never happen - container children must be either container or counter
		if childRun == nil {
			continue
		}

		// if our child has not completed, we have not completed
		if childRun.GetRunStatus() == dashboardtypes.DashboardRunReady {
			r.Status = dashboardtypes.DashboardRunReady
		}
		r.children = append(r.children, childRun)
	}
	// add r into execution tree
	executionTree.runs[r.Name] = r
	return r, nil
}

// IsSnapshotPanel implements SnapshotPanel
func (*DashboardContainerRun) IsSnapshotPanel() {}

// Initialise implements DashboardRunNode
func (r *DashboardContainerRun) Initialise(ctx context.Context) {
	// initialise our children
	if err := r.initialiseChildren(ctx); err != nil {
		r.SetError(ctx, err)
	}
}

// Execute implements DashboardRunNode
// execute all children and wait for them to complete
func (r *DashboardContainerRun) Execute(ctx context.Context) {
	// execute all children asynchronously
	for _, child := range r.children {
		go child.Execute(ctx)
	}

	// wait for children to complete
	var errors []error
	for !r.ChildrenComplete() {
		completeChild := <-r.childComplete
		if completeChild.GetRunStatus() == dashboardtypes.DashboardRunError {
			errors = append(errors, completeChild.GetError())
		}
		// fall through to recheck ChildrenComplete
	}

	// so all children have completed - check for errors
	err := error_helpers.CombineErrors(errors...)
	if err == nil {
		// set complete status on dashboard
		r.SetComplete(ctx)
	} else {
		r.SetError(ctx, err)
	}
}

// SetError implements DashboardTreeRun
// tell parent we are done
func (r *DashboardContainerRun) SetError(_ context.Context, err error) {
	r.err = err
	// error type does not serialise to JSON so copy into a string
	r.ErrorString = err.Error()
	r.Status = dashboardtypes.DashboardRunError
	// raise container error event
	r.executionTree.workspace.PublishDashboardEvent(&dashboardevents.ContainerError{
		Container:   r,
		Session:     r.executionTree.sessionId,
		ExecutionId: r.executionTree.id,
	})
	r.parent.ChildCompleteChan() <- r
}

// SetComplete implements DashboardTreeRun
func (r *DashboardContainerRun) SetComplete(context.Context) {
	r.Status = dashboardtypes.DashboardRunComplete
	// raise container complete event
	r.executionTree.workspace.PublishDashboardEvent(&dashboardevents.ContainerComplete{
		Container:   r,
		Session:     r.executionTree.sessionId,
		ExecutionId: r.executionTree.id,
	})
	// tell parent we are done
	r.parent.ChildCompleteChan() <- r
}