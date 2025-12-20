package entity

type Service struct {
	ServiceID   uint64 `gorm:"primaryKey;autoIncrement" json:"service_id"`
	ServiceName string `gorm:"type:varchar(225);unique;not null" json:"service_name"`
	CurrentCost int64  `gorm:"type:bigint;not null;default:0" json:"current_cost"`
	IsAvailable bool   `gorm:"not null;default:true" json:"is_available"`
}

func (Service) TableName() string {
	return "services"
}
