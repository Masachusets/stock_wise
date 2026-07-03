package web

import (
	"context"
	"html/template"
	"net/http"
	"strings"

	"goa.design/clue/log"
)

func normalizeDate(s *string) *string {
	if s == nil || *s == "" {
		return nil
	}
	v := *s
	if len(v) == 7 {
		v += "-01"
	}
	return &v
}

var funcMap = template.FuncMap{
	"statusLabel": func(s string) string {
		labels := map[string]string{
			"exp":         "Эксплуатируемое",
			"exp_int":     "Эксплуатируемое (интернет)",
			"exp_sp":      "Эксплуатируемое (категорир.)",
			"broken":      "Неисправное",
			"written_off": "Списанное",
		}
		if l, ok := labels[s]; ok {
			return l
		}
		return s
	},
	"statusLabelWaybill": func(s string) string {
		labels := map[string]string{
			"draft":    "Черновик",
			"signed":   "Подписан",
			"archived": "В архиве",
		}
		if l, ok := labels[s]; ok {
			return l
		}
		return s
	},
	"targetLabel": func(s string) string {
		labels := map[string]string{
			"employee":   "Сотрудник",
			"department": "Подразделение",
			"warehouse":  "Склад",
		}
		if l, ok := labels[s]; ok {
			return l
		}
		return s
	},
}

func LoadTemplates() *template.Template {
	return template.Must(
		template.New("").Funcs(funcMap).ParseGlob("web/templates/*.html"),
	)
}

func RenderPage(w http.ResponseWriter, tpl *template.Template, pageTmpl string, data map[string]interface{}) {
	var buf strings.Builder
	if err := tpl.ExecuteTemplate(&buf, pageTmpl, data); err != nil {
		log.Errorf(context.Background(), err, "template error")
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
	data["Content"] = template.HTML(buf.String())
	if err := tpl.ExecuteTemplate(w, "layout.html", data); err != nil {
		log.Errorf(context.Background(), err, "layout error")
	}
}
