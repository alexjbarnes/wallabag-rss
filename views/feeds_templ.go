// Code generated by templ - DO NOT EDIT.

// templ: version: v0.3.906
package views

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

import "wallabag-rss-tool/pkg/models"
import "strconv"

type FeedsData struct {
	PageData
	Feeds               []models.Feed
	DefaultPollInterval int
}

func getFeedPollIntervalValue(feed models.Feed) string {
	if feed.PollInterval == 0 {
		return "0"
	}
	return strconv.Itoa(feed.PollInterval)
}

func Feeds(data FeedsData) templ.Component {
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
		templ_7745c5c3_Var2 := templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
			templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
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
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 1, "<div class=\"container mt-4\"><h1>Manage RSS Feeds</h1><p>Add, edit, or remove RSS feeds that Wallabag RSS Tool will monitor.</p><div class=\"card mb-4\"><div class=\"card-header\">Add New Feed</div><div class=\"card-body\"><form hx-post=\"/feeds\" hx-target=\"#feeds-list\" hx-swap=\"beforeend\" hx-on::after-request=\"this.reset()\"><input type=\"hidden\" name=\"csrf_token\" value=\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var3 string
			templ_7745c5c3_Var3, templ_7745c5c3_Err = templ.JoinStringErrs(data.CSRFToken)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/feeds.templ`, Line: 30, Col: 67}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var3))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 2, "\"><div class=\"mb-3\"><label for=\"feedName\" class=\"form-label\">Feed Name</label> <input type=\"text\" class=\"form-control\" id=\"feedName\" name=\"name\" required></div><div class=\"mb-3\"><label for=\"feedURL\" class=\"form-label\">Feed URL</label> <input type=\"url\" class=\"form-control\" id=\"feedURL\" name=\"url\" required></div><div class=\"mb-3\"><label for=\"pollInterval\" class=\"form-label\">Poll Interval (Current default:  ")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if data.DefaultPollInterval == 1440 {
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 3, "1 day ")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			} else if data.DefaultPollInterval == 60 {
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 4, "1 hour ")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			} else if data.DefaultPollInterval%1440 == 0 {
				var templ_7745c5c3_Var4 string
				templ_7745c5c3_Var4, templ_7745c5c3_Err = templ.JoinStringErrs(strconv.Itoa(data.DefaultPollInterval / 1440))
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/feeds.templ`, Line: 46, Col: 54}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var4))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 5, " days ")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			} else if data.DefaultPollInterval%60 == 0 {
				var templ_7745c5c3_Var5 string
				templ_7745c5c3_Var5, templ_7745c5c3_Err = templ.JoinStringErrs(strconv.Itoa(data.DefaultPollInterval / 60))
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/feeds.templ`, Line: 48, Col: 52}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var5))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 6, " hours ")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			} else {
				var templ_7745c5c3_Var6 string
				templ_7745c5c3_Var6, templ_7745c5c3_Err = templ.JoinStringErrs(strconv.Itoa(data.DefaultPollInterval))
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/feeds.templ`, Line: 50, Col: 49}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var6))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 7, " minutes ")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 8, ")</label><div class=\"row\"><div class=\"col-md-6\"><input type=\"number\" class=\"form-control\" id=\"pollInterval\" name=\"poll_interval\" value=\"0\" min=\"0\" disabled></div><div class=\"col-md-6\"><select class=\"form-control\" id=\"pollIntervalUnit\" name=\"poll_interval_unit\" onchange=\"togglePollInterval()\"><option value=\"default\" selected>Default</option> <option value=\"minutes\">Minutes</option> <option value=\"hours\">Hours</option> <option value=\"days\">Days</option></select></div></div></div><div class=\"mb-3\"><label for=\"syncMode\" class=\"form-label\">Historical Articles Sync</label> <select class=\"form-control\" id=\"syncMode\" name=\"sync_mode\" onchange=\"toggleSyncOptions()\"><option value=\"none\">None - Only sync new articles from now</option> <option value=\"all\">All - Sync all available articles</option> <option value=\"count\">Count - Sync last N articles</option> <option value=\"date_from\">Date From - Sync articles from specific date</option></select></div><div class=\"mb-3\" id=\"syncCountDiv\" style=\"display: none;\"><label for=\"syncCount\" class=\"form-label\">Number of Articles</label> <input type=\"number\" class=\"form-control\" id=\"syncCount\" name=\"sync_count\" min=\"1\" max=\"1000\" value=\"10\"></div><div class=\"mb-3\" id=\"syncDateFromDiv\" style=\"display: none;\"><label for=\"syncDateFrom\" class=\"form-label\">Sync From Date</label> <input type=\"date\" class=\"form-control\" id=\"syncDateFrom\" name=\"sync_date_from\"></div><button type=\"submit\" class=\"btn btn-primary\">Add Feed</button></form></div></div><h2>Existing Feeds</h2><div id=\"feeds-list\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			for _, feed := range data.Feeds {
				templ_7745c5c3_Err = FeedRow(feed, data.DefaultPollInterval, data.CSRFToken).Render(ctx, templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 9, "</div></div><script type=\"text/javascript\">\n\t\t\tfunction togglePollInterval() {\n\t\t\t\tvar unit = document.getElementById('pollIntervalUnit');\n\t\t\t\tvar input = document.getElementById('pollInterval');\n\t\t\t\tif (unit && input) {\n\t\t\t\t\tif (unit.value === 'default') {\n\t\t\t\t\t\tinput.disabled = true;\n\t\t\t\t\t\tinput.value = '0';\n\t\t\t\t\t} else {\n\t\t\t\t\t\tinput.disabled = false;\n\t\t\t\t\t\tinput.value = '1';\n\t\t\t\t\t}\n\t\t\t\t}\n\t\t\t}\n\t\t\t\n\t\t\tfunction toggleSyncOptions() {\n\t\t\t\tvar syncMode = document.getElementById('syncMode');\n\t\t\t\tvar countDiv = document.getElementById('syncCountDiv');\n\t\t\t\tvar dateDiv = document.getElementById('syncDateFromDiv');\n\t\t\t\t\n\t\t\t\tif (syncMode && countDiv && dateDiv) {\n\t\t\t\t\tcountDiv.style.display = syncMode.value === 'count' ? 'block' : 'none';\n\t\t\t\t\tdateDiv.style.display = syncMode.value === 'date_from' ? 'block' : 'none';\n\t\t\t\t}\n\t\t\t}\n\t\t\t\n\t\t\tfunction toggleEditPollInterval(feedId) {\n\t\t\t\tvar unitSelect = document.getElementById('editPollIntervalUnit-' + feedId);\n\t\t\t\tvar input = document.getElementById('editPollInterval-' + feedId);\n\t\t\t\t\n\t\t\t\tif (unitSelect && input) {\n\t\t\t\t\tvar unit = unitSelect.value;\n\t\t\t\t\tif (unit === 'default') {\n\t\t\t\t\t\tinput.disabled = true;\n\t\t\t\t\t\tinput.value = '0';\n\t\t\t\t\t} else {\n\t\t\t\t\t\tinput.disabled = false;\n\t\t\t\t\t\tif (input.value === '0') input.value = '1';\n\t\t\t\t\t}\n\t\t\t\t}\n\t\t\t}\n\t\t\t\n\t\t\t\n\t\t\t// Make functions globally available\n\t\t\twindow.togglePollInterval = togglePollInterval;\n\t\t\twindow.toggleSyncOptions = toggleSyncOptions;\n\t\t\twindow.toggleEditPollInterval = toggleEditPollInterval;\n\t\t\t\n\t\t\tfunction initializeEverything() {\n\t\t\t\t// Initialize main form\n\t\t\t\ttogglePollInterval();\n\t\t\t\ttoggleSyncOptions();\n\t\t\t\t\n\t\t\t\t// Add event listeners to main form\n\t\t\t\tvar pollUnit = document.getElementById('pollIntervalUnit');\n\t\t\t\tvar syncMode = document.getElementById('syncMode');\n\t\t\t\t\n\t\t\t\tif (pollUnit) {\n\t\t\t\t\tpollUnit.addEventListener('change', togglePollInterval);\n\t\t\t\t}\n\t\t\t\tif (syncMode) {\n\t\t\t\t\tsyncMode.addEventListener('change', toggleSyncOptions);\n\t\t\t\t}\n\t\t\t\t\n\t\t\t\t// Initialize edit forms\n\t\t\t\tvar editPollSelects = document.querySelectorAll('[id^=\"editPollIntervalUnit-\"]');\n\t\t\t\t\n\t\t\t\teditPollSelects.forEach(function(select) {\n\t\t\t\t\tvar feedId = select.id.replace('editPollIntervalUnit-', '');\n\t\t\t\t\ttoggleEditPollInterval(feedId);\n\t\t\t\t\tselect.addEventListener('change', function() {\n\t\t\t\t\t\ttoggleEditPollInterval(feedId);\n\t\t\t\t\t});\n\t\t\t\t});\n\t\t\t}\n\t\t\t\n\t\t\t// Initialize immediately if DOM is ready, otherwise wait\n\t\t\tif (document.readyState === 'loading') {\n\t\t\t\tdocument.addEventListener('DOMContentLoaded', initializeEverything);\n\t\t\t} else {\n\t\t\t\tinitializeEverything();\n\t\t\t}\n\t\t\t\n\t\t\t// HTMX event handlers\n\t\t\tdocument.body.addEventListener('htmx:afterSwap', function() {\n\t\t\t\tsetTimeout(initializeEverything, 100);\n\t\t\t});\n\t\t\t\n\t\t\tdocument.body.addEventListener('htmx:afterSettle', function() {\n\t\t\t\tsetTimeout(initializeEverything, 100);\n\t\t\t});\n\t\t</script>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			return nil
		})
		templ_7745c5c3_Err = Layout(data.PageData).Render(templ.WithChildren(ctx, templ_7745c5c3_Var2), templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return nil
	})
}

