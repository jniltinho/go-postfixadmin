package database

type Domain struct {
	Domain    string `gorm:"column:domain;primaryKey"`
	Transport string `gorm:"column:transport"`
}

func (Domain) TableName() string {
	return "domain"
}

type TransportList struct {
	Domain    string `gorm:"column:domain;primaryKey"`
	Transport string `gorm:"column:transport"`
}

func (TransportList) TableName() string {
	return "transport_list"
}

type Mailbox struct {
	Username  string `gorm:"column:username;primaryKey"`
	Transport string `gorm:"column:transport"`
}

func (Mailbox) TableName() string {
	return "mailbox"
}
