package basic

import (
	"database/sql/driver"
	jsoniter "github.com/json-iterator/go"
	"gorm.io/gorm"
)

type Page struct {
	ID       uint64       `json:"id"`
	Parent   uint64       `gorm:"index:idx_parent_fragment,unique;default:0" json:"parent"`
	Fragment string       `gorm:"type:varchar;not null;index:idx_parent_fragment,unique" json:"fragment"`
	Name     string       `gorm:"type:varchar;not null" json:"name"`
	Router   RouterOption `gorm:"type:jsonb;default:'{}'" json:"router"`
	Nav      *bool        `gorm:"default:false" json:"nav"`
	Icon     string       `gorm:"type:varchar" json:"icon"`
	Sort     uint8        `gorm:"default:0" json:"sort"`
}

type RouterOption struct {
	Schema   string       `json:"schema,omitempty"`
	Template string       `json:"template,omitempty"`
	Fetch    bool         `json:"fetch,omitempty"`
	Columns  []ViewColumn `json:"columns,omitempty"`
}

func (x *RouterOption) Scan(input interface{}) error {
	return jsoniter.Unmarshal(input.([]byte), x)
}

func (x RouterOption) Value() (driver.Value, error) {
	return jsoniter.Marshal(x)
}

type ViewColumn struct {
	Key string `json:"key"`
}

func GeneratePage(tx *gorm.DB) (err error) {
	if tx.Migrator().HasTable(&Page{}) {
		if err = tx.Migrator().DropTable(&Page{}); err != nil {
			return
		}
	}
	if err = tx.AutoMigrate(&Page{}); err != nil {
		return
	}
	if err = tx.Exec("create index router_gin on page using gin(router)").Error; err != nil {
		return
	}
	return tx.Transaction(func(txx *gorm.DB) (err error) {
		dashboard := Page{
			Fragment: "dashboard",
			Name:     "仪表盘",
			Nav:      True(),
			Router: RouterOption{
				Template: "manual",
			},
			Icon: "dashboard",
		}
		if err = txx.Create(&dashboard).Error; err != nil {
			return
		}
		center := Page{
			Fragment: "center",
			Name:     "个人中心",
		}
		if err = txx.Create(&center).Error; err != nil {
			return
		}
		centerChildren := []Page{
			{
				Parent:   center.ID,
				Fragment: "profile",
				Name:     "我的信息",
				Router: RouterOption{
					Template: "manual",
				},
			},
			{
				Parent:   center.ID,
				Fragment: "notification",
				Name:     "消息通知",
				Router: RouterOption{
					Template: "manual",
				},
			},
		}
		if err = txx.Create(&centerChildren).Error; err != nil {
			return
		}
		settings := Page{
			Fragment: "settings",
			Name:     "设置",
			Nav:      True(),
			Icon:     "setting",
		}
		if err = txx.Create(&settings).Error; err != nil {
			return
		}
		settingsChildren := []Page{
			{
				Parent:   settings.ID,
				Fragment: "schema",
				Name:     "模型管理",
				Nav:      True(),
				Router: RouterOption{
					Template: "manual",
				},
			},
			{
				Parent:   settings.ID,
				Fragment: "page",
				Name:     "页面管理",
				Nav:      True(),
				Router: RouterOption{
					Template: "manual",
				},
			},
			{
				Parent:   settings.ID,
				Fragment: "role",
				Name:     "权限管理",
				Nav:      True(),
				Router: RouterOption{
					Schema:   "role",
					Template: "list",
				},
			},
			{
				Parent:   settings.ID,
				Fragment: "admin",
				Name:     "成员管理",
				Nav:      True(),
				Router: RouterOption{
					Schema:   "admin",
					Template: "list",
				},
			},
		}
		if err = txx.Create(&settingsChildren).Error; err != nil {
			return
		}
		var role Page
		if err = txx.
			Where("parent = ?", settings.ID).
			Where("fragment = ?", "role").
			First(&role).Error; err != nil {
			return
		}
		roleChildren := []Page{
			{
				Parent:   role.ID,
				Fragment: "create",
				Name:     "创建资源",
				Router: RouterOption{
					Schema:   "role",
					Template: "form",
				},
			},
			{
				Parent:   role.ID,
				Fragment: "update",
				Name:     "更新资源",
				Router: RouterOption{
					Schema:   "role",
					Template: "form",
					Fetch:    true,
				},
			},
		}
		if err = txx.Create(&roleChildren).Error; err != nil {
			return
		}
		var admin Page
		if err = txx.
			Where("parent = ?", settings.ID).
			Where("fragment = ?", "admin").
			First(&admin).Error; err != nil {
			return
		}
		adminChildren := []Page{
			{
				Parent:   admin.ID,
				Fragment: "create",
				Name:     "创建资源",
				Router: RouterOption{
					Schema:   "admin",
					Template: "form",
				},
			},
			{
				Parent:   admin.ID,
				Fragment: "update",
				Name:     "更新资源",
				Router: RouterOption{
					Schema:   "admin",
					Template: "form",
					Fetch:    true,
				},
			},
		}
		if err = txx.Create(&adminChildren).Error; err != nil {
			return
		}
		return
	})
}
