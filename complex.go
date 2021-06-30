package bit

import "github.com/gin-gonic/gin"

const context = "complex"

type complexVar struct {
	body interface{}
	data interface{}
}

type Operator func(*complexVar)

func SetBody(body interface{}) Operator {
	return func(c *complexVar) {
		c.body = body
	}
}

func SetData(data interface{}) Operator {
	return func(c *complexVar) {
		c.data = data
	}
}

func Complex(c *gin.Context, operator ...Operator) {
	v := new(complexVar)
	for _, operator := range operator {
		operator(v)
	}
	c.Set(context, v)
}
