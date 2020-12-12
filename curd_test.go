package curd_test

import (
	"errors"
	jsoniter "github.com/json-iterator/go"
	curd "github.com/kainonly/gin-curd"
	"github.com/kainonly/gin-curd/model"
	. "gopkg.in/check.v1"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"os"
	"testing"
)

var err error

func Test(t *testing.T) { TestingT(t) }

type MySuite struct {
	curd *curd.Curd
	db   *gorm.DB
}

var _ = Suite(&MySuite{})

func (s *MySuite) SetUpTest(c *C) {
	dsn := os.Getenv("DSN")
	if s.db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{}); err != nil {
		c.Error(err)
	}
	s.curd = curd.Initialize(s.db)
	if err = s.db.Migrator().DropTable(&model.Example{}); err != nil {
		c.Error(err)
	}
	if err = s.db.AutoMigrate(&model.Example{}); err != nil {
		c.Error(err)
	}
	data := []model.Example{
		{KeyId: "main", Name: "Common Module"},
		{KeyId: "resource", Name: "Resource Module"},
		{KeyId: "acl", Name: "Acl Module"},
		{KeyId: "policy", Name: "Policy Module"},
		{KeyId: "admin", Name: "Admin Module"},
		{KeyId: "role", Name: "Role Module"},
	}
	if err = s.db.Create(&data).Error; err != nil {
		c.Error(err)
	}
	s.curd.Set(curd.Option{
		Orders:       curd.Orders{"id": "desc", "create_time": "desc"},
		UpdateStatus: "status",
		UpdateOmit:   []string{"id", "create_time"},
	})
}

type originListsBody struct {
	curd.OriginLists
}

func (s *MySuite) TestOriginLists(c *C) {
	var body originListsBody
	if err = jsoniter.Unmarshal(
		[]byte(`{"where":[["name","like","%R%"]]}`),
		&body,
	); err != nil {
		c.Error(err)
	}
	result := s.curd.Operates(
		curd.Plan(&model.Example{}, body),
		curd.OrderBy(curd.Orders{"id": "asc"}),
		curd.Field([]string{"create_time", "update_time"}, true),
	).Originlists()
	c.Log(result)
	c.Assert(result, FitsTypeOf, make([]map[string]interface{}, 0))
	lists := result.([]map[string]interface{})
	c.Assert(lists, HasLen, 2)
	keys := make([]string, len(lists))
	for index, data := range lists {
		keys[index] = data["key_id"].(string)
	}
	c.Assert(keys, DeepEquals, []string{"resource", "role"})
}

func (s *MySuite) TestOriginListsQuery(c *C) {
	var body originListsBody
	if err = jsoniter.Unmarshal(
		[]byte(`{}`),
		&body,
	); err != nil {
		c.Error(err)
	}
	result := s.curd.Operates(
		curd.Plan(&model.Example{}, body),
		curd.SubQuery(func(tx *gorm.DB) *gorm.DB {
			tx = tx.Where("key_id = ?", "admin")
			return tx
		}),
	).Originlists()
	c.Log(result)
	c.Assert(result, FitsTypeOf, make([]map[string]interface{}, 0))
	lists := result.([]map[string]interface{})
	c.Assert(lists, HasLen, 1)
	keys := make([]string, len(lists))
	for index, data := range lists {
		keys[index] = data["key_id"].(string)
	}
	c.Assert(keys, DeepEquals, []string{"admin"})
}

type listsBody struct {
	curd.Lists
}

func (s *MySuite) TestLists(c *C) {
	var body listsBody
	if err = jsoniter.Unmarshal(
		[]byte(`{"page":{"index":2,"limit":2},"order":{"id":"asc"}}`),
		&body,
	); err != nil {
		c.Error(err)
	}
	result := s.curd.Operates(
		curd.Plan(&model.Example{}, body),
		curd.Field([]string{"id", "key_id", "name"}, false),
	).Lists()
	c.Log(result)
	c.Assert(result, FitsTypeOf, make(map[string]interface{}))
	data := result.(map[string]interface{})
	c.Assert(data["total"], Equals, int64(6))
	c.Assert(data["lists"], FitsTypeOf, make([]map[string]interface{}, 0))
	c.Assert(data["lists"], HasLen, 2)
	lists := data["lists"].([]map[string]interface{})
	keys := make([]string, len(lists))
	for index, data := range lists {
		keys[index] = data["key_id"].(string)
	}
	c.Assert(keys, DeepEquals, []string{"acl", "policy"})
}

