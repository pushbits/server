package model

// Application holds information like the name, the token, and the associated user of an application.
type Application struct {
	ID       uint   `gorm:"AUTO_INCREMENT;primary_key" json:"id"`
	Token    string `gorm:"type:string;size:64;unique" json:"token"`
	UserID   uint   `json:"-"`
	Name     string `gorm:"type:string" json:"name"`
	MatrixID string `gorm:"type:string" json:"-"`
}

// CreateApplication is used to process queries for creating applications.
type CreateApplication struct {
	Name                string `form:"name" query:"name" json:"name" binding:"required"`
	StrictCompatibility bool   `form:"strict_compatibility" query:"strict_compatibility" json:"strict_compatibility"`
}

// UpdateApplication is used to process queries for updating applications.
type UpdateApplication struct {
	Name                *string `form:"new_name" query:"new_name" json:"new_name"`
	RefreshToken        *bool   `form:"refresh_token" query:"refresh_token" json:"refresh_token"`
	StrictCompatibility *bool   `form:"strict_compatibility" query:"strict_compatibility" json:"strict_compatibility"`
}
