package entity

type Service struct {
	ServiceID   uint64  `gorm:"primaryKey;autoIncrement" json:"service_id"`
	ServiceName string  `gorm:"type:varchar(256);unique;not null" json:"service_name"`
	CurrentCost float64 `gorm:"type:decimal(10,2);not null;default:0.00" json:"current_cost"`
	IsAvailable bool    `gorm:"not null;default:true" json:"is_available"`
}

func (Service) TableName() string {
	return "services"
}