func (s *MySuite) TestListsQuery(c *C) {
	var body listsBody
	if err = jsoniter.Unmarshal(
		[]byte(`{"page":{"index":1,"limit":10},"order":{"id":"asc"}}`),
		&body,
	); err != nil {
		c.Error(err)
	}
	result := s.curd.Operates(
		curd.Plan(&model.Example{}, body),
		curd.SubQuery(func(tx *gorm.DB) *gorm.DB {
			tx = tx.Where("key_id in ?", []string{"resource", "acl", "policy"})
			return tx
		}),
	).Lists()
	c.Log(result)
	c.Assert(result, FitsTypeOf, make(map[string]interface{}))
	data := result.(map[string]interface{})
	c.Assert(data["total"], Equals, int64(3))
	c.Assert(data["lists"], FitsTypeOf, make([]map[string]interface{}, 0))
	c.Assert(data["lists"], HasLen, 3)
	lists := data["lists"].([]map[string]interface{})
	keys := make([]string, len(lists))
	for index, data := range lists {
		keys[index] = data["key_id"].(string)
	}
	c.Assert(keys, DeepEquals, []string{"resource", "acl", "policy"})
}

type getBody struct {
	curd.Get
}

func (s *MySuite) TestGet(c *C) {
	var body getBody
	if err = jsoniter.Unmarshal(
		[]byte(`{"id":1}`),
		&body,
	); err != nil {
		c.Error(err)
	}
	result := s.curd.Operates(
		curd.Plan(&model.Example{}, body),
	).Get()
	c.Log(result)
	c.Assert(result, FitsTypeOf, make(map[string]interface{}))
	data := result.(map[string]interface{})
	c.Assert(data["key_id"].(string), Equals, "main")
}

func (s *MySuite) TestGetWithoutId(c *C) {
	var body getBody
	if err = jsoniter.Unmarshal(
		[]byte(`{"where":[["key_id","=","main"]]}`),
		&body,
	); err != nil {
		c.Error(err)
	}
	result := s.curd.Operates(
		curd.Plan(&model.Example{}, body),
	).Get()
	c.Log(result)
	c.Assert(result, FitsTypeOf, make(map[string]interface{}))
	data := result.(map[string]interface{})
	c.Assert(data["id"].(uint64), Equals, uint64(1))
}

func (s *MySuite) TestGetQuery(c *C) {
	var body getBody
	if err = jsoniter.Unmarshal(
		[]byte(`{"where":[["key_id","=","main"]]}`),
		&body,
	); err != nil {
		c.Error(err)
	}
	result := s.curd.Operates(
		curd.Plan(&model.Example{}, body),
		curd.SubQuery(func(tx *gorm.DB) *gorm.DB {
			tx = tx.Where("status = ?", true)
			return tx
		}),
	).Get()
	c.Log(result)
	c.Assert(result, FitsTypeOf, make(map[string]interface{}))
	data := result.(map[string]interface{})
	c.Assert(data["id"].(uint64), Equals, uint64(1))
}

func (s *MySuite) TestGetQualification(c *C) {
	var body getBody
	if err = jsoniter.Unmarshal(
		[]byte(`{"where":[["key_id","=","main"]]}`),
		&body,
	); err != nil {
		c.Error(err)
	}
	result := s.curd.Operates(
		curd.Plan(&model.Example{}, body),
		curd.Where(curd.Conditions{
			{"status", "=", true},
		}),
	).Get()
	c.Log(result)
	c.Assert(result, FitsTypeOf, make(map[string]interface{}))
	data := result.(map[string]interface{})
	c.Assert(data["id"].(uint64), Equals, uint64(1))
}

