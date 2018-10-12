package controller

import (
	"context"
	"testing"
	"time"

	"github.com/fabric8-services/fabric8-env/app"
	"github.com/fabric8-services/fabric8-env/app/test"
	"github.com/goadesign/goa"
	"github.com/stretchr/testify/assert"
)

func TestShowStatus(t *testing.T) {
	service := goa.New("status-test")
	ctrl := NewStatusController(service)

	_, res := test.ShowStatusOK(t, context.Background(), service, ctrl)

	assert.Equal(t, app.Commit, res.Commit, "Commit is not correct")
	assert.Equal(t, app.StartTime, res.StartTime, "StartTime is not correct")
	_, err := time.Parse("2006-01-02T15:04:05Z", res.StartTime)
	assert.Nil(t, err, "Incorrect layout of StartTime")
}
