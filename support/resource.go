package support

import (
	"database/sql/driver"
	jsoniter "github.com/json-iterator/go"
	"gorm.io/gorm"
)

type Resource struct {
	ID       int64  `json:"id"`
	Parent   int64  `gorm:"default:0" json:"parent"`
	Name     string `gorm:"type:varchar;not null" json:"name"`
	Fragment string `gorm:"type:varchar;not null" json:"fragment"`
	Router   *bool  `gorm:"default:false" json:"router"`
	Nav      *bool  `gorm:"default:false" json:"nav"`
	Icon     string `gorm:"type:varchar" json:"icon"`
	Schema   Schema `gorm:"type:jsonb;default:'{}'" json:"schema,omitempty"`
	Sort     int8   `gorm:"default:0" json:"sort"`
}

type Schema struct {
	Key     string   `json:"key"`
	Type    string   `json:"type"`
	Columns []Column `json:"columns"`
	System  bool     `json:"system,omitempty"`
}

func (x *Schema) Scan(input interface{}) error {
	return jsoniter.Unmarshal(input.([]byte), x)
}

func (x Schema) Value() (driver.Value, error) {
	return jsoniter.Marshal(x)
}

type Column struct {
	Key      string   `json:"key"`
	Label    string   `json:"label"`
	Type     string   `json:"type"`
	Default  string   `json:"default,omitempty"`
	Unique   bool     `json:"unique,omitempty"`
	Require  bool     `json:"require,omitempty"`
	Hide     bool     `json:"hide,omitempty"`
	Relation Relation `json:"relation,omitempty"`
	System   bool     `json:"system,omitempty"`
}

type Relation struct {
	Mode       string `json:"mode,omitempty"`
	Target     string `json:"target,omitempty"`
	References string `json:"references,omitempty"`
}

func GenerateResources(tx *gorm.DB) (err error) {
	if tx.Migrator().HasTable(&Resource{}) {
		if err = tx.Migrator().DropTable(&Resource{}); err != nil {
			return
		}
	}
	if err = tx.AutoMigrate(&Resource{}); err != nil {
		return
	}
	tx.Exec("create index schema_gin on resource using gin(schema)")
	return tx.Transaction(func(tx *gorm.DB) (err error) {
		dashboard := Resource{
			Fragment: "dashboard",
			Name:     "仪表盘",
			Nav:      True(),
			Router:   True(),
			Icon:     "dashboard",
		}
		if err = tx.Create(&dashboard).Error; err != nil {
			return
		}
		center := Resource{
			Fragment: "center",
			Name:     "个人中心",
		}
		if err = tx.Create(&center).Error; err != nil {
			return
		}
		centerChildren := []Resource{
			{
				Parent:   center.ID,
				Fragment: "profile",
				Name:     "我的信息",
				Router:   True(),
			},
			{
				Parent:   center.ID,
				Fragment: "notification",
				Name:     "消息通知",
				Router:   True(),
			},
		}
		if err = tx.Create(&centerChildren).Error; err != nil {
			return
		}
		settings := Resource{
			Fragment: "settings",
			Name:     "设置",
			Nav:      True(),
			Router:   False(),
			Icon:     "setting",
		}
		if err = tx.Create(&settings).Error; err != nil {
			return
		}
		resource := Resource{
			Parent:   settings.ID,
			Fragment: "resource",
			Name:     "资源管理",
			Nav:      True(),
			Router:   True(),
			Schema: Schema{
				Key:     "resource",
				Type:    "customize",
				Columns: []Column{},
				System:  true,
			},
		}
		if err = tx.Create(&resource).Error; err != nil {
			return
		}
		role := Resource{
			Parent:   settings.ID,
			Fragment: "role",
			Name:     "权限管理",
			Nav:      True(),
			Router:   True(),
			Schema: Schema{
				Key:  "role",
				Type: "collection",
				Columns: []Column{
					{
						Key:     "key",
						Label:   "权限代码",
						Type:    "varchar",
						Require: true,
						Unique:  true,
						System:  true,
					},
					{
						Key:     "name",
						Label:   "权限名称",
						Type:    "varchar",
						Require: true,
						System:  true,
					},
					{
						Key:    "description",
						Label:  "描述",
						Type:   "text",
						System: true,
					},
					{
						Key:     "routers",
						Label:   "路由",
						Type:    "rel",
						Default: "'[]'",
						Relation: Relation{
							Mode:   "customize",
							Target: "resource",
						},
						System: true,
					},
					{
						Key:     "permissions",
						Label:   "策略",
						Type:    "rel",
						Default: "'[]'",
						Relation: Relation{
							Mode:   "customize",
							Target: "resource",
						},
						System: true,
					},
				},
				System: true,
			},
		}
		if err = tx.Create(&role).Error; err != nil {
			return
		}
		roleChildren := []Resource{
			{
				Parent:   role.ID,
				Fragment: "create",
				Name:     "创建资源",
				Router:   True(),
			},
			{
				Parent:   role.ID,
				Fragment: "update",
				Name:     "更新资源",
				Router:   True(),
			},
		}
		if err = tx.Create(&roleChildren).Error; err != nil {
			return
		}
		admin := Resource{
			Parent:   settings.ID,
			Fragment: "admin",
			Name:     "成员管理",
			Nav:      True(),
			Router:   True(),
			Schema: Schema{
				Key:  "admin",
				Type: "collection",
				Columns: []Column{
					{
						Key:     "uuid",
						Label:   "唯一标识",
						Type:    "uuid",
						Default: "uuid_generate_v4()",
						Require: true,
						Unique:  true,
						Hide:    true,
						System:  true,
					},
					{
						Key:     "username",
						Label:   "用户名",
						Type:    "varchar",
						Require: true,
						Unique:  true,
						System:  true,
					},
					{
						Key:     "password",
						Label:   "密码",
						Type:    "varchar",
						Require: true,
						System:  true,
					},
					{
						Key:     "roles",
						Label:   "权限",
						Type:    "rel",
						Require: true,
						Default: "'[]'",
						Relation: Relation{
							Mode:       "many",
							Target:     "role",
							References: "key",
						},
						System: true,
					},
					{
						Key:    "name",
						Label:  "姓名",
						Type:   "varchar",
						System: true,
					},
					{
						Key:    "email",
						Label:  "邮件",
						Type:   "varchar",
						System: true,
					},
					{
						Key:    "phone",
						Label:  "联系方式",
						Type:   "varchar",
						System: true,
					},
					{
						Key:     "avatar",
						Label:   "头像",
						Type:    "array",
						Default: "'[]'",
						System:  true,
					},
					{
						Key:     "routers",
						Label:   "路由",
						Type:    "rel",
						Default: "'[]'",
						Relation: Relation{
							Mode:   "customize",
							Target: "resource",
						},
						System: true,
					},
					{
						Key:     "permissions",
						Label:   "策略",
						Type:    "rel",
						Default: "'[]'",
						Relation: Relation{
							Mode:   "customize",
							Target: "resource",
						},
						System: true,
					},
				},
				System: true,
			},
		}
		if err = tx.Create(&admin).Error; err != nil {
			return
		}
		adminChildren := []Resource{
			{
				Parent:   admin.ID,
				Fragment: "create",
				Name:     "创建资源",
				Router:   True(),
			},
			{
				Parent:   admin.ID,
				Fragment: "update",
				Name:     "更新资源",
				Router:   True(),
			},
		}
		if err = tx.Create(&adminChildren).Error; err != nil {
			return
		}
		return
	})
}
