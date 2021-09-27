package support

import (
	"database/sql/driver"
	jsoniter "github.com/json-iterator/go"
	"gorm.io/gorm"
)

type Planx struct {
	ID       uint64      `json:"id"`
	Parent   uint64      `gorm:"index:idx_parent_fragment,unique;default:0" json:"parent"`
	Fragment string      `gorm:"type:varchar;not null;index:idx_parent_fragment,unique" json:"fragment"`
	Name     string      `gorm:"type:varchar;not null" json:"name"`
	Nav      *bool       `gorm:"default:false" json:"nav"`
	Schema   string      `gorm:"type:varchar" json:"schema"`
	Template string      `gorm:"type:varchar" json:"template"`
	Option   PlanxOption `gorm:"type:jsonb;default:'{}'" json:"option"`
	Icon     string      `gorm:"type:varchar" json:"icon"`
	Sort     uint8       `gorm:"default:0" json:"sort"`
}

type PlanxOption struct {
	Fetch   bool         `json:"fetch,omitempty"`
	Columns []ViewColumn `json:"columns,omitempty"`
}

func (x *PlanxOption) Scan(input interface{}) error {
	return jsoniter.Unmarshal(input.([]byte), x)
}

func (x PlanxOption) Value() (driver.Value, error) {
	return jsoniter.Marshal(x)
}

type ViewColumn struct {
	Key string `json:"key"`
}

func GeneratePlanx(tx *gorm.DB) (err error) {
	if tx.Migrator().HasTable(&Planx{}) {
		if err = tx.Migrator().DropTable(&Planx{}); err != nil {
			return
		}
	}
	if err = tx.AutoMigrate(&Planx{}); err != nil {
		return
	}
	if err = tx.Exec("create index option_gin on planx using gin(option)").Error; err != nil {
		return
	}
	return tx.Transaction(func(txx *gorm.DB) (err error) {
		dashboard := Planx{
			Fragment: "dashboard",
			Name:     "仪表盘",
			Nav:      True(),
			Template: "manual",
			Icon:     "dashboard",
		}
		if err = txx.Create(&dashboard).Error; err != nil {
			return
		}
		center := Planx{
			Fragment: "center",
			Name:     "个人中心",
		}
		if err = txx.Create(&center).Error; err != nil {
			return
		}
		centerChildren := []Planx{
			{
				Parent:   center.ID,
				Fragment: "profile",
				Name:     "我的信息",
				Template: "manual",
			},
			{
				Parent:   center.ID,
				Fragment: "notification",
				Name:     "消息通知",
				Template: "manual",
			},
		}
		if err = txx.Create(&centerChildren).Error; err != nil {
			return
		}
		settings := Planx{
			Fragment: "settings",
			Name:     "设置",
			Nav:      True(),
			Icon:     "setting",
		}
		if err = txx.Create(&settings).Error; err != nil {
			return
		}
		settingsChildren := []Planx{
			{
				Parent:   settings.ID,
				Fragment: "schema",
				Name:     "模型管理",
				Nav:      True(),
				Template: "manual",
			},
			{
				Parent:   settings.ID,
				Fragment: "planx",
				Name:     "布局管理",
				Nav:      True(),
				Template: "manual",
			},
			{
				Parent:   settings.ID,
				Fragment: "role",
				Name:     "权限管理",
				Nav:      True(),
				Schema:   "role",
				Template: "list",
			},
			{
				Parent:   settings.ID,
				Fragment: "admin",
				Name:     "成员管理",
				Nav:      True(),
				Schema:   "admin",
				Template: "list",
			},
		}
		if err = txx.Create(&settingsChildren).Error; err != nil {
			return
		}
		var role Planx
		if err = txx.
			Where("parent = ?", settings.ID).
			Where("fragment = ?", "role").
			First(&role).Error; err != nil {
			return
		}
		roleChildren := []Planx{
			{
				Parent:   role.ID,
				Fragment: "create",
				Name:     "创建资源",
				Schema:   "role",
				Template: "page",
			},
			{
				Parent:   role.ID,
				Fragment: "update",
				Name:     "更新资源",
				Schema:   "role",
				Template: "page",
			},
		}
		if err = txx.Create(&roleChildren).Error; err != nil {
			return
		}
		var admin Planx
		if err = txx.
			Where("parent = ?", settings.ID).
			Where("fragment = ?", "admin").
			First(&admin).Error; err != nil {
			return
		}
		adminChildren := []Planx{
			{
				Parent:   admin.ID,
				Fragment: "create",
				Name:     "创建资源",
				Schema:   "admin",
				Template: "page",
			},
			{
				Parent:   admin.ID,
				Fragment: "update",
				Name:     "更新资源",
				Schema:   "admin",
				Template: "page",
			},
		}
		if err = txx.Create(&adminChildren).Error; err != nil {
			return
		}

		return
	})
}
