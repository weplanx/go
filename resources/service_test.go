package resources_test

import (
	"github.com/bytedance/sonic"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/weplanx/utils/resources"
	"testing"
)

func MockSubscribe(t *testing.T, ch chan resources.PublishDto) {
	_, err := js.QueueSubscribe("dev.events.projects", "dev:events:projects", func(msg *nats.Msg) {
		var data resources.PublishDto
		err := sonic.Unmarshal(msg.Data, &data)
		assert.NoError(t, err)

		ch <- data

		msg.Ack()
	}, nats.ManualAck())

	assert.NoError(t, err)
}

func RemoveStream(t *testing.T) {
	err := js.DeleteStream("dev:events:projects")
	assert.NoError(t, err)
}

func RecoverStream(t *testing.T) {
	_, err := js.AddStream(&nats.StreamConfig{
		Name:      "dev:events:projects",
		Subjects:  []string{"dev.events.projects"},
		Retention: nats.WorkQueuePolicy,
	})
	assert.NoError(t, err)
}
