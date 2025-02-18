package provider

import (
	"context"
	"sync"
	"testing"
	"time"

	neon "github.com/kislerdm/neon-sdk-go"
	"github.com/stretchr/testify/assert"
)

func Test_waitUnfinishedOperations(t *testing.T) {
	operations := []neon.Operation{
		{
			ID:     "0",
			Status: neon.OperationStatusFinished,
		},
		{
			ID:     "1",
			Status: neon.OperationStatusRunning,
		},
		{
			ID:     "2",
			Status: neon.OperationStatusScheduling,
		},
	}

	reader := mockOpsReader{
		rec: make(map[string][]time.Time),
		mu:  new(sync.Mutex),
		maxRequests: map[string]int{
			"0": 0,
			"1": 1,
			"2": 1,
		},
	}
	waitUnfinishedOperations(context.TODO(), reader, operations)
	assert.Nil(t, reader.rec["0"])
	for _, op := range operations[1:] {
		assert.Len(t, reader.rec[op.ID], 2)
		gotDelay := reader.rec[op.ID][1].Sub(reader.rec[op.ID][0])
		assert.GreaterOrEqual(t, gotDelay, operationsCompletionDelay)
	}
}