func (s *MySuite) TestAdd(c *C) {
	data := &model.Example{
		KeyId:  "test",
		Name:   "Test Module",
		Status: true,
	}
	success := s.curd.Operates().Add(&data)
	c.Assert(success, Equals, true)
	var expect model.Example
	s.db.Model(&model.Example{}).Where("key_id = ?", "test").First(&expect)
	c.Assert(expect.Name, Equals, "Test Module")
	duplicate := s.curd.Operates().Add(&data)
	c.Assert(duplicate.(error), NotNil)
}

func (s *MySuite) TestAddAfter(c *C) {
	data1 := &model.Example{
		KeyId:  "test",
		Name:   "Test Module",
		Status: true,
	}
	success := s.curd.Operates(
		curd.After(func(tx *gorm.DB) error {
			c.Log("after success")
			return nil
		}),
	).Add(&data1)
	c.Assert(success, Equals, true)
	var expect model.Example
	s.db.Model(&model.Example{}).Where("key_id = ?", "test").First(&expect)
	c.Assert(expect.Name, Equals, "Test Module")
	duplicate := s.curd.Operates(
		curd.After(func(tx *gorm.DB) error {
			c.Log("after success")
			return nil
		}),
	).Add(&data1)
	c.Assert(duplicate.(error), NotNil)
	data2 := &model.Example{
		KeyId:  "failed",
		Name:   "Failed Module",
		Status: true,
	}
	failed := s.curd.Operates(
		curd.After(func(tx *gorm.DB) error {
			return errors.New("after failed")
		}),
	).Add(&data2)
	c.Assert(failed.(error), NotNil)
}

type editBody struct {
	KeyId  string `json:"key_id"`
	Name   string `json:"name"`
	Status bool   `json:"status"`
	curd.Edit
}

func (s *MySuite) TestEdit(c *C) {
	var body editBody
	var expect model.Example
	if err = jsoniter.Unmarshal(
		[]byte(`{"id":1,"key_id":"main","name":"Update Module","status":true,"switch":false}`),
		&body,
	); err != nil {
		c.Error(err)
	}
	data1 := model.Example{
		KeyId:  body.KeyId,
		Name:   body.Name,
		Status: body.Status,
	}
	result1 := s.curd.Operates(
		curd.Plan(&model.Example{}, body),
	).Edit(data1)
	c.Assert(result1, Equals, true)
	s.db.Model(&model.Example{}).Where("id = ?", 1).First(&expect)
	c.Assert(expect.Name, Equals, "Update Module")

	data1.KeyId = "acl"
	duplicate := s.curd.Operates(
		curd.Plan(&model.Example{}, body),
	).Edit(data1)
	c.Assert(duplicate.(error), NotNil)

	if err = jsoniter.Unmarshal(
		[]byte(`{"id":1,"key_id":"main_update","switch":false}`),
		&body,
	); err != nil {
		c.Error(err)
	}
	data2 := model.Example{
		KeyId: body.KeyId,
	}
	result2 := s.curd.Operates(
		curd.Plan(&model.Example{}, body),
		curd.Field([]string{"key_id"}, false),
	).Edit(data2)
	c.Assert(result2, Equals, true)
	s.db.Model(&model.Example{}).Where("id = ?", 1).First(&expect)
	c.Assert(expect.KeyId, Equals, "main_update")

	if err = jsoniter.Unmarshal(
		[]byte(`{"id":1,"key_id":"main_next","switch":false}`),
		&body,
	); err != nil {
		c.Error(err)
	}
	data3 := model.Example{
		KeyId: body.KeyId,
	}
	result3 := s.curd.Operates(
		curd.Plan(&model.Example{}, body),
		curd.Field([]string{"id", "name", "status", "create_time"}, true),
	).Edit(data3)
	c.Assert(result3, Equals, true)
	s.db.Model(&model.Example{}).Where("id = ?", 1).First(&expect)
	c.Assert(expect.KeyId, Equals, "main_next")

}

