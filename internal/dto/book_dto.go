package dto

type CreateBookDTO struct {
	Title       string `json:"title"       binding:"required,min=1,max=255" example:"Мастер и Маргарита"`
	Author      string `json:"author"      binding:"required,min=1,max=255" example:"Михаил Булгаков"`
	Description string `json:"description"                                  example:"Роман о добре и зле"`
	Year        int    `json:"year"        binding:"min=0,max=2100"          example:"1967"`
}

type UpdateBookDTO struct {
	Title       string `json:"title"       binding:"required,min=1,max=255" example:"Белая гвардия"`
	Author      string `json:"author"      binding:"required,min=1,max=255" example:"Михаил Булгаков"`
	Description string `json:"description"                                  example:"Роман о гражданской войне"`
	Year        int    `json:"year"        binding:"min=0,max=2100"          example:"1925"`
}

type PatchBookDTO struct {
	Title       *string `json:"title"       binding:"omitempty,min=1,max=255" example:"Новое название"`
	Author      *string `json:"author"      binding:"omitempty,min=1,max=255" example:"Новый автор"`
	Description *string `json:"description"                                   example:"Новое описание"`
	Year        *int    `json:"year"        binding:"omitempty,min=0,max=2100" example:"2000"`
}

type PaginationDTO struct {
	Page  int `form:"page"  binding:"min=0" example:"1"`
	Limit int `form:"limit" binding:"min=0,max=100" example:"10"`
}

func (p *PaginationDTO) SetDefaults() {
	if p.Page == 0 {
		p.Page = 1
	}
	if p.Limit == 0 {
		p.Limit = 10
	}
}
