package gauges

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIdleSessions(t *testing.T) {
	var assert = assert.New(t)
	var db = connect(t)
	defer db.Close()
	var metrics = evaluate(t, IdleSessions(db, labels))
	assert.Len(metrics, 1)
	assertGreaterThan(t, -1, metrics[0])
}