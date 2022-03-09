package sequence

type Sequence struct {
	Key   string `json:"key" gorm:"primaryKey"`
	Value int    `json:"value"`
}

func (*Sequence) TableName() string {
	return "t_sequence"
}