func (s *MySuite) TestEditStatus(c *C) {
	var body editBody
	var expect model.Example
	if err = jsoniter.Unmarshal(
		[]byte(`{"id":1,"status":false,"switch":true}`),
		&body,
	); err != nil {
		c.Error(err)
	}
	data1 := model.Example{
		KeyId:  body.KeyId,
		Name:   body.Name,
		Status: body.Status,
	}
	result1 := s.curd.Operates(
		curd.Plan(&model.Example{}, body),
	).Edit(data1)
	c.Assert(result1, Equals, true)
	s.db.Model(&model.Example{}).Where("id = ?", 1).First(&expect)
	c.Assert(expect.Status, Equals, false)

	if err = jsoniter.Unmarshal(
		[]byte(`{"id":1,"status":true,"switch":true}`),
		&body,
	); err != nil {
		c.Error(err)
	}
	data2 := model.Example{
		KeyId:  body.KeyId,
		Name:   body.Name,
		Status: body.Status,
	}
	result2 := s.curd.Operates(
		curd.Plan(&model.Example{}, body),
		curd.Update("status"),
	).Edit(data2)
	c.Assert(result2, Equals, true)
	s.db.Model(&model.Example{}).Where("id = ?", 1).First(&expect)
	c.Assert(expect.Status, Equals, true)
}

func (s *MySuite) TestEditQuery(c *C) {
	var body editBody
	var expect model.Example
	if err = jsoniter.Unmarshal(
		[]byte(`{"id":1,"key_id":"main","name":"Update Module","status":true,"switch":false}`),
		&body,
	); err != nil {
		c.Error(err)
	}
	data := model.Example{
		KeyId:  body.KeyId,
		Name:   body.Name,
		Status: body.Status,
	}
	result := s.curd.Operates(
		curd.Plan(&model.Example{}, body),
		curd.SubQuery(func(tx *gorm.DB) *gorm.DB {
			return tx
		}),
	).Edit(data)
	c.Assert(result, Equals, true)
	s.db.Model(&model.Example{}).Where("id = ?", 1).First(&expect)
	c.Assert(expect.Name, Equals, "Update Module")
}

func (s *MySuite) TestEditAfter(c *C) {
	var body editBody
	var expect model.Example
	if err = jsoniter.Unmarshal(
		[]byte(`{"id":1,"key_id":"main","name":"Update Module","status":true,"switch":false}`),
		&body,
	); err != nil {
		c.Error(err)
	}
	data := model.Example{
		KeyId:  body.KeyId,
		Name:   body.Name,
		Status: body.Status,
	}
	result := s.curd.Operates(
		curd.Plan(&model.Example{}, body),
		curd.After(func(tx *gorm.DB) error {
			c.Log("after success")
			return nil
		}),
	).Edit(data)
	c.Assert(result, Equals, true)
	s.db.Model(&model.Example{}).Where("id = ?", 1).First(&expect)
	c.Assert(expect.Name, Equals, "Update Module")

	data.KeyId = "acl"
	duplicate := s.curd.Operates(
		curd.Plan(&model.Example{}, body),
		curd.After(func(tx *gorm.DB) error {
			c.Log("after success")
			return nil
		}),
	).Edit(data)
	c.Assert(duplicate.(error), NotNil)
	data.KeyId = "test"
	failed := s.curd.Operates(
		curd.Plan(&model.Example{}, body),
		curd.After(func(tx *gorm.DB) error {
			return errors.New("after failed")
		}),
	).Edit(data)
	c.Assert(failed.(error), NotNil)
}

type deleteBody struct {
	curd.Delete
}

func (s *MySuite) TestDelete(c *C) {
	var body deleteBody
	var count int64
	if err = jsoniter.Unmarshal(
		[]byte(`{"id":[1,2]}`),
		&body,
	); err != nil {
		c.Error(err)
	}
	result := s.curd.Operates(
		curd.Plan(&model.Example{}, body),
	).Delete()
	c.Log(result)
	c.Assert(result, Equals, true)
	s.db.Model(&model.Example{}).Where("id in ?", []uint64{1, 2}).Count(&count)
	c.Assert(count, Equals, int64(0))
}

