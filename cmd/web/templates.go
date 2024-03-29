package main

import (
	"html/template"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackcode/suitenet/pkg/forms"
	"github.com/jackcode/suitenet/pkg/models"
)

type templateData struct {
	AuthenticatedUser *models.SysUser
	ClockedInUsers    []*models.SysUser
	CSRFToken         string
	CurrentYear       int
	Flash             string
	Department        string
	Form              *forms.Form
	Request           *models.Request
	RequestNotes      []*models.RequestNote
	Requests          []*models.Request
	Locations         []*models.Location
	Positions         []*models.Position
	Users             []*models.SysUser
}

func humanDate(t time.Time) string {
	return t.Local().Format("1/2/06 at 3:04 PM")
}

func lastElement(length int) int {
	return length - 1
}

// Checks SiteRole slice for given role
func rolesContain(roles []*models.SiteRole, role string) bool {
	for _, a := range roles {
		if a.Title == role {
			return true
		}
	}
	return false
}

func capitalizeFirstLetter(word string) string {
	return strings.Title(strings.ToLower(word))
}

// Initialize a template.FuncMap object and store it in a global variable. This is
// essentially a string-keyed map which acts as a lookup between the names of our
// custom template functions and the functions themselves.
var functions = template.FuncMap{
	"humanDate":             humanDate,
	"lastElement":           lastElement,
	"rolesContain":          rolesContain,
	"capitalizeFirstLetter": capitalizeFirstLetter,
}

func newTemplateCache(dir string) (map[string]*template.Template, error) {
	// Initialize a new map to act as the cache.
	cache := map[string]*template.Template{}

	// Use the filepath.Glob function to get a slice of all filepaths with
	// the extension '.page.tmpl'. This essentially gives us a slice of all the
	// 'page' templates for the application.
	pages, err := filepath.Glob(filepath.Join(dir, "*.page.tmpl"))
	if err != nil {
		return nil, err
	}

	// Loop through the pages one-by-one.
	for _, page := range pages {
		// Extract the file name (like 'home.page.tmpl') from the full file path
		// and assign it to the name variable.
		name := filepath.Base(page)

		// Parse the page template file in to a template set.
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return nil, err
		}

		// Use the ParseGlob method to add any 'layout' templates to the
		// template set (in our case, it's just the 'base' layout at the
		// moment).
		ts, err = ts.ParseGlob(filepath.Join(dir, "*.layout.tmpl"))
		if err != nil {
			return nil, err
		}

		// Use the ParseGlob method to add any 'partial' templates to the
		// template set (in our case, it's just the 'footer' partial at the
		// moment).
		ts, err = ts.ParseGlob(filepath.Join(dir, "*.partial.tmpl"))
		if err != nil {
			return nil, err
		}

		// Add the template set to the cache, using the name of the page
		// (like 'home.page.tmpl') as the key.
		cache[name] = ts
	}

	// Return the map.
	return cache, nil
}
