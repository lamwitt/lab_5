package dto

type CreateBookDTO struct {
	Title       string `json:"title"       binding:"required,min=1,max=255"`
	Author      string `json:"author"      binding:"required,min=1,max=255"`
	Description string `json:"description"`
	Year        int    `json:"year"        binding:"min=0,max=2100"`
}

type UpdateBookDTO struct {
	Title       string `json:"title"       binding:"required,min=1,max=255"`
	Author      string `json:"author"      binding:"required,min=1,max=255"`
	Description string `json:"description"`
	Year        int    `json:"year"        binding:"min=0,max=2100"`
}

type PatchBookDTO struct {
	Title       *string `json:"title"       binding:"omitempty,min=1,max=255"`
	Author      *string `json:"author"      binding:"omitempty,min=1,max=255"`
	Description *string `json:"description"`
	Year        *int    `json:"year"        binding:"omitempty,min=0,max=2100"`
}

type PaginationDTO struct {
	Page  int `form:"page"  binding:"min=0"`
	Limit int `form:"limit" binding:"min=0,max=100"`
}

func (p *PaginationDTO) SetDefaults() {
	if p.Page == 0 {
		p.Page = 1
	}
	if p.Limit == 0 {
		p.Limit = 10
	}
}
