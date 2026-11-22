// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/kui_dashboard.go; edit that file, not this copy.
// Host logic of examples/dashboard/main.cpp: fills #table-body with seven .table-row rows, makes
// the clicked nav item the only .active one, and Refresh writes a random 1400..1599 session
// count into #stat-sessions.
package main

import (
	"fmt"
	"math/rand"
	"strconv"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

type rowData struct {
	name, status, badgeClass, value, change string
	positive                                bool
}

var rows = []rowData{
	{"Project Alpha", "Active", "badge-success", "$12,400", "+5.2%", true},
	{"Widget Pro", "Active", "badge-success", "$8,200", "+12.1%", true},
	{"API Gateway", "Warning", "badge-warning", "$4,800", "-1.5%", false},
	{"Mobile App", "Active", "badge-success", "$15,300", "+8.7%", true},
	{"Auth Service", "Critical", "badge-danger", "$2,100", "-15.3%", false},
	{"Data Pipeline", "Active", "badge-success", "$6,700", "+3.4%", true},
	{"ML Platform", "Warning", "badge-warning", "$9,500", "+0.2%", true},
}

var navIDs = []string{"nav-overview", "nav-analytics", "nav-reports", "nav-settings"}

func ready(ctx *runner.Context) error {
	doc := ctx.Document
	body, err := doc.GetElementByID("table-body")
	if err != nil || body == nil {
		return fmt.Errorf("没有 #table-body: %v", err)
	}
	defer body.Close()
	for _, row := range rows {
		rowElem, err := doc.CreateElement("div")
		if err != nil {
			return err
		}
		color := "var(--danger)"
		if row.positive {
			color = "var(--success)"
		}
		_ = rowElem.SetClassName("table-row")
		_ = rowElem.SetInnerHTML(fmt.Sprintf(`<div class="col-name">%s</div>`+
			`<div class="col-status"><span class="badge %s">%s</span></div>`+
			`<div class="col-value">%s</div>`+
			`<div class="col-change" style="color: %s">%s</div>`,
			row.name, row.badgeClass, row.status, row.value, color, row.change))
		err = body.AppendChild(rowElem)
		rowElem.Close()
		if err != nil {
			return err
		}
	}

	for _, id := range navIDs {
		item, err := doc.GetElementByID(id)
		if err != nil || item == nil {
			return fmt.Errorf("没有 #%s: %v", id, err)
		}
		ctx.Keep(item)
		if _, err := item.AddEventListener("click", func(e *kui.Event) {
			for _, nid := range navIDs {
				if el, err := doc.GetElementByID(nid); err == nil && el != nil {
					_ = el.RemoveClass("active")
					el.Close()
				}
			}
			if target, err := e.Target(); err == nil && target != nil {
				_ = target.AddClass("active")
				target.Close()
			}
		}); err != nil {
			return err
		}
	}

	refresh, err := doc.GetElementByID("refreshBtn")
	if err != nil || refresh == nil {
		return fmt.Errorf("没有 #refreshBtn: %v", err)
	}
	ctx.Keep(refresh)
	_, err = refresh.AddEventListener("click", func(*kui.Event) {
		if sessions, err := doc.GetElementByID("stat-sessions"); err == nil && sessions != nil {
			_ = sessions.SetTextContent(strconv.Itoa(1400 + rand.Intn(200)))
			sessions.Close()
		}
	})
	return err
}
