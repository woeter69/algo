package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/flosch/pongo2/v6"
)

var (
	templateSet *pongo2.TemplateSet
)

// SessionDict provides Jinja-compatible session access in Pongo2 templates
type SessionDict struct {
	values map[interface{}]interface{}
}

func (s SessionDict) Get(key string, defaultVal ...interface{}) interface{} {
	if s.values != nil {
		if v, ok := s.values[key]; ok && v != nil && v != "" {
			return v
		}
	}
	if len(defaultVal) > 0 {
		return defaultVal[0]
	}
	return nil
}

func (s SessionDict) Username() string {
	if s.values != nil {
		if v, ok := s.values["username"].(string); ok {
			return v
		}
	}
	return "User"
}

func (s SessionDict) Role() string {
	if s.values != nil {
		if v, ok := s.values["role"].(string); ok {
			return v
		}
	}
	return "user"
}

func (s SessionDict) UserID() int {
	if s.values != nil {
		if v, ok := s.values["user_id"].(int); ok {
			return v
		}
	}
	return 0
}

func (s SessionDict) PfpPath() string {
	if s.values != nil {
		if v, ok := s.values["pfp_path"].(string); ok {
			return v
		}
	}
	return ""
}

// PreprocessTemplate cleans up Jinja2-isms for Pongo2 compatibility
func PreprocessTemplate(input string) string {
	// 1. Remove newlines inside {{ ... }} and {% ... %}
	reTags := regexp.MustCompile(`(?s)(\{\{.*?\}\}|\{%.*?%\})`)
	s := reTags.ReplaceAllStringFunc(input, func(m string) string {
		lines := strings.Split(m, "\n")
		var clean []string
		for _, l := range lines {
			clean = append(clean, strings.TrimSpace(l))
		}
		return strings.Join(clean, " ")
	})

	// 2. Fix {% with messages = get_flashed_messages(...) %} -> {% with messages=get_flashed_messages() %}
	reWith := regexp.MustCompile(`\{%\s*with\s+messages\s*=\s*get_flashed_messages\([^)]*\)\s*%\}`)
	s = reWith.ReplaceAllString(s, `{% with messages=get_flashed_messages() %}`)

	// 3. Replace url_for('static', filename='...') with "/static/..."
	reStatic := regexp.MustCompile(`url_for\s*\(\s*['"]static['"]\s*,\s*filename\s*=\s*['"]([^'"]+)['"]\s*\)`)
	s = reStatic.ReplaceAllString(s, `"/static/$1"`)

	// 4. Replace url_for('static', '...') with "/static/..."
	reStatic2 := regexp.MustCompile(`url_for\s*\(\s*['"]static['"]\s*,\s*['"]([^'"]+)['"]\s*\)`)
	s = reStatic2.ReplaceAllString(s, `"/static/$1"`)

	// 5. Replace url_for('dashboard.resolve_contact_query', query_id=inquiry.id)
	reUrlQuery := regexp.MustCompile(`url_for\s*\(\s*['"]dashboard\.resolve_contact_query['"]\s*,\s*query_id\s*=\s*([^)]+)\)`)
	s = reUrlQuery.ReplaceAllString(s, `"/admin/contact/" + $1 + "/resolve"`)

	reUrlProfile := regexp.MustCompile(`url_for\s*\(\s*['"]profile['"]\s*,\s*username\s*=\s*([^)]+)\)`)
	s = reUrlProfile.ReplaceAllString(s, `"/profile/" + $1`)

	// 6. Handle inline ternary {{ A if COND else B }} -> {% if COND %}{{ A }}{% else %}{{ B }}{% endif %}
	reTernary := regexp.MustCompile(`\{\{\s*([^}]+?)\s+if\s+([^}]+?)\s+else\s+([^}]+?)\s*\}\}`)
	s = reTernary.ReplaceAllString(s, `{% if $2 %}{{ $1 }}{% else %}{{ $3 }}{% endif %}`)

	// 7. Fix method calls: .title() -> |title, .lower() -> |lower, .upper() -> |upper
	s = strings.ReplaceAll(s, `.title()`, `|title`)
	s = strings.ReplaceAll(s, `.lower()`, `|lower`)
	s = strings.ReplaceAll(s, `.upper()`, `|upper`)

	// 8. Fix slice indexing: foo[0] -> foo.0, bar[1][2] -> bar.1.2
	reIndex := regexp.MustCompile(`([a-zA-Z0-9_]+)\[([0-9]+)\]`)
	s = reIndex.ReplaceAllString(s, `$1.$2`)
	s = reIndex.ReplaceAllString(s, `$1.$2`) // twice for double brackets foo[0][1]

	// 9. Fix 'is none' -> '== nil'
	s = strings.ReplaceAll(s, ` is none`, ` == nil`)
	s = strings.ReplaceAll(s, ` is not none`, ` != nil`)

	return s
}

type PreprocessingTemplateLoader struct {
	baseDir string
}

func (l *PreprocessingTemplateLoader) Abs(base, name string) string {
	return name
}

