package pagination

import (
	validation "github.com/go-ozzo/ozzo-validation"
)

type PaginationMeta struct {
	Page    int `json:"page"`
	PerPage int `json:"limit"`
	Total   int `json:"total"`
}

type PaginationOptions struct {
	Page    int
	PerPage int
	Offset  uint
}

func CreatePaginationOptions(page int, perPage int) (*PaginationOptions, error) {

	paginationOptions := &PaginationOptions{
		PerPage: perPage,
		Page:    page,
		Offset:  0,
	}

	paginationOptions.Offset = uint(paginationOptions.PerPage * (paginationOptions.Page - 1))

	err := validation.ValidateStruct(paginationOptions,
		validation.Field(&paginationOptions.PerPage, validation.Min(-1), validation.Max(100)),
		validation.Field(&paginationOptions.Page, validation.Min(1)),
	)

	if err != nil {
		return nil, err
	}

	return paginationOptions, nil
}
