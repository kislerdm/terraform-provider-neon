package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	neon "github.com/kislerdm/neon-sdk-go"
)

type opsReader interface {
	GetProjectOperation(projectID string, operationID string) (neon.OperationResponse, error)
}

const operationsCompletionDelay = 100 * time.Millisecond

func waitUnfinishedOperations(ctx context.Context, c opsReader, ops []neon.Operation) {
	var unfinishedOps = make([]neon.Operation, 0, len(ops))
	for _, op := range ops {
		if unfinishedOperation(op) {
			unfinishedOps = append(unfinishedOps, op)
		}
	}

	maxN := len(unfinishedOps)
	var flags = make(chan struct{}, maxN)

	for _, op := range unfinishedOps {
		go func(op neon.Operation) {
			var finished bool
			for !finished {
				tflog.Trace(ctx, "wait for unfinished operation", map[string]interface{}{
					"projectID":   op.ProjectID,
					"operationID": op.ID,
				})
				time.Sleep(operationsCompletionDelay)
				resp, err := c.GetProjectOperation(op.ProjectID, op.ID)
				if err != nil {
					tflog.Error(ctx, "error getting operation status", map[string]interface{}{
						"projectID":   op.ProjectID,
						"operationID": op.ID,
						"error":       err,
					})
				} else {
					finished = !unfinishedOperation(resp.Operation)
				}
			}
			flags <- struct{}{}
		}(op)
	}

	for maxN > 0 {
		<-flags
		maxN--
	}
}

func unfinishedOperation(op neon.Operation) bool {
	var o bool
	switch op.Status {
	case neon.OperationStatusRunning, neon.OperationStatusScheduling:
		o = true
	}
	return o
}
