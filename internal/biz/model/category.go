package model

type Category struct {
	Id        int64  `json:"id,omitempty" bson:"id" form:"id"`
	Mode      uint8  `json:"mode,omitempty" bson:"mode" form:"mode"`
	ClassName string `json:"class_name,omitempty" bson:"class_name" form:"class_name"`
	Name      string `json:"name,omitempty" bson:"name" form:"name"`
	Image     string `json:"image,omitempty" bson:"image" form:"image"`
	SortIndex uint16 `json:"sort_index,omitempty" bson:"sort_index" form:"sort_index"`
	IconId    int32  `json:"icon_id,omitempty" bson:"icon_id" form:"icon_id"`
}

func (o *Category) TableName() string {
	return "categories"
}
