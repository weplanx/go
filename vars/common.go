package vars

import (
	"github.com/google/wire"
	"github.com/thoas/go-funk"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var Provides = wire.NewSet(
	wire.Struct(new(Controller), "*"),
	wire.Struct(new(Service), "*"),
)

var Key = "vars"

type Var struct {
	ID    primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	Key   string             `json:"key" bson:"key"`
	Value interface{}        `json:"value" bson:"value"`
}

var Secrets = []string{
	"tencent_secret_key",
	"tencent_pulsar_token",
	"feishu_app_secret",
	"feishu_encrypt_key",
	"feishu_verification_token",
	"email_password",
	"openapi_secret",
}

func SecretText(key string) bool {
	return funk.Contains(Secrets, key)
}
