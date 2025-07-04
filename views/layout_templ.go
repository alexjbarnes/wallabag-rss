// Code generated by templ - DO NOT EDIT.

// templ: version: v0.3.906
package views

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

type PageData struct {
	Title     string
	CSRFToken string
}

func Layout(data PageData) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 1, "<!doctype html><html lang=\"en\"><head><meta charset=\"UTF-8\"><meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\"><title>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var2 string
		templ_7745c5c3_Var2, templ_7745c5c3_Err = templ.JoinStringErrs(data.Title)
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/layout.templ`, Line: 14, Col: 22}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var2))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 2, "</title><script src=\"https://cdn.jsdelivr.net/npm/htmx.org@1.9.12/dist/htmx.min.js\" integrity=\"sha384-ujb1lZYygJmzgSwoxRggbCHcjc0rB2XoQrxeTUQyRjrOnlCoYta87iKBWq3EsdM2\" crossorigin=\"anonymous\"></script><script src=\"https://unpkg.com/htmx.org/dist/ext/json-enc.js\"></script><link rel=\"stylesheet\" href=\"https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/css/bootstrap.min.css\" integrity=\"sha384-QWTKZyjpPEjISv5WaRU9OFeRpok6YctnYmDr5pNlyT2bRjXh0JMhjY6hW+ALEwIH\" crossorigin=\"anonymous\"><style>\n\t\t\t\tbody { \n\t\t\t\t\tpadding-top: 56px; /* Adjust for fixed navbar */\n\t\t\t\t\toverflow-x: hidden; /* Prevent horizontal scroll on body */\n\t\t\t\t}\n\t\t\t\t.navbar {\n\t\t\t\t\tz-index: 1030; /* Ensure navbar stays on top */\n\t\t\t\t\twidth: 100vw; /* Force navbar to full viewport width */\n\t\t\t\t\tposition: fixed !important;\n\t\t\t\t\ttop: 0;\n\t\t\t\t\tleft: 0;\n\t\t\t\t\tright: 0;\n\t\t\t\t}\n\t\t\t\tmain {\n\t\t\t\t\toverflow-x: auto; /* Allow horizontal scrolling in main content */\n\t\t\t\t\tmax-width: 100vw; /* Prevent main from exceeding viewport width */\n\t\t\t\t}\n\t\t\t\t/* Ensure tables don't break layout on mobile */\n\t\t\t\t.table-responsive {\n\t\t\t\t\tborder: none;\n\t\t\t\t}\n\t\t\t</style></head><body><nav class=\"navbar navbar-expand-lg navbar-dark bg-dark fixed-top\"><div class=\"container-fluid\"><a class=\"navbar-brand\" href=\"/\">Wallabag RSS</a> <button class=\"navbar-toggler\" type=\"button\" data-bs-toggle=\"collapse\" data-bs-target=\"#navbarNav\" aria-controls=\"navbarNav\" aria-expanded=\"false\" aria-label=\"Toggle navigation\"><span class=\"navbar-toggler-icon\"><svg xmlns=\"http://www.w3.org/2000/svg\" width=\"30\" height=\"30\" viewBox=\"0 0 30 30\"><path stroke=\"rgba(255, 255, 255, 0.75)\" stroke-linecap=\"round\" stroke-miterlimit=\"10\" stroke-width=\"2\" d=\"M4 7h22M4 15h22M4 23h22\"></path></svg></span></button><div class=\"collapse navbar-collapse\" id=\"navbarNav\"><ul class=\"navbar-nav me-auto mb-2 mb-lg-0\"><li class=\"nav-item\"><a class=\"nav-link\" href=\"/feeds\">Feeds</a></li><li class=\"nav-item\"><a class=\"nav-link\" href=\"/articles\">Articles</a></li><li class=\"nav-item\"><a class=\"nav-link\" href=\"/settings\">Settings</a></li></ul></div></div></nav><main class=\"container mt-4 pb-5\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templ_7745c5c3_Var1.Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 3, "</main><script src=\"https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/js/bootstrap.bundle.min.js\" integrity=\"sha384-YvpcrYf0tY3lHB60NNkmXc5s9fDVZLESaAA55NDzOxhy9GkcIdslK1eN7N6jIeHz\" crossorigin=\"anonymous\"></script></body></html>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return nil
	})
}

var _ = templruntime.GeneratedTemplate
