package restmodel

import "github.com/henomis/mailqueue-go/internal/pkg/storagemodel"

type Templates struct {
	Templates []Template `json:"templates"`
	Count     int64      `json:"count"`
}

func (t *Templates) FromStorage(storageItems []storagemodel.Template, count int64) {
	for _, storageItem := range storageItems {
		var template Template
		template.FromStorageTemplate(&storageItem)
		t.Templates = append(t.Templates, template)
	}
	t.Count = count
}

type Template struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Template string `json:"template"`
}

func (t *Template) ToStorageTemplate() *storagemodel.Template {
	return &storagemodel.Template{
		ID:       t.ID,
		Name:     t.Name,
		Template: t.Template,
	}
}

func (t *Template) FromStorageTemplate(s *storagemodel.Template) {
	t.ID = s.ID
	t.Name = s.Name
	t.Template = s.Template
}

type TemplateID struct {
	ID string `json:"id"`
}