func FeedRow(feed models.Feed, defaultPollInterval int, csrfToken string) templ.Component {
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
		templ_7745c5c3_Var7 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var7 == nil {
			templ_7745c5c3_Var7 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 10, "<div id=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var8 string
		templ_7745c5c3_Var8, templ_7745c5c3_Err = templ.JoinStringErrs("feed-" + strconv.Itoa(feed.ID))
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/feeds.templ`, Line: 192, Col: 42}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var8))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 11, "\" class=\"card mb-2\"><div class=\"card-body d-flex justify-content-between align-items-center\"><div><h5 class=\"card-title\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var9 string
		templ_7745c5c3_Var9, templ_7745c5c3_Err = templ.JoinStringErrs(feed.Name)
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/feeds.templ`, Line: 195, Col: 38}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var9))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 12, "</h5><p class=\"card-text mb-0\"><small class=\"text-muted\">URL: ")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var10 string
		templ_7745c5c3_Var10, templ_7745c5c3_Err = templ.JoinStringErrs(feed.URL)
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/feeds.templ`, Line: 196, Col: 71}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var10))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 13, "</small></p><p class=\"card-text mb-0\"><small class=\"text-muted\">Poll Interval:  ")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if feed.PollInterval == 0 {
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 14, "Default ( ")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if defaultPollInterval == 1440 {
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 15, "1 day")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			} else if defaultPollInterval == 60 {
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 16, "1 hour")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			} else if defaultPollInterval%1440 == 0 {
				var templ_7745c5c3_Var11 string
				templ_7745c5c3_Var11, templ_7745c5c3_Err = templ.JoinStringErrs(strconv.Itoa(defaultPollInterval / 1440))
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/feeds.templ`, Line: 205, Col: 47}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var11))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 17, " days")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			} else if defaultPollInterval%60 == 0 {
				var templ_7745c5c3_Var12 string
				templ_7745c5c3_Var12, templ_7745c5c3_Err = templ.JoinStringErrs(strconv.Itoa(defaultPollInterval / 60))
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/feeds.templ`, Line: 207, Col: 45}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var12))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 18, " hours")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			} else {
				var templ_7745c5c3_Var13 string
				templ_7745c5c3_Var13, templ_7745c5c3_Err = templ.JoinStringErrs(strconv.Itoa(defaultPollInterval))
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/feeds.templ`, Line: 209, Col: 42}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var13))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 19, " minutes")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 20, " )")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		} else {
			var templ_7745c5c3_Var14 string
			templ_7745c5c3_Var14, templ_7745c5c3_Err = templ.JoinStringErrs(strconv.Itoa(feed.PollInterval))
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/feeds.templ`, Line: 213, Col: 39}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var14))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 21, " ")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var15 string
			templ_7745c5c3_Var15, templ_7745c5c3_Err = templ.JoinStringErrs(string(feed.PollIntervalUnit))
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/feeds.templ`, Line: 213, Col: 73}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var15))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 22, "</small></p>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if feed.LastFetched != nil {
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 23, "<p class=\"card-text mb-0\"><small class=\"text-muted\">Last Fetched: ")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var16 string
			templ_7745c5c3_Var16, templ_7745c5c3_Err = templ.JoinStringErrs(feed.LastFetched.Format("02/01/2006 15:04:05"))
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/feeds.templ`, Line: 217, Col: 119}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var16))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 24, "</small></p>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 25, "</div><div><button class=\"btn btn-sm btn-warning me-2\" hx-get=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var17 string
		templ_7745c5c3_Var17, templ_7745c5c3_Err = templ.JoinStringErrs("/feeds/edit/" + strconv.Itoa(feed.ID))
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/feeds.templ`, Line: 221, Col: 95}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var17))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 26, "\" hx-target=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var18 string
		templ_7745c5c3_Var18, templ_7745c5c3_Err = templ.JoinStringErrs("#feed-" + strconv.Itoa(feed.ID))
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/feeds.templ`, Line: 221, Col: 142}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var18))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 27, "\" hx-swap=\"outerHTML\">Edit</button> <button class=\"btn btn-sm btn-danger\" hx-delete=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var19 string
		templ_7745c5c3_Var19, templ_7745c5c3_Err = templ.JoinStringErrs("/feeds/" + strconv.Itoa(feed.ID))
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/feeds.templ`, Line: 222, Col: 87}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var19))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 28, "\" hx-confirm=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var20 string
		templ_7745c5c3_Var20, templ_7745c5c3_Err = templ.JoinStringErrs("Are you sure you want to delete '" + feed.Name + "'?")
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/feeds.templ`, Line: 222, Col: 157}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var20))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 29, "\" hx-target=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var21 string
		templ_7745c5c3_Var21, templ_7745c5c3_Err = templ.JoinStringErrs("#feed-" + strconv.Itoa(feed.ID))
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/feeds.templ`, Line: 222, Col: 204}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var21))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 30, "\" hx-swap=\"outerHTML swap:0.5s\" hx-headers=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var22 string
		templ_7745c5c3_Var22, templ_7745c5c3_Err = templ.JoinStringErrs("{\"X-CSRF-Token\": \"" + csrfToken + "\"}")
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/feeds.templ`, Line: 222, Col: 293}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var22))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 31, "\">Delete</button></div></div></div>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return nil
	})
}

type FeedEditData struct {
	Feed                models.Feed
	DefaultPollInterval int
	CSRFToken           string
}

func FeedEditForm(data FeedEditData) templ.Component {
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
		templ_7745c5c3_Var23 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var23 == nil {
			templ_7745c5c3_Var23 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 32, "<div id=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var24 string
		templ_7745c5c3_Var24, templ_7745c5c3_Err = templ.JoinStringErrs("feed-" + strconv.Itoa(data.Feed.ID))
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/feeds.templ`, Line: 235, Col: 47}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var24))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 33, "\" class=\"card mb-2\"><div class=\"card-body\"><form hx-put=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var25 string
		templ_7745c5c3_Var25, templ_7745c5c3_Err = templ.JoinStringErrs("/feeds/" + strconv.Itoa(data.Feed.ID))
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/feeds.templ`, Line: 237, Col: 56}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var25))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 34, "\" hx-target=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var26 string
		templ_7745c5c3_Var26, templ_7745c5c3_Err = templ.JoinStringErrs("#feed-" + strconv.Itoa(data.Feed.ID))
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/feeds.templ`, Line: 237, Col: 108}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var26))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 35, "\" hx-swap=\"outerHTML\" hx-headers=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var27 string
		templ_7745c5c3_Var27, templ_7745c5c3_Err = templ.JoinStringErrs("{\"X-CSRF-Token\": \"" + data.CSRFToken + "\"}")
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/feeds.templ`, Line: 237, Col: 192}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var27))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 36, "\"><div class=\"mb-3\"><label for=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var28 string
		templ_7745c5c3_Var28, templ_7745c5c3_Err = templ.JoinStringErrs("editFeedName-" + strconv.Itoa(data.Feed.ID))
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/feeds.templ`, Line: 239, Col: 62}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var28))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 37, "\" class=\"form-label\">Feed Name</label> <input type=\"text\" class=\"form-control\" id=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var29 string
		templ_7745c5c3_Var29, templ_7745c5c3_Err = templ.JoinStringErrs("editFeedName-" + strconv.Itoa(data.Feed.ID))
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/feeds.templ`, Line: 240, Col: 94}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var29))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 38, "\" name=\"name\" value=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var30 string
		templ_7745c5c3_Var30, templ_7745c5c3_Err = templ.JoinStringErrs(data.Feed.Name)
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/feeds.templ`, Line: 240, Col: 131}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var30))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 39, "\" required></div><div class=\"mb-3\"><label for=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var31 string
		templ_7745c5c3_Var31, templ_7745c5c3_Err = templ.JoinStringErrs("editFeedURL-" + strconv.Itoa(data.Feed.ID))
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/feeds.templ`, Line: 243, Col: 61}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var31))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 40, "\" class=\"form-label\">Feed URL</label> <input type=\"url\" class=\"form-control\" id=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var32 string
		templ_7745c5c3_Var32, templ_7745c5c3_Err = templ.JoinStringErrs("editFeedURL-" + strconv.Itoa(data.Feed.ID))
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/feeds.templ`, Line: 244, Col: 92}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var32))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 41, "\" name=\"url\" value=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var33 string
		templ_7745c5c3_Var33, templ_7745c5c3_Err = templ.JoinStringErrs(data.Feed.URL)
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/feeds.templ`, Line: 244, Col: 127}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var33))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 42, "\" required></div><div class=\"mb-3\"><label for=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var34 string
		templ_7745c5c3_Var34, templ_7745c5c3_Err = templ.JoinStringErrs("editPollInterval-" + strconv.Itoa(data.Feed.ID))
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/feeds.templ`, Line: 247, Col: 66}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var34))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 43, "\" class=\"form-label\">Poll Interval (Current default:  ")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if data.DefaultPollInterval == 1440 {
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 44, "1 day ")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		} else if data.DefaultPollInterval == 60 {
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 45, "1 hour ")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		} else if data.DefaultPollInterval%1440 == 0 {
			var templ_7745c5c3_Var35 string
			templ_7745c5c3_Var35, templ_7745c5c3_Err = templ.JoinStringErrs(strconv.Itoa(data.DefaultPollInterval / 1440))
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/feeds.templ`, Line: 253, Col: 52}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var35))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 46, " days ")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		} else if data.DefaultPollInterval%60 == 0 {
			var templ_7745c5c3_Var36 string
			templ_7745c5c3_Var36, templ_7745c5c3_Err = templ.JoinStringErrs(strconv.Itoa(data.DefaultPollInterval / 60))
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/feeds.templ`, Line: 255, Col: 50}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var36))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 47, " hours ")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		} else {
			var templ_7745c5c3_Var37 string
			templ_7745c5c3_Var37, templ_7745c5c3_Err = templ.JoinStringErrs(strconv.Itoa(data.DefaultPollInterval))
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/feeds.templ`, Line: 257, Col: 47}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var37))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 48, " minutes ")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 49, ")</label><div class=\"row\"><div class=\"col-md-6\"><input type=\"number\" class=\"form-control\" id=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var38 string
		templ_7745c5c3_Var38, templ_7745c5c3_Err = templ.JoinStringErrs("editPollInterval-" + strconv.Itoa(data.Feed.ID))
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/feeds.templ`, Line: 262, Col: 102}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var38))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 50, "\" name=\"poll_interval\" value=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var39 string
		templ_7745c5c3_Var39, templ_7745c5c3_Err = templ.JoinStringErrs(getFeedPollIntervalValue(data.Feed))
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/feeds.templ`, Line: 262, Col: 169}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var39))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 51, "\" min=\"0\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if data.Feed.PollInterval == 0 {
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 52, " disabled")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 53, "></div><div class=\"col-md-6\"><select class=\"form-control\" id=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var40 string
		templ_7745c5c3_Var40, templ_7745c5c3_Err = templ.JoinStringErrs("editPollIntervalUnit-" + strconv.Itoa(data.Feed.ID))
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/feeds.templ`, Line: 265, Col: 93}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var40))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 54, "\" name=\"poll_interval_unit\"><option value=\"default\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if data.Feed.PollInterval == 0 {
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 55, " selected")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 56, ">Default</option> <option value=\"minutes\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if data.Feed.PollInterval > 0 && data.Feed.PollIntervalUnit == "minutes" {
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 57, " selected")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 58, ">Minutes</option> <option value=\"hours\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if data.Feed.PollInterval > 0 && data.Feed.PollIntervalUnit == "hours" {
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 59, " selected")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 60, ">Hours</option> <option value=\"days\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if data.Feed.PollInterval > 0 && data.Feed.PollIntervalUnit == "days" {
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 61, " selected")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 62, ">Days</option></select></div></div></div><button type=\"submit\" class=\"btn btn-primary me-2\">Save</button> <button type=\"button\" class=\"btn btn-secondary\" hx-get=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var41 string
		templ_7745c5c3_Var41, templ_7745c5c3_Err = templ.JoinStringErrs("/feeds/row/" + strconv.Itoa(data.Feed.ID))
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/feeds.templ`, Line: 275, Col: 103}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var41))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 63, "\" hx-target=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var42 string
		templ_7745c5c3_Var42, templ_7745c5c3_Err = templ.JoinStringErrs("#feed-" + strconv.Itoa(data.Feed.ID))
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/feeds.templ`, Line: 275, Col: 155}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var42))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 64, "\" hx-swap=\"outerHTML\">Cancel</button></form></div></div>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return nil
	})
}

var _ = templruntime.GeneratedTemplate
