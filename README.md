# gin-curd

Provide CURD auxiliary library for gin

[![Github Actions](https://img.shields.io/github/workflow/status/kainonly/gin-curd/tests?style=flat-square)](https://github.com/kainonly/gin-curd/actions)
[![Coveralls github](https://img.shields.io/coveralls/github/kainonly/gin-curd.svg?style=flat-square)](https://coveralls.io/github/kainonly/gin-curd)
[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/kainonly/gin-curd?style=flat-square)](https://github.com/kainonly/gin-curd)
[![Go Report Card](https://goreportcard.com/badge/github.com/kainonly/gin-curd?style=flat-square)](https://goreportcard.com/report/github.com/kainonly/gin-curd)
[![Release](https://img.shields.io/github/v/release/kainonly/gin-curd.svg?style=flat-square)](https://github.com/kainonly/gin-curd)
[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)](https://raw.githubusercontent.com/kainonly/gin-curd/master/LICENSE)

## Setup

Install the curd auxiliary library

```shell script
go get github.com/kainonly/gin-curd
```

## Quick Start

First you need to define gorm

```golang
var db *gorm.DB
var err error

if db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{}); err != nil {
    c.Error(err)
}

curd = curd.Initialize(db)
```

You can also refer to the [lab-api](https://github.com/kain-lab/lab-api) project to use dependency injection initialization

### Operates(operates ...Operator) *execute

Used to plan curd operations, Executable as `Originlists`, `Lists`, `Get`, `Add`, `Edit`, `Delete` results, the following are some examples of [lab-api](https://github.com/kain-lab/lab-api)

#### Originlists() interface{}

Execute origin lists

```golang
type originListsBody struct {
	curd.OriginLists
}

func (c *Controller) OriginLists(ctx *gin.Context) interface{} {
	var body originListsBody
	var err error
	if err = ctx.ShouldBindJSON(&body); err != nil {
		return err
	}
	return c.Curd.Operates(
		curd.Plan(model.Acl{}, body),
	).Originlists()
}
```

#### Lists() interface{}

Execute lists

```golang
type listsBody struct {
	curd.Lists
}

func (c *Controller) Lists(ctx *gin.Context) interface{} {
	var body listsBody
	var err error
	if err = ctx.ShouldBindJSON(&body); err != nil {
		return err
	}
	return c.Curd.Operates(
		curd.Plan(model.Acl{}, body),
	).Lists()
}
```

#### Get() interface{}

Execute get

```golang
type getBody struct {
	curd.Get
}

func (c *Controller) Get(ctx *gin.Context) interface{} {
	var body getBody
	var err error
	if err = ctx.ShouldBindJSON(&body); err != nil {
		return err
	}
	return c.Curd.Operates(
		curd.Plan(model.Acl{}, body),
	).Get()
}
```

#### Add(value interface{}) interface{}

Execute add

- **value** `interface{}` insert data

```golang
type addBody struct {
	Key    string              `binding:"required"`
	Name   datatype.JSONObject `binding:"required"`
	Read   datatype.JSONArray
	Write  datatype.JSONArray
	Status bool
}

func (c *Controller) Add(ctx *gin.Context) interface{} {
	var body addBody
	var err error
	if err = ctx.ShouldBindJSON(&body); err != nil {
		return err
	}
	data := model.Acl{
		Key:    body.Key,
		Name:   body.Name,
		Read:   body.Read,
		Write:  body.Write,
		Status: body.Status,
	}
	return c.Curd.Operates(
		curd.After(func(tx *gorm.DB) error {
			c.clearcache()
			return nil
		}),
	).Add(&data)
}
```

#### Edit(value interface{}) interface{}

Execute edit

- **value** `interface{}` update data

```golang
type editBody struct {
	curd.Edit
	Key    string
	Name   map[string]interface{} `json:"name"`
	Read   datatype.JSONArray     `json:"read"`
	Write  datatype.JSONArray     `json:"write"`
	Status bool
}

func (c *Controller) Edit(ctx *gin.Context) interface{} {
	var body editBody
	var err error
	if err = ctx.ShouldBindJSON(&body); err != nil {
		return err
	}
	data := model.Acl{
		Key:    body.Key,
		Name:   body.Name,
		Read:   body.Read,
		Write:  body.Write,
		Status: body.Status,
	}
	return c.Curd.Operates(
		curd.Plan(model.Acl{}, body),
		curd.After(func(tx *gorm.DB) error {
			c.clearcache()
			return nil
		}),
	).Edit(data)
}
```

#### Delete() interface{}

Execute delete

```golang
type deleteBody struct {
	curd.Delete
}

func (c *Controller) Delete(ctx *gin.Context) interface{} {
	var body deleteBody
	var err error
	if err = ctx.ShouldBindJSON(&body); err != nil {
		return err
	}
	return c.Curd.Operates(
		curd.Plan(model.Acl{}, body),
		curd.After(func(tx *gorm.DB) error {
			c.clearcache()
			return nil
		}),
	).Delete()
}
```

### Option

Global default configuration

- **Orders** `Orders` Default order by, Orders{"id": "desc"}
- **UpdateStatus** `string` Default updated status field, "status"
- **UpdateOmit** `[]string` Default updated exclude fields, []string{"id", "create_time"}

```golang
curd.Set(curd.Option{
    ...
})
```

### Operator

#### Plan(model interface{}, body interface{}) Operator

Plan a model expression

- **model** `interface{}` GORM defined model
- **body** `interface{}` Request body

#### Where(value Conditions) Operator

Set condition array

- **value** `Conditions` Conditions Condition array, **Conditions** `[][]interface{}`

#### SubQuery(fn func(tx *gorm.DB) *gorm.DB) Operator

Set sub query

- **fn** `func(tx *gorm.DB) *gorm.DB`

#### OrderBy(value Orders) Operator

Set order by

- **value** `Orders` Orders, **Orders** `map[string]string`

#### Field(value []string, exclude bool) Operator

Set selecting specific fields

- **value** `[]string` fields
- **exclude** `bool` 

#### Update(status string) Operator

Set update

- **status** `string` When switch is true, update the status field

#### After(fn func(tx *gorm.DB) error) Operator

After hook, when the return is error, the transaction will be rolled back

- **fn** `fn func(tx *gorm.DB) error`

#### Prep(fn func(tx *gorm.DB) error) Operator

Preparation hook, the transaction will be terminated early when the return is error

- **fn** `fn func(tx *gorm.DB) error`

### Body Definition

Used to request structure body embedding, include `go-playground/validator` verification of json

```golang
// General definition of origin list request body
type OriginLists struct {
	Conditions `json:"where" binding:"omitempty,gte=0,dive,len=3,dive,required"`
	Orders     `json:"order" binding:"omitempty,gte=0,dive,keys,endkeys,oneof=asc desc,required"`
}

// General definition of list request body
type Lists struct {

	// Condition array
	Conditions `json:"where" binding:"omitempty,gte=0,dive,len=3,dive,required"`

	// Order by
	Orders `json:"order" binding:"omitempty,gte=0,dive,keys,endkeys,oneof=asc desc,required"`

	// Page definition
	Pagination `json:"page" binding:"required"`
}

// General definition of get request body
type Get struct {

	// Primary key
	Id interface{} `json:"id" binding:"required_without=Conditions"`

	// Condition array
	Conditions `json:"where" binding:"required_without=Id,gte=0,dive,len=3,dive,required"`

	// Order by
	Orders `json:"order" binding:"omitempty,gte=0,dive,keys,endkeys,oneof=asc desc,required"`
}

// General definition of edit request body, choose one of primary key or condition array
type Edit struct {

	// Primary key
	Id interface{} `json:"id" binding:"required_without=Conditions"`

	// Only the status field is updated
	Switch bool `json:"switch"`

	// Condition array
	Conditions `json:"where" binding:"required_without=Id,gte=0,dive,len=3,dive,required"`
}

// General definition of delete request body, choose one of primary key or condition array
type Delete struct {

	// Primary key
	Id interface{} `json:"id" binding:"required_without=Conditions"`

	// Condition array
	Conditions `json:"where" binding:"required_without=Id,gte=0,dive,len=3,dive,required"`
}
```