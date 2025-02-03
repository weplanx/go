package help

import (
	"github.com/google/uuid"
	"github.com/sony/sonyflake"
	"strconv"
)

func Uuid() string {
	return uuid.New().String()
}

var SF = sonyflake.NewSonyflake(sonyflake.Settings{})

func SID() string {
	var id uint64
	id, _ = SF.NextID()
	return strconv.FormatUint(id, 10)
}