func (l *PreprocessingTemplateLoader) Get(path string) (io.Reader, error) {
	fullPath := l.baseDir + "/" + path
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}
	clean := PreprocessTemplate(string(data))
	return strings.NewReader(clean), nil
}

func initTemplates() {
	templateDir := "../templates"
	if _, err := os.Stat(templateDir); os.IsNotExist(err) {
		templateDir = "templates"
		if _, err := os.Stat(templateDir); os.IsNotExist(err) {
			templateDir = "app/templates"
		}
	}

	loader := &PreprocessingTemplateLoader{baseDir: templateDir}
	templateSet = pongo2.NewSet("main", loader)

	// Register filters
	_ = pongo2.RegisterFilter("tojson", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		b, err := json.Marshal(in.Interface())
		if err != nil {
			return pongo2.AsValue("{}"), nil
		}
		return pongo2.AsSafeValue(string(b)), nil
	})

	_ = pongo2.RegisterFilter("random", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		if in.CanSlice() {
			sliceLen := in.Len()
			if sliceLen > 0 {
				idx := rand.Intn(sliceLen)
				return in.Index(idx), nil
			}
		}
		return pongo2.AsValue(rand.Intn(1000)), nil
	})

	_ = pongo2.RegisterFilter("string", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		return pongo2.AsValue(fmt.Sprint(in.Interface())), nil
	})

	_ = pongo2.RegisterFilter("length", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		return pongo2.AsValue(in.Len()), nil
	})

	_ = pongo2.RegisterFilter("replace", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		str := in.String()
		oldStr := param.String()
		return pongo2.AsValue(strings.ReplaceAll(str, oldStr, "")), nil
	})
}

var routeMap = map[string]string{
	"core.home":                  "/",
	"core.about":                 "/about",
	"core.contact":               "/contact",
	"core.thanks":                "/thanks",
	"auth.login":                 "/login",
	"auth.register":              "/register",
	"auth.logout":                "/logout",
	"auth.check_email":           "/check_email",
	"auth.confirmation":          "/confirmation",
	"dashboard.user_dashboard":   "/dashboard",
	"dashboard.admin_dashboard":  "/admin_dashboard",
	"dashboard.limited_dashboard": "/limited_dashboard",
	"dashboard.verification_request": "/verification_request",
	"profile.user_profile":       "/profile",
	"profile.complete_profile":   "/complete_profile",
	"profile.interests":          "/interests",
	"connections.connect":        "/connect",
	"connections.requests_page":  "/requests",
	"communities.channels":       "/channels",
	"chat.chat_list":             "/chat",
	"chat.chat":                  "/chat",
	"settings.settings_page":     "/settings",
	"create_community":           "/create_community",
}

func renderTemplate(w http.ResponseWriter, r *http.Request, name string, context pongo2.Context) {
	tmpl, err := templateSet.FromFile(name)
	if err != nil {
		log.Printf("❌ Error loading template %s: %v", name, err)
		http.Error(w, fmt.Sprintf("Error loading template %s: %v", name, err), http.StatusInternalServerError)
		return
	}

	if context == nil {
		context = make(pongo2.Context)
	}

	session, _ := store.Get(r, "algo-session")
	var sessionVals map[interface{}]interface{}
	if session != nil {
		sessionVals = session.Values
	}
	context["session"] = SessionDict{values: sessionVals}

	// Flash messages
	context["get_flashed_messages"] = func(options ...interface{}) [][]string {
		return popFlashes(w, r)
	}

	context["request"] = r

	// Jinja url_for helper
	context["url_for"] = func(endpoint string, args ...interface{}) string {
		if endpoint == "static" {
			for _, arg := range args {
				if str, ok := arg.(string); ok && strings.HasPrefix(str, "filename=") {
					return "/static/" + strings.TrimPrefix(str, "filename=")
				}
			}
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
		if strings.HasPrefix(endpoint, "/") {
			return endpoint
		}
		return "/" + endpoint
	}

	// Avatar generator
	context["generate_default_avatar"] = func(name string) string {
		return GenerateDefaultAvatar(name)
	}

	// IST time helpers
	context["to_ist"] = func(t time.Time) time.Time {
		return ToIST(t)
	}

	context["format_ist_time"] = func(t interface{}, format ...string) string {
		switch v := t.(type) {
		case time.Time:
			return FormatISTTime(v, format...)
		case *time.Time:
			if v != nil {
				return FormatISTTime(*v, format...)
			}
			return "now"
		default:
			return fmt.Sprint(t)
		}
	}

	// range function helper: range(start, end)
	context["range"] = func(args ...int) []int {
		start := 0
		end := 10
		if len(args) == 1 {
			end = args[0]
		} else if len(args) >= 2 {
			start = args[0]
			end = args[1]
		}
		var res []int
		for i := start; i < end; i++ {
			res = append(res, i)
		}
		return res
	}

	err = tmpl.ExecuteWriter(context, w)
	if err != nil {
		log.Printf("❌ Error executing template %s: %v", name, err)
		http.Error(w, fmt.Sprintf("Error executing template %s: %v", name, err), http.StatusInternalServerError)
	}
}
