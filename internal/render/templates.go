package render

import (
	"errors"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"text/template"

	"github.com/cloud-barista/cm-ant/internal/utils"
	"github.com/labstack/echo/v4"
)

type Template struct { //the map[key] in key means 'Your html file name'
	templates map[string]*template.Template
}

func NewTemplate() *Template {
	templates, err := generateTemplateCache()

	if err != nil {
		log.Fatal("template parsing error", err)
	}

	return &Template{
		templates: templates,
	}
}

func (t *Template) Render(w io.Writer, templateName string, data interface{}, c echo.Context) error {
	if tmpl, exist := t.templates[templateName]; exist { //Check existence of the t.templates[html_name]
		return tmpl.Execute(w, data) // ** It wll execute the map[string]interface{} data
	} else {
		return errors.New("There is no " + templateName + " in Template map.")
	}

}

func generateTemplateCache() (map[string]*template.Template, error) {
	pathToTemplates := utils.JoinRootPathWith("/web/templates")
	templateCache := map[string]*template.Template{}
	pages, err := filepath.Glob(fmt.Sprintf("%s/*.page.tmpl", pathToTemplates))
	if err != nil {
		return templateCache, err
	}

	layouts, err := filepath.Glob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates))
	if err != nil {
		return templateCache, err
	}

	for _, page := range pages {
		name := filepath.Base(page)
		createdTemplate, err := template.New(name).ParseFiles(page)
		if err != nil {
			return templateCache, err
		}

		if len(layouts) > 0 {
			associatedTemplateWithLayout, err := createdTemplate.ParseGlob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates))
			if err != nil {
				return templateCache, err
			}
			createdTemplate = associatedTemplateWithLayout
		}

		templateCache[name] = createdTemplate
	}

	return templateCache, nil
}
