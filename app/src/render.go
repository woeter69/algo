package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/flosch/pongo2/v6"
)

var (
	templateSet *pongo2.TemplateSet
)

func initTemplates() {
	// Root directory is app/src, templates are in app/templates
	templateDir := "../templates"
	
	// Create a new template set
	templateSet = pongo2.NewSet("main", pongo2.MustNewLocalFileSystemLoader(templateDir))

	// Global url_for implementation
	pongo2.RegisterFilter("url_for", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		// This handles the filter case, but url_for is used as a function in Jinja2
		return in, nil
	})
}

// Map of Flask-style endpoints to Go paths
var routeMap = map[string]string{
	"core.home":     "/",
	"core.about":    "/about",
	"core.contact":   "/contact",
	"core.thanks":    "/thanks",
	"auth.login":    "/auth/login",
	"auth.register": "/auth/register",
	"auth.logout":   "/auth/logout",
	"dashboard.user_dashboard": "/dashboard",
	"profile.complete_profile": "/profile/complete",
}

func renderTemplate(w http.ResponseWriter, r *http.Request, name string, context pongo2.Context) {
	tmpl, err := templateSet.FromCache(name)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error loading template %s: %v", name, err), http.StatusInternalServerError)
		return
	}

	if context == nil {
		context = make(pongo2.Context)
	}

	// Add session data to context
	session, _ := store.Get(r, "algo-session")
	context["session"] = session.Values

	// Mock get_flashed_messages for now
	context["get_flashed_messages"] = func(options ...interface{}) [][]string {
		// In a real implementation, we would pull these from the session
		return [][]string{}
	}

	// Add common variables
	context["request"] = r
	
	// Implementation of url_for that matches Jinja2's function call syntax
	context["url_for"] = func(endpoint string, args ...interface{}) string {
		if endpoint == "static" {
			// Find filename in args
			for _, arg := range args {
				if str, ok := arg.(string); ok && strings.HasPrefix(str, "filename=") {
					return "/static/" + strings.TrimPrefix(str, "filename=")
				}
				// Pongo2 might pass it as a simple string if called like url_for('static', 'file.css')
				// But Jinja2 uses filename='file.css'
			}
			// Fallback for different call styles
			if len(args) > 0 {
				if str, ok := args[0].(string); ok {
					return "/static/" + str
				}
			}
			return "/static/"
		}
		
		if path, ok := routeMap[endpoint]; ok {
			return path
		}
		return "/" + endpoint
	}

	err = tmpl.ExecuteWriter(context, w)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error executing template %s: %v", name, err), http.StatusInternalServerError)
	}
}

