package test

//// 创建文档不合规的 URL
//func TestCreateUrlError(t *testing.T) {
//	res := httptest.NewRecorder()
//	req, _ := http.NewRequest("POST", "/Privileges", nil)
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 400, res.Code)
//}
//
//// 创建文档不合规的请求体
//func TestCreateBodyError(t *testing.T) {
//	res := httptest.NewRecorder()
//	req, _ := http.NewRequest("POST", "/privileges", nil)
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 400, res.Code)
//}
//
//var createInsertedID string
//
//// 创建文档
//func TestCreate(t *testing.T) {
//	res := httptest.NewRecorder()
//	body, err := jsoniter.Marshal(CreateBody{
//		Doc: map[string]interface{}{
//			"name": "agent",
//		},
//	})
//	if err != nil {
//		panic(err)
//	}
//	req, _ := http.NewRequest("POST", "/privileges", bytes.NewBuffer(body))
//
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 201, res.Code)
//
//	var result map[string]interface{}
//	if err := jsoniter.Unmarshal(res.Body.Bytes(), &result); err != nil {
//		panic(err)
//	}
//	count, err := db.Collection("privileges").CountDocuments(context.TODO(), bson.M{
//		"name": "agent",
//	})
//	if err != nil {
//		t.Error(err)
//	}
//	assert.Equal(t, int64(1), count)
//	createInsertedID = result["InsertedID"].(string)
//}
//
//// 创建文档不合规的引用
//func TestCreateRefError(t *testing.T) {
//	res := httptest.NewRecorder()
//	body, err := jsoniter.Marshal(CreateBody{
//		Doc: map[string]interface{}{
//			"privileges": []string{"abc", "d1"},
//			"name":       "Kenny Boyer",
//			"account":    "Lyda_Mosciski",
//			"email":      "Lempi.Larkin60@yahoo.com",
//			"phone":      "(403) 332-1896 x64468",
//			"address":    "924 Braulio Radial",
//		},
//		Ref: []string{"privileges"},
//	})
//	if err != nil {
//		panic(err)
//	}
//	req, _ := http.NewRequest("POST", "/example", bytes.NewBuffer(body))
//
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 400, res.Code)
//}
//
//// 创建文档包含引用
//func TestCreateWithRef(t *testing.T) {
//	res := httptest.NewRecorder()
//	body, err := jsoniter.Marshal(CreateBody{
//		Doc: map[string]interface{}{
//			"privileges": []string{createInsertedID},
//			"name":       "Kenny Boyer",
//			"account":    "Lyda_Mosciski",
//			"email":      "Lempi.Larkin60@yahoo.com",
//			"phone":      "(403) 332-1896 x64468",
//			"address":    "924 Braulio Radial",
//		},
//		Ref: []string{"privileges", "nothing"},
//	})
//	if err != nil {
//		panic(err)
//	}
//	req, _ := http.NewRequest("POST", "/example", bytes.NewBuffer(body))
//
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 201, res.Code)
//
//	var result map[string]interface{}
//	oid, _ := primitive.ObjectIDFromHex(createInsertedID)
//	if err = db.Collection("example").FindOne(context.TODO(), bson.M{
//		"privileges": bson.M{"$in": bson.A{oid}},
//	}).Decode(&result); err != nil {
//		t.Error(err)
//	}
//	assert.Equal(t, "Kenny Boyer", result["name"])
//}
//
//// 创建文档格式转换
//func TestCreateFormat(t *testing.T) {
//	res := httptest.NewRecorder()
//	body, err := jsoniter.Marshal(CreateBody{
//		Doc: map[string]interface{}{
//			"name":     "admin",
//			"alias":    "61f7ef84dfdb15138a09cdad",
//			"password": "adx8090",
//		},
//		Format: map[string]interface{}{
//			"alias":    "object_id",
//			"password": "password",
//		},
//	})
//	if err != nil {
//		panic(err)
//	}
//	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
//
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 201, res.Code)
//
//	var result map[string]interface{}
//	if err = db.Collection("users").FindOne(context.TODO(), bson.M{
//		"name": "admin",
//	}).Decode(&result); err != nil {
//		t.Error(err)
//	}
//	assert.Nil(t, password.Verify("adx8090", result["password"].(string)))
//	assert.True(t, true, result["alias"].(primitive.ObjectID))
//}
//
//// 创建文档格式转换不存在字段忽略
//func TestCreateFormatIgnore(t *testing.T) {
//	res := httptest.NewRecorder()
//	body, err := jsoniter.Marshal(CreateBody{
//		Doc: map[string]interface{}{
//			"name": "agent",
//		},
//		Format: map[string]interface{}{
//			"parent": "object_id",
//		},
//	})
//	if err != nil {
//		t.Error(err)
//	}
//	req, _ := http.NewRequest("POST", "/privileges", bytes.NewBuffer(body))
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 201, res.Code)
//}
//
//// 创建文档格式转换错误
//func TestCreateFormatError(t *testing.T) {
//	res := httptest.NewRecorder()
//	body, err := jsoniter.Marshal(CreateBody{
//		Doc: map[string]interface{}{
//			"name": "agent",
//		},
//		Format: map[string]interface{}{
//			"name": "object_id",
//		},
//	})
//	if err != nil {
//		t.Error(err)
//	}
//	req, _ := http.NewRequest("POST", "/privileges", bytes.NewBuffer(body))
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 400, res.Code)
//}
//
//// 创建多个文档
//func TestCreateMany(t *testing.T) {
//	res := httptest.NewRecorder()
//	body, err := jsoniter.Marshal(CreateBody{Docs: mock})
//	if err != nil {
//		panic(err)
//	}
//	req, _ := http.NewRequest("POST", "/services", bytes.NewBuffer(body))
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 201, res.Code)
//	count, err := db.Collection("services").
//		CountDocuments(context.TODO(), bson.M{})
//	if err != nil {
//		t.Error(err)
//	}
//	assert.Equal(t, int64(50), count)
//}
//
//// 创建多个文档不合规的引用
//func TestCreateManyRefError(t *testing.T) {
//	res := httptest.NewRecorder()
//	body, err := jsoniter.Marshal(CreateBody{
//		Docs: []map[string]interface{}{
//			{
//				"privileges": []string{"abc", "d1"},
//			},
//		},
//		Ref: []string{"privileges"},
//	})
//	if err != nil {
//		t.Error(err)
//	}
//	req, _ := http.NewRequest("POST", "/privileges", bytes.NewBuffer(body))
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 400, res.Code)
//}
//
//// 创建多个文档格式转换错误
//func TestCreateManyFormatError(t *testing.T) {
//	res := httptest.NewRecorder()
//	body, err := jsoniter.Marshal(CreateBody{
//		Docs: []map[string]interface{}{
//			{
//				"name": "agent",
//			},
//		},
//		Format: map[string]interface{}{
//			"name": "object_id",
//		},
//	})
//	if err != nil {
//		t.Error(err)
//	}
//	req, _ := http.NewRequest("POST", "/privileges", bytes.NewBuffer(body))
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 400, res.Code)
//}
//
//// 获取所有文档不合规的 URL
//func TestFindUrlError(t *testing.T) {
//	res := httptest.NewRecorder()
//	req, _ := http.NewRequest("GET", "/Services", nil)
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 400, res.Code)
//}
//
//// 获取所有文档
//func TestFind(t *testing.T) {
//	res := httptest.NewRecorder()
//	req, _ := http.NewRequest("GET", "/services", nil)
//
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 200, res.Code)
//
//	var data []map[string]interface{}
//	if err := jsoniter.Unmarshal(res.Body.Bytes(), &data); err != nil {
//		t.Error(err)
//	}
//	a, b := funk.Difference(
//		funk.Map(mock, func(x map[string]interface{}) string {
//			return x["name"].(string)
//		}),
//		funk.Map(data, func(x map[string]interface{}) string {
//			return x["name"].(string)
//		}),
//	)
//	assert.Empty(t, a)
//	assert.Empty(t, b)
//}
//
//// 获取多个文档不合规的查询
//func TestFindWithWhereError(t *testing.T) {
//	res := httptest.NewRecorder()
//	where, err := jsoniter.Marshal(map[string]interface{}{
//		"$x": "",
//	})
//	if err != nil {
//		t.Error(err)
//	}
//	req, _ := http.NewRequest("GET", "/services", nil)
//	query := req.URL.Query()
//	query.Add("where", string(where))
//	req.URL.RawQuery = query.Encode()
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 400, res.Code)
//}
//
//// 获取多个文档不合规的排序
//func TestFindWithSortError(t *testing.T) {
//	res := httptest.NewRecorder()
//	req, _ := http.NewRequest("GET", "/services", nil)
//	query := req.URL.Query()
//	query.Add("sort", "price.2")
//	req.URL.RawQuery = query.Encode()
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 400, res.Code)
//}
//
//var findWithWhereData []map[string]interface{}
//
//// 获取多个文档（过滤）
//func TestFindWithWhere(t *testing.T) {
//	res := httptest.NewRecorder()
//	where, err := jsoniter.Marshal(map[string]interface{}{
//		"number": map[string]interface{}{"$in": []string{"55826199", "57277117"}},
//	})
//	if err != nil {
//		t.Error(err)
//	}
//	req, _ := http.NewRequest("GET", "/services", nil)
//	query := req.URL.Query()
//	query.Add("where", string(where))
//	query.Add("sort", "price.1")
//	req.URL.RawQuery = query.Encode()
//
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 200, res.Code)
//
//	if err := jsoniter.Unmarshal(res.Body.Bytes(), &findWithWhereData); err != nil {
//		t.Error(err)
//	}
//	a, b := funk.Difference(
//		[]string{"Handmade Soft Salad", "Intelligent Fresh Shoes"},
//		funk.Map(findWithWhereData, func(x map[string]interface{}) string {
//			return x["name"].(string)
//		}),
//	)
//	assert.Empty(t, a)
//	assert.Empty(t, b)
//}
//
//// 获取多个文档（ID）
//func TestFindWithId(t *testing.T) {
//	res := httptest.NewRecorder()
//	req, _ := http.NewRequest("GET", "/services", nil)
//	query := req.URL.Query()
//	for _, x := range findWithWhereData {
//		query.Add("id", x["_id"].(string))
//	}
//	req.URL.RawQuery = query.Encode()
//
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 200, res.Code)
//
//	var data []map[string]interface{}
//	if err := jsoniter.Unmarshal(res.Body.Bytes(), &data); err != nil {
//		t.Error(err)
//	}
//	a, b := funk.Difference(
//		[]float64{float64(727), float64(47)},
//		funk.Map(data, func(x map[string]interface{}) float64 {
//			return x["price"].(float64)
//		}),
//	)
//	assert.Empty(t, a)
//	assert.Empty(t, b)
//}
//
//// 获取分页文档不合规的请求头部
//func TestFindPageHeaderError(t *testing.T) {
//	res := httptest.NewRecorder()
//	req, _ := http.NewRequest("GET", "/services", nil)
//	req.Header.Set("x-page-size", "5")
//	req.Header.Set("x-page", "-1")
//
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 400, res.Code)
//}
//
//// 获取分页文档不合规的查询条件
//func TestFindPageWithWhereError(t *testing.T) {
//	res := httptest.NewRecorder()
//	where, err := jsoniter.Marshal(map[string]interface{}{
//		"$x": "",
//	})
//	if err != nil {
//		t.Error(err)
//	}
//	req, _ := http.NewRequest("GET", "/services", nil)
//	req.Header.Set("x-page-size", "5")
//	req.Header.Set("x-page", "1")
//	query := req.URL.Query()
//	query.Add("where", string(where))
//	req.URL.RawQuery = query.Encode()
//
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 400, res.Code)
//}
//
//// 获取分页文档
//func TestFindPage(t *testing.T) {
//	res := httptest.NewRecorder()
//	req, _ := http.NewRequest("GET", "/services", nil)
//	req.Header.Set("x-page-size", "5")
//	req.Header.Set("x-page", "1")
//	query := req.URL.Query()
//	query.Add("sort", "price.1")
//	req.URL.RawQuery = query.Encode()
//
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 200, res.Code)
//
//	var data []map[string]interface{}
//	if err := jsoniter.Unmarshal(res.Body.Bytes(), &data); err != nil {
//		t.Error(err)
//	}
//	assert.Equal(t, 5, len(data))
//	assert.Equal(t,
//		[]string{"44243580", "57277117", "87239908", "60599531", "38438365"},
//		funk.Map(data, func(x map[string]interface{}) string {
//			return x["number"].(string)
//		}),
//	)
//	assert.Equal(t, "50", res.Header().Get("X-Page-Total"))
//}
//
//// 获取当个文档不合规的查询
//func TestFindOneWithWhereError(t *testing.T) {
//	res := httptest.NewRecorder()
//	where, err := jsoniter.Marshal(map[string]interface{}{
//		"$x": "",
//	})
//	if err != nil {
//		t.Error(err)
//	}
//	req, _ := http.NewRequest("GET", "/services", nil)
//	query := req.URL.Query()
//	query.Add("where", string(where))
//	query.Add("single", "true")
//	req.URL.RawQuery = query.Encode()
//
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 400, res.Code)
//}
//
//var findOneData map[string]interface{}
//
//// 获取当个文档（过滤）
//func TestFindOne(t *testing.T) {
//	res := httptest.NewRecorder()
//	where, err := jsoniter.Marshal(map[string]interface{}{
//		"number": "55826199",
//	})
//	if err != nil {
//		t.Error(err)
//	}
//	req, _ := http.NewRequest("GET", "/services", nil)
//	query := req.URL.Query()
//	query.Add("where", string(where))
//	query.Add("single", "true")
//	req.URL.RawQuery = query.Encode()
//
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 200, res.Code)
//
//	if err = jsoniter.Unmarshal(res.Body.Bytes(), &findOneData); err != nil {
//		t.Error(err)
//	}
//	assert.Equal(t, "Handmade Soft Salad", findOneData["name"])
//}
//
//// 获取单个文档，非 object_id 返回错误
//func TestFindOneByIdNotObjectId(t *testing.T) {
//	res := httptest.NewRecorder()
//	url := fmt.Sprintf(`/services/%s`, "abc")
//	req, _ := http.NewRequest("GET", url, nil)
//
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 400, res.Code)
//}
//
//// 获取单个文档，不存在的 ID
//func TestFindOneByIdNotExists(t *testing.T) {
//	res := httptest.NewRecorder()
//	url := fmt.Sprintf(`/services/%s`, primitive.NewObjectID().Hex())
//	req, _ := http.NewRequest("GET", url, nil)
//
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 400, res.Code)
//}
//
//// 获取单个文档（ID）
//func TestFindOneById(t *testing.T) {
//	res := httptest.NewRecorder()
//	url := fmt.Sprintf(`/services/%s`, findOneData["_id"].(string))
//	req, _ := http.NewRequest("GET", url, nil)
//
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 200, res.Code)
//
//	if err := jsoniter.Unmarshal(res.Body.Bytes(), &findOneData); err != nil {
//		t.Error(err)
//	}
//	assert.Equal(t, float64(727), findOneData["price"])
//}
//
//// 局部更新文档不合规的 URL
//func TestUpdateManyUrlError(t *testing.T) {
//	res := httptest.NewRecorder()
//	req, _ := http.NewRequest("PATCH", "/Services", nil)
//
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 400, res.Code)
//}
//
//// 局部更新文档空条件
//func TestUpdateManyEmptyWhere(t *testing.T) {
//	res := httptest.NewRecorder()
//	req, _ := http.NewRequest("PATCH", "/services", nil)
//	query := req.URL.Query()
//	req.URL.RawQuery = query.Encode()
//
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 400, res.Code)
//}
//
//// 局部更新文档不合规的查询
//func TestUpdateManyWithWhereError(t *testing.T) {
//	res := httptest.NewRecorder()
//	where, err := jsoniter.Marshal(map[string]interface{}{
//		"$x": "",
//	})
//	if err != nil {
//		t.Error(err)
//	}
//	req, _ := http.NewRequest("PATCH", "/services", nil)
//	query := req.URL.Query()
//	query.Add("where", string(where))
//	req.URL.RawQuery = query.Encode()
//
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 400, res.Code)
//}
//
//// 局部更新格式转换错误
//func TestUpdateManyFormatError(t *testing.T) {
//	res := httptest.NewRecorder()
//	where, err := jsoniter.Marshal(map[string]interface{}{
//		"number": map[string]interface{}{
//			"$in": []string{"66502334", "43678700"},
//		},
//	})
//	if err != nil {
//		t.Error(err)
//	}
//	body, err := jsoniter.Marshal(engine.UpdateBody{
//		Update: map[string]interface{}{
//			"$set": map[string]interface{}{
//				"name": "agent",
//			},
//		},
//		Format: map[string]interface{}{
//			"name": "object_id",
//		},
//	})
//	if err != nil {
//		t.Error(err)
//	}
//	req, _ := http.NewRequest("PATCH", "/services", bytes.NewBuffer(body))
//	query := req.URL.Query()
//	query.Add("where", string(where))
//	req.URL.RawQuery = query.Encode()
//
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 400, res.Code)
//}
//
//// 局部更新不合规的引用
//func TestUpdateManyRefError(t *testing.T) {
//	res := httptest.NewRecorder()
//	where, err := jsoniter.Marshal(map[string]interface{}{
//		"number": map[string]interface{}{
//			"$in": []string{"66502334", "43678700"},
//		},
//	})
//	if err != nil {
//		t.Error(err)
//	}
//	body, err := jsoniter.Marshal(engine.UpdateBody{
//		Update: map[string]interface{}{
//			"$set": map[string]interface{}{
//				"tag": []string{"a1", "a2"},
//			},
//		},
//		Format: map[string]interface{}{
//			"tag": "ref",
//		},
//	})
//	if err != nil {
//		t.Error(err)
//	}
//	req, _ := http.NewRequest("PATCH", "/services", bytes.NewBuffer(body))
//	query := req.URL.Query()
//	query.Add("where", string(where))
//	req.URL.RawQuery = query.Encode()
//
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 400, res.Code)
//}
//
//var updateManyData []map[string]interface{}
//
//// 局部更新多个文档（过滤）
//func TestUpdateMany(t *testing.T) {
//	res := httptest.NewRecorder()
//	where, err := jsoniter.Marshal(map[string]interface{}{
//		"number": map[string]interface{}{
//			"$in": []string{"66502334", "43678700"},
//		},
//	})
//	if err != nil {
//		t.Error(err)
//	}
//	body, err := jsoniter.Marshal(engine.UpdateBody{
//		Update: map[string]interface{}{
//			"$set": map[string]interface{}{
//				"price": 512.00,
//			},
//		},
//	})
//
//	req, _ := http.NewRequest("PATCH", "/services", bytes.NewBuffer(body))
//	query := req.URL.Query()
//	query.Add("where", string(where))
//	req.URL.RawQuery = query.Encode()
//
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 200, res.Code)
//
//	cursor, err := db.Collection("services").Find(context.TODO(), bson.M{
//		"number": bson.M{"$in": bson.A{"66502334", "43678700"}},
//	})
//	if err != nil {
//		t.Error(err)
//	}
//	if err = cursor.All(context.TODO(), &updateManyData); err != nil {
//		t.Error(err)
//	}
//
//	assert.Equal(t,
//		[]float64{512, 512},
//		funk.Map(updateManyData, func(x map[string]interface{}) float64 {
//			return x["price"].(float64)
//		}),
//	)
//}
//
//// 局部更新多个文档（ID）格式转换错误
//func TestUpdateManyByIdFormatError(t *testing.T) {
//	res := httptest.NewRecorder()
//	body, err := jsoniter.Marshal(engine.UpdateBody{
//		Update: map[string]interface{}{
//			"$set": map[string]interface{}{
//				"name": "agent",
//			},
//		},
//		Format: map[string]interface{}{
//			"name": "object_id",
//		},
//	})
//	if err != nil {
//		t.Error(err)
//	}
//	req, _ := http.NewRequest("PATCH", "/services", bytes.NewBuffer(body))
//	query := req.URL.Query()
//	for _, x := range updateManyData {
//		query.Add("id", x["_id"].(primitive.ObjectID).Hex())
//	}
//	req.URL.RawQuery = query.Encode()
//
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 400, res.Code)
//}
//
//// 局部更新多个文档（ID）
//func TestUpdateManyById(t *testing.T) {
//	res := httptest.NewRecorder()
//	body, err := jsoniter.Marshal(engine.UpdateBody{
//		Update: map[string]interface{}{
//			"$set": map[string]interface{}{
//				"price": 1024.00,
//			},
//		},
//	})
//	req, _ := http.NewRequest("PATCH", "/services", bytes.NewBuffer(body))
//	query := req.URL.Query()
//	for _, x := range updateManyData {
//		query.Add("id", x["_id"].(primitive.ObjectID).Hex())
//	}
//	req.URL.RawQuery = query.Encode()
//
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 200, res.Code)
//
//	cursor, err := db.Collection("services").Find(context.TODO(), bson.M{
//		"number": bson.M{"$in": bson.A{"66502334", "43678700"}},
//	})
//	if err != nil {
//		t.Error(err)
//	}
//	var data []map[string]interface{}
//	if err = cursor.All(context.TODO(), &data); err != nil {
//		t.Error(err)
//	}
//
//	assert.Equal(t,
//		[]float64{1024, 1024},
//		funk.Map(data, func(x map[string]interface{}) float64 {
//			return x["price"].(float64)
//		}),
//	)
//}
//
//var updateOneData map[string]interface{}
//
//// 局部更新多个文档格式化错误
//func TestUpdateOneFormatError(t *testing.T) {
//	res := httptest.NewRecorder()
//	where, err := jsoniter.Marshal(map[string]interface{}{
//		"number": "38438365",
//	})
//	if err != nil {
//		t.Error(err)
//	}
//	body, err := jsoniter.Marshal(engine.UpdateBody{
//		Update: map[string]interface{}{
//			"$set": map[string]interface{}{
//				"name": "agent",
//			},
//		},
//		Format: map[string]interface{}{
//			"name": "object_id",
//		},
//	})
//	if err != nil {
//		t.Error(err)
//	}
//	req, _ := http.NewRequest("PATCH", "/services", bytes.NewBuffer(body))
//	query := req.URL.Query()
//	query.Add("where", string(where))
//	query.Add("single", "true")
//	req.URL.RawQuery = query.Encode()
//
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 400, res.Code)
//}
//
//// 局部更新多个文档引用错误
//func TestUpdateOneRefError(t *testing.T) {
//	res := httptest.NewRecorder()
//	where, err := jsoniter.Marshal(map[string]interface{}{
//		"number": "38438365",
//	})
//	if err != nil {
//		t.Error(err)
//	}
//	body, err := jsoniter.Marshal(engine.UpdateBody{
//		Update: map[string]interface{}{
//			"$set": map[string]interface{}{
//				"tag": []string{"a1", "a2"},
//			},
//		},
//		Ref: []string{"tag"},
//	})
//	if err != nil {
//		t.Error(err)
//	}
//	req, _ := http.NewRequest("PATCH", "/services", bytes.NewBuffer(body))
//	query := req.URL.Query()
//	query.Add("where", string(where))
//	query.Add("single", "true")
//	req.URL.RawQuery = query.Encode()
//
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 400, res.Code)
//}
//
//// 局部更新单个文档（过滤）
//func TestUpdateOne(t *testing.T) {
//	res := httptest.NewRecorder()
//	where, err := jsoniter.Marshal(map[string]interface{}{
//		"number": "38438365",
//	})
//	if err != nil {
//		t.Error(err)
//	}
//	body, err := jsoniter.Marshal(engine.UpdateBody{
//		Update: map[string]interface{}{
//			"$set": map[string]interface{}{
//				"price": 512.00,
//			},
//		},
//	})
//	if err != nil {
//		t.Error(err)
//	}
//	req, _ := http.NewRequest("PATCH", "/services", bytes.NewBuffer(body))
//	query := req.URL.Query()
//	query.Add("where", string(where))
//	query.Add("single", "true")
//	req.URL.RawQuery = query.Encode()
//
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 200, res.Code)
//
//	if err = db.Collection("services").
//		FindOne(context.TODO(), bson.M{"number": "38438365"}).
//		Decode(&updateOneData); err != nil {
//		t.Error(err)
//	}
//
//	assert.Equal(t, float64(512), updateOneData["price"])
//}
//
//// 股部更新单个文档（ID）,非 object_id
//func TestUpdateOneByIdNotObjectId(t *testing.T) {
//	res := httptest.NewRecorder()
//	url := fmt.Sprintf(`/services/%s`, "abc")
//	req, _ := http.NewRequest("PATCH", url, nil)
//
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 400, res.Code)
//}
//
//// 局部更新单个文档空条件
//func TestUpdateOneByIdEmptyWhere(t *testing.T) {
//	res := httptest.NewRecorder()
//	url := fmt.Sprintf(`/services/%s`, primitive.NewObjectID().Hex())
//	req, _ := http.NewRequest("PATCH", url, nil)
//	query := req.URL.Query()
//	req.URL.RawQuery = query.Encode()
//
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 400, res.Code)
//}
//
//// 局部更新单个文档（ID）
//func TestUpdateOneByIdFormatError(t *testing.T) {
//	res := httptest.NewRecorder()
//	body, err := jsoniter.Marshal(engine.UpdateBody{
//		Update: map[string]interface{}{
//			"$set": map[string]interface{}{
//				"name": "agent",
//			},
//		},
//		Format: map[string]interface{}{
//			"name": "object_id",
//		},
//	})
//	if err != nil {
//		t.Error(err)
//	}
//	url := fmt.Sprintf(`/services/%s`, updateOneData["_id"].(primitive.ObjectID).Hex())
//	req, _ := http.NewRequest("PATCH", url, bytes.NewBuffer(body))
//
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 400, res.Code)
//}
//
//// 局部更新单个文档（ID）
//func TestUpdateOneById(t *testing.T) {
//	res := httptest.NewRecorder()
//	body, err := jsoniter.Marshal(engine.UpdateBody{
//		Update: map[string]interface{}{
//			"$set": map[string]interface{}{
//				"price": 1024.00,
//			},
//		},
//	})
//	if err != nil {
//		t.Error(err)
//	}
//	url := fmt.Sprintf(`/services/%s`, updateOneData["_id"].(primitive.ObjectID).Hex())
//	req, _ := http.NewRequest("PATCH", url, bytes.NewBuffer(body))
//
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 200, res.Code)
//
//	var data map[string]interface{}
//	if err = db.Collection("services").
//		FindOne(context.TODO(), bson.M{"number": "38438365"}).
//		Decode(&data); err != nil {
//		t.Error(err)
//	}
//
//	assert.Equal(t, float64(1024), data["price"])
//}
//
//// 更新文档，非 object_id 返回错误
//func TestReplaceOneNotObjectId(t *testing.T) {
//	// 不合规的 object_id
//	res := httptest.NewRecorder()
//	url := fmt.Sprintf(`/services/%s`, "abc")
//	req, _ := http.NewRequest("PUT", url, nil)
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 400, res.Code)
//}
//
//// 更新文档不合规的请求内容
//func TestReplaceOneBodyError(t *testing.T) {
//	res := httptest.NewRecorder()
//	url := fmt.Sprintf(`/services/%s`, primitive.NewObjectID().Hex())
//	req, _ := http.NewRequest("PUT", url, nil)
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 400, res.Code)
//}
//
//// 更新文档不合规的格式化处理
//func TestReplaceOneFormatError(t *testing.T) {
//	body, err := jsoniter.Marshal(engine.ReplaceOneBody{
//		Doc: map[string]interface{}{
//			"name": "abc",
//		},
//		Format: map[string]interface{}{
//			"name": "object_id",
//		},
//	})
//	if err != nil {
//		t.Error(err)
//	}
//	res := httptest.NewRecorder()
//	url := fmt.Sprintf(`/services/%s`, primitive.NewObjectID().Hex())
//	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(body))
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 400, res.Code)
//}
//
//// 更新文档，不合规的引用
//func TestReplaceOneRefError(t *testing.T) {
//	body, err := jsoniter.Marshal(engine.ReplaceOneBody{
//		Doc: map[string]interface{}{
//			"tag": []string{"a1", "a2"},
//		},
//		Ref: []string{"tag"},
//	})
//	if err != nil {
//		t.Error(err)
//	}
//	res := httptest.NewRecorder()
//	url := fmt.Sprintf(`/services/%s`, primitive.NewObjectID().Hex())
//	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(body))
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 400, res.Code)
//}
//
//// 更新文档
//func TestReplaceOne(t *testing.T) {
//	var doc map[string]interface{}
//	if err := db.Collection("services").FindOne(context.TODO(), bson.M{
//		"number": "62787493",
//	}).Decode(&doc); err != nil {
//		return
//	}
//
//	id := doc["_id"].(primitive.ObjectID).Hex()
//	delete(doc, "_id")
//	delete(doc, "create_time")
//	delete(doc, "update_time")
//	doc["price"] = 777.00
//
//	body, err := jsoniter.Marshal(engine.ReplaceOneBody{
//		Doc: doc,
//	})
//
//	res := httptest.NewRecorder()
//	url := fmt.Sprintf(`/services/%s`, id)
//	if err != nil {
//		t.Error(err)
//	}
//	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(body))
//
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 200, res.Code)
//
//	var update map[string]interface{}
//	if err = db.Collection("services").FindOne(context.TODO(), bson.M{
//		"number": "62787493",
//	}).Decode(&update); err != nil {
//		t.Error(err)
//	}
//	assert.Equal(t, float64(777), doc["price"])
//}
//
//// 删除文档，非 object_id 返回错误
//func TestDeleteOneNotObjectId(t *testing.T) {
//	res := httptest.NewRecorder()
//	url := fmt.Sprintf(`/services/%s`, "abc")
//	req, _ := http.NewRequest("DELETE", url, nil)
//
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 400, res.Code)
//}
//
//// 删除文档
//func TestDeleteOne(t *testing.T) {
//	var doc map[string]interface{}
//	if err := db.Collection("services").FindOne(context.TODO(), bson.M{
//		"number": "35433318",
//	}).Decode(&doc); err != nil {
//		t.Error(err)
//	}
//	res := httptest.NewRecorder()
//	url := fmt.Sprintf(`/services/%s`, doc["_id"].(primitive.ObjectID).Hex())
//	req, _ := http.NewRequest("DELETE", url, nil)
//
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 200, res.Code)
//
//	count, err := db.Collection("services").CountDocuments(context.TODO(), bson.M{
//		"number": "35433318",
//	})
//	if err != nil {
//		t.Error(err)
//	}
//	assert.Equal(t, int64(0), count)
//}
//
//// 模型预定义测试
//func TestPredefinedModel(t *testing.T) {
//	var data map[string]interface{}
//	if err := db.Collection("services").FindOne(context.TODO(), bson.M{
//		"number": "55826199",
//	}).Decode(&data); err != nil {
//		t.Error(err)
//	}
//	res := httptest.NewRecorder()
//	id := data["_id"].(primitive.ObjectID).Hex()
//	req, _ := http.NewRequest("GET", fmt.Sprintf(`/svc/%s`, id), nil)
//	r.ServeHTTP(res, req)
//	var result map[string]interface{}
//	if err := jsoniter.Unmarshal(res.Body.Bytes(), &result); err != nil {
//		t.Error(err)
//	}
//	assert.Equal(t, data["name"], result["name"])
//}
//
//// 获取多个文档，固定字段测试
//func TestFindForStaticProjection(t *testing.T) {
//	res := httptest.NewRecorder()
//	req, _ := http.NewRequest("GET", "/users", nil)
//
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 200, res.Code)
//
//	var data []map[string]interface{}
//	if err := jsoniter.Unmarshal(res.Body.Bytes(), &data); err != nil {
//		t.Error(err)
//	}
//
//	assert.Nil(t, data[0]["password"])
//}
//
//// 获取单个文档，固定字段测试
//func TestFindOneForStaticProjection(t *testing.T) {
//	res := httptest.NewRecorder()
//	where, err := jsoniter.Marshal(map[string]interface{}{
//		"name": "admin",
//	})
//	if err != nil {
//		t.Error(err)
//	}
//	req, _ := http.NewRequest("GET", "/users", nil)
//	query := req.URL.Query()
//	query.Add("where", string(where))
//	query.Add("single", "true")
//	req.URL.RawQuery = query.Encode()
//
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 200, res.Code)
//
//	var data map[string]interface{}
//	if err := jsoniter.Unmarshal(res.Body.Bytes(), &data); err != nil {
//		t.Error(err)
//	}
//
//	assert.Nil(t, data["password"])
//}
//
//// 创建文档，队列事件测试
//func TestCreateForStaticEvent(t *testing.T) {
//	var wg sync.WaitGroup
//	wg.Add(1)
//	subj := "test.events.pages"
//	queue := "test:events:pages"
//	sub, err := js.QueueSubscribe(subj, queue, func(msg *nats.Msg) {
//		assert.NotEmpty(t, msg.Data)
//		wg.Done()
//	})
//	if err != nil {
//		t.Error(err)
//	}
//	defer sub.Unsubscribe()
//	res := httptest.NewRecorder()
//	body, err := jsoniter.Marshal(CreateBody{
//		Doc: map[string]interface{}{
//			"name": "首页",
//		},
//	})
//	if err != nil {
//		panic(err)
//	}
//	req, _ := http.NewRequest("POST", "/pages", bytes.NewBuffer(body))
//
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 201, res.Code)
//	wg.Wait()
//}
//
//func TestUpdateForStaticEvent(t *testing.T) {
//	var wg sync.WaitGroup
//	wg.Add(1)
//	subj := "test.events.pages"
//	queue := "test:events:pages"
//	sub, err := js.QueueSubscribe(subj, queue, func(msg *nats.Msg) {
//		assert.NotEmpty(t, msg.Data)
//		wg.Done()
//	})
//	if err != nil {
//		t.Error(err)
//	}
//	defer sub.Unsubscribe()
//	res := httptest.NewRecorder()
//	where, err := jsoniter.Marshal(map[string]interface{}{
//		"name": "首页",
//	})
//	if err != nil {
//		t.Error(err)
//	}
//	body, err := jsoniter.Marshal(engine.UpdateBody{
//		Update: map[string]interface{}{
//			"$set": map[string]interface{}{
//				"sort": 1,
//			},
//		},
//	})
//	if err != nil {
//		t.Error(err)
//	}
//	req, _ := http.NewRequest("PATCH", "/pages", bytes.NewBuffer(body))
//	query := req.URL.Query()
//	query.Add("where", string(where))
//	query.Add("single", "true")
//	req.URL.RawQuery = query.Encode()
//
//	r.ServeHTTP(res, req)
//	assert.Equal(t, 200, res.Code)
//
//	wg.Wait()
//}
