package rest_test

import (
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/weplanx/go/rest"
	"testing"
)

func MockSubscribe(t *testing.T, ch chan rest.PublishDto) {
	name := fmt.Sprintf(`%s:events:%s`, service.Namespace, "projects")
	subject := fmt.Sprintf(`%s.events.%s`, service.Namespace, "projects")
	_, err := service.JetStream.QueueSubscribe(subject, name, func(msg *nats.Msg) {
		var data rest.PublishDto
		err := sonic.Unmarshal(msg.Data, &data)
		assert.NoError(t, err)

		ch <- data

		msg.Ack()
	}, nats.ManualAck())

	assert.NoError(t, err)
}

func RemoveStream(t *testing.T) {
	name := fmt.Sprintf(`%s:events:%s`, service.Namespace, "projects")
	err := service.JetStream.DeleteStream(name)
	assert.NoError(t, err)
}

func RecoverStream(t *testing.T) {
	name := fmt.Sprintf(`%s:events:%s`, service.Namespace, "projects")
	subject := fmt.Sprintf(`%s.events.%s`, service.Namespace, "projects")
	_, err := service.JetStream.AddStream(&nats.StreamConfig{
		Name:      name,
		Subjects:  []string{subject},
		Retention: nats.WorkQueuePolicy,
	})
	assert.NoError(t, err)
}
