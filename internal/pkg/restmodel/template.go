package restmodel

import "github.com/henomis/mailqueue-go/internal/pkg/storagemodel"

type Templates []Template

func (t *Templates) FromStorageModel(storageItems []storagemodel.Template) {
	for _, storageItem := range storageItems {
		var template Template
		template.FromStorageModel(&storageItem)
		*t = append(*t, template)
	}
}

type TemplatesCount struct {
	Templates Templates `json:"templates"`
	Count     int64     `json:"count"`
}

func (t *TemplatesCount) FromStorageModel(storageItems []storagemodel.Template, count int64) {

	t.Templates.FromStorageModel(storageItems)
	t.Count = count
}

type Template struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Template string `json:"template"`
}

func (t *Template) ToStorageModel() *storagemodel.Template {
	return &storagemodel.Template{
		ID:       t.ID,
		Name:     t.Name,
		Template: t.Template,
	}
}

func (t *Template) FromStorageModel(s *storagemodel.Template) {
	t.ID = s.ID
	t.Name = s.Name
	t.Template = s.Template
}

type TemplateID struct {
	ID string `json:"id"`
}
