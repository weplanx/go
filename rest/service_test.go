package rest_test

import (
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/weplanx/go/rest"
	"testing"
)

func TestMorePipe(t *testing.T) {
	input := M{
		"data": M{
			"date": "abc",
			"dates": []interface{}{
				"a",
				"b",
			},
			"timestamps": []interface{}{
				"a",
				"b",
			},
			"metadata": nil,
			"items": []interface{}{
				M{"sn": "123"},
				M{"sn": "456"},
			},
		},
	}
	err := service.Pipe(input, []string{"data", "date"}, "date")
	assert.Error(t, err)
	err = service.Pipe(input, []string{"data", "dates"}, "dates")
	assert.Error(t, err)
	err = service.Pipe(input, []string{"data", "timestamps"}, "timestamps")
	assert.Error(t, err)
	err = service.Pipe(input, []string{"data", "metadata", "$", "v"}, "password")
	assert.NoError(t, err)
	err = service.Pipe(input, []string{"data", "items", "$", "sn"}, "date")
	assert.Error(t, err)
	err = service.Pipe(input, []string{"data", "unkown"}, "date")
	assert.NoError(t, err)
}

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
