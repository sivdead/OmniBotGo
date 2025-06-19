package entity

// SystemConfig 系统配置实体，存储系统级配置参数
type SystemConfig struct {
	BaseEntity
	ConfigKey   string `json:"config_key" gorm:"column:config_key;type:varchar(100);uniqueIndex;not null;comment:配置键"`
	ConfigValue string `json:"config_value" gorm:"column:config_value;type:text;comment:配置值"`
	ConfigType  string `json:"config_type" gorm:"column:config_type;type:varchar(50);comment:配置类型"`
	ConfigGroup string `json:"config_group" gorm:"column:config_group;type:varchar(50);comment:配置分组"`
	Description string `json:"description" gorm:"column:description;type:text;comment:配置描述"`
	IsEncrypted bool   `json:"is_encrypted" gorm:"column:is_encrypted;default:false;comment:是否加密存储"`
	IsSystem    bool   `json:"is_system" gorm:"column:is_system;default:false;comment:是否为系统配置"`
}

// TableName 指定表名
func (SystemConfig) TableName() string {
	return "system_configs"
}

// IsString 检查配置类型是否为字符串
func (sc *SystemConfig) IsString() bool {
	return sc.ConfigType == "string" || sc.ConfigType == ""
}

// IsInt 检查配置类型是否为整数
func (sc *SystemConfig) IsInt() bool {
	return sc.ConfigType == "int" || sc.ConfigType == "integer"
}

// IsBool 检查配置类型是否为布尔值
func (sc *SystemConfig) IsBool() bool {
	return sc.ConfigType == "bool" || sc.ConfigType == "boolean"
}

// IsFloat 检查配置类型是否为浮点数
func (sc *SystemConfig) IsFloat() bool {
	return sc.ConfigType == "float" || sc.ConfigType == "double"
}

// IsJSON 检查配置类型是否为JSON
func (sc *SystemConfig) IsJSON() bool {
	return sc.ConfigType == "json"
}

// NeedsEncryption 检查是否需要加密存储
func (sc *SystemConfig) NeedsEncryption() bool {
	return sc.IsEncrypted
}

// IsUserEditable 检查用户是否可以编辑此配置
func (sc *SystemConfig) IsUserEditable() bool {
	return !sc.IsSystem
}

// GetGroupPrefix 获取配置组前缀
func (sc *SystemConfig) GetGroupPrefix() string {
	if sc.ConfigGroup == "" {
		return "default"
	}
	return sc.ConfigGroup
}

// GetFullKey 获取包含组前缀的完整配置键
func (sc *SystemConfig) GetFullKey() string {
	if sc.ConfigGroup == "" {
		return sc.ConfigKey
	}
	return sc.ConfigGroup + "." + sc.ConfigKey
}

// Validate 验证SystemConfig实体数据
func (sc *SystemConfig) Validate() error {
	if sc.ConfigKey == "" {
		return NewValidationError("config_key", "配置键不能为空")
	}
	if sc.ConfigType != "" {
		validTypes := []string{"string", "int", "integer", "bool", "boolean", "float", "double", "json"}
		isValid := false
		for _, validType := range validTypes {
			if sc.ConfigType == validType {
				isValid = true
				break
			}
		}
		if !isValid {
			return NewValidationError("config_type", "配置类型无效")
		}
	}
	return nil
}
