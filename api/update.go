package api

//// UpdateBody Update resource request body
//type UpdateBody struct {
//	Data map[string]interface{} `json:"data" binding:"required"`
//
//	Conditions `json:"where" binding:"gte=0,dive,len=3,dive,required"`
//}
//
//// Update resources
//func (x *API) Update(c *gin.Context) interface{} {
//	uri, err := x.getUri(c)
//	if err != nil {
//		return err
//	}
//	var body UpdateBody
//	if err := c.ShouldBindJSON(&body); err != nil {
//		return err
//	}
//	tx := x.Db.WithContext(c).Table(uri.Model)
//	tx = x.where(tx, body.Conditions)
//	if err := tx.Updates(body.Data).Error; err != nil {
//		return err
//	}
//	return "ok"
//}