func (s *MySuite) TestDeleteWithoutId(c *C) {
	var body deleteBody
	var count int64
	if err = jsoniter.Unmarshal(
		[]byte(`{"where":[["key_id","=","acl"]]}`),
		&body,
	); err != nil {
		c.Error(err)
	}
	result := s.curd.Operates(
		curd.Plan(&model.Example{}, body),
	).Delete()
	c.Log(result)
	c.Assert(result, Equals, true)
	s.db.Model(&model.Example{}).Where("key_id = ?", "acl").Count(&count)
	c.Assert(count, Equals, int64(0))

	failed := s.curd.Operates(
		curd.Plan(&model.Example{}, body),
		curd.Where(curd.Conditions{
			{"unknow", "=", "abcd"},
		}),
	).Delete()
	c.Assert(failed.(error), NotNil)
}

func (s *MySuite) TestDeleteQuery(c *C) {
	var body deleteBody
	var count int64
	if err = jsoniter.Unmarshal(
		[]byte(`{"id":[1,2]}`),
		&body,
	); err != nil {
		c.Error(err)
	}
	result := s.curd.Operates(
		curd.Plan(&model.Example{}, body),
		curd.SubQuery(func(tx *gorm.DB) *gorm.DB {
			tx = tx.Or("id = ?", 3)
			return tx
		}),
	).Delete()
	c.Log(result)
	c.Assert(result, Equals, true)
	s.db.Model(&model.Example{}).Where("id in ?", []uint64{1, 2, 3}).Count(&count)
	c.Assert(count, Equals, int64(0))
}

func (s *MySuite) TestDeleteHook(c *C) {
	var body deleteBody
	var count int64
	if err = jsoniter.Unmarshal(
		[]byte(`{"id":[1,2]}`),
		&body,
	); err != nil {
		c.Error(err)
	}
	result := s.curd.Operates(
		curd.Plan(&model.Example{}, body),
		curd.Prep(func(tx *gorm.DB) error {
			c.Log("prep success")
			return nil
		}),
		curd.After(func(tx *gorm.DB) error {
			c.Log("after success")
			return nil
		}),
	).Delete()
	c.Assert(result, Equals, true)
	s.db.Model(&model.Example{}).Where("id in ?", []uint64{1, 2}).Count(&count)
	c.Assert(count, Equals, int64(0))

}

func (s *MySuite) TestDeleteHookFailed(c *C) {
	var body deleteBody
	if err = jsoniter.Unmarshal(
		[]byte(`{"where":[]}`),
		&body,
	); err != nil {
		c.Error(err)
	}
	failed := s.curd.Operates(
		curd.Plan(&model.Example{}, body),
		curd.Where(curd.Conditions{
			{"unknow", "=", "abcd"},
		}),
		curd.Prep(func(tx *gorm.DB) error {
			c.Log("prep success")
			return nil
		}),
		curd.After(func(tx *gorm.DB) error {
			c.Log("after success")
			return nil
		}),
	).Delete()
	c.Log(failed)
	c.Assert(failed.(error), NotNil)
}

func (s *MySuite) TestDeletePrepHookFailed(c *C) {
	var body deleteBody
	if err = jsoniter.Unmarshal(
		[]byte(`{"id":[1,2]}`),
		&body,
	); err != nil {
		c.Error(err)
	}
	failed := s.curd.Operates(
		curd.Plan(&model.Example{}, body),
		curd.Prep(func(tx *gorm.DB) error {
			return errors.New("prep failed")
		}),
		curd.After(func(tx *gorm.DB) error {
			c.Log("after success")
			return nil
		}),
	).Delete()
	c.Log(failed)
	c.Assert(failed.(error), NotNil)
}

func (s *MySuite) TestDeleteAfterHookFailed(c *C) {
	var body deleteBody
	if err = jsoniter.Unmarshal(
		[]byte(`{"id":[1,2]}`),
		&body,
	); err != nil {
		c.Error(err)
	}
	failed := s.curd.Operates(
		curd.Plan(&model.Example{}, body),
		curd.Prep(func(tx *gorm.DB) error {
			c.Log("prep success")
			return nil
		}),
		curd.After(func(tx *gorm.DB) error {
			return errors.New("after failed")
		}),
	).Delete()
	c.Log(failed)
	c.Assert(failed.(error), NotNil)
}
