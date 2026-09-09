package ui

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/paradoxe35/encre/internal/stt"
)

// ModelList shows the catalog with per-row download state. Rows are recycled by
// widget.List, so every field is rebuilt in update rather than captured.
type ModelList struct {
	widget.BaseWidget

	store    *stt.Store
	host     stt.Machine
	window   fyne.Window
	selected string
	onSelect func(stt.Model)

	mu       sync.Mutex
	filtered []stt.Model
	progress map[string]stt.Progress
	rows     map[fyne.CanvasObject]*modelRow

	list   *widget.List
	search *widget.Entry
	filter *widget.Select
	active *widget.Label
}

func NewModelList(store *stt.Store, window fyne.Window, selected string, onSelect func(stt.Model)) *ModelList {
	m := &ModelList{
		store:    store,
		host:     stt.Host(),
		window:   window,
		selected: selected,
		onSelect: onSelect,
		progress: make(map[string]stt.Progress),
		rows:     make(map[fyne.CanvasObject]*modelRow),
	}
	m.ExtendBaseWidget(m)
	m.build()
	return m
}

func (m *ModelList) build() {
	m.list = widget.NewList(m.count, m.template, m.update)

	m.search = widget.NewEntry()
	m.search.SetPlaceHolder("Search models")

	m.filter = widget.NewSelect(
		[]string{"All", "Downloaded", "Recommended", "Multilingual", "English"}, nil)
	m.filter.SetSelected("All")
	m.active = widget.NewLabel("")
	m.active.TextStyle.Bold = true
	m.updateActiveLabel()

	// Handlers are attached after the initial selection so neither fires before
	// the list they refresh exists.
	m.search.OnChanged = func(string) { m.apply() }
	m.filter.OnChanged = func(string) { m.apply() }

	m.apply()
}

func (m *ModelList) count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.filtered)
}

func (m *ModelList) at(i widget.ListItemID) (stt.Model, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if i < 0 || i >= len(m.filtered) {
		return stt.Model{}, false
	}
	return m.filtered[i], true
}

func (m *ModelList) apply() {
	query := strings.ToLower(strings.TrimSpace(m.search.Text))
	mode := m.filter.Selected

	var out []stt.Model
	for _, model := range stt.Catalogue() {
		if !matches(model, query, mode, m.store) {
			continue
		}
		out = append(out, model)
	}
	stt.RankForMachine(out, m.host, m.store.Downloaded)

	m.mu.Lock()
	m.filtered = out
	m.mu.Unlock()
	m.list.Refresh()
}

func matches(model stt.Model, query, mode string, store *stt.Store) bool {
	switch mode {
	case "Downloaded":
		if !store.Downloaded(model) {
			return false
		}
	case "Recommended":
		if !model.Recommended {
			return false
		}
	case "Multilingual":
		if !model.Multilingual() {
			return false
		}
	case "English":
		if !model.Speaks("en") {
			return false
		}
	}

	if query == "" {
		return true
	}
	return strings.Contains(strings.ToLower(model.Name), query) ||
		strings.Contains(strings.ToLower(model.Slug), query) ||
		model.Speaks(query)
}

type modelRow struct {
	title  *widget.Label
	meta   *widget.Label
	action *widget.Button
	remove *widget.Button
}

func (m *ModelList) template() fyne.CanvasObject {
	row := &modelRow{
		title:  widget.NewLabel(""),
		meta:   widget.NewLabel(""),
		action: widget.NewButton("Get", nil),
		remove: widget.NewButtonWithIcon("", theme.DeleteIcon(), nil),
	}
	row.title.TextStyle.Bold = true
	row.title.Truncation = fyne.TextTruncateEllipsis
	row.meta.TextStyle.Italic = true
	row.meta.Truncation = fyne.TextTruncateEllipsis
	row.remove.Importance = widget.LowImportance

	content := container.NewVBox(
		container.NewBorder(nil, nil, nil,
			container.NewHBox(row.action, row.remove), row.title),
		row.meta,
	)

	// Rows are recycled, so the widgets are found by the object List hands back
	// rather than by walking the container tree.
	m.mu.Lock()
	m.rows[content] = row
	m.mu.Unlock()

	return content
}

func (m *ModelList) update(i widget.ListItemID, item fyne.CanvasObject) {
	m.mu.Lock()
	row, known := m.rows[item]
	m.mu.Unlock()
	if !known {
		return
	}

	model, ok := m.at(i)
	if !ok {
		return
	}

	row.title.SetText(model.Name)
	row.meta.SetText(summarise(model, m.host))

	m.mu.Lock()
	progress, downloading := m.progress[model.ID]
	m.mu.Unlock()

	row.remove.Hide()
	row.action.Show()

	switch {
	case downloading && progress.Stage == stt.StageVerifying:
		row.action.SetText("Verifying")
		row.action.OnTapped = nil
		row.action.Disable()

	case downloading:
		row.action.SetText(fmt.Sprintf("Cancel (%.0f%%)", progress.Fraction()*100))
		row.action.OnTapped = func() { m.store.CancelDownload(model) }
		row.action.Enable()

	case m.store.Downloaded(model):
		row.remove.Show()
		row.remove.OnTapped = func() { m.confirmDelete(model) }
		if model.ID == m.selected {
			row.action.SetText("In use")
			row.action.OnTapped = nil
			row.action.Disable()
		} else {
			row.action.SetText("Use")
			row.action.OnTapped = func() { m.choose(model) }
			row.action.Enable()
		}

	default:
		row.action.SetText(fmt.Sprintf("Get %.0f MB", model.SizeMB()))
		row.action.OnTapped = func() { m.download(model) }
		row.action.Enable()
	}

	row.action.Refresh()
}

// summarise keeps a row to one short line. Word error rates and realtime
// factors are what a benchmark wants; what a person choosing wants is whether
// it speaks their language, how big it is, and whether it will keep up here.
func summarise(model stt.Model, host stt.Machine) string {
	parts := []string{model.LanguageSummary(), fmt.Sprintf("%.0f MB", model.SizeMB())}

	switch model.Fit(host) {
	case stt.FitTooLarge:
		parts = append(parts, "too large for this machine")
	case stt.FitSlow:
		parts = append(parts, "slow here")
	default:
		if model.EstimatedRealtime(host) >= 10 {
			parts = append(parts, "very fast here")
		} else {
			parts = append(parts, "fast here")
		}
	}

	return strings.Join(parts, " · ")
}

func (m *ModelList) choose(model stt.Model) {
	m.selected = model.ID
	m.updateActiveLabel()
	if m.onSelect != nil {
		m.onSelect(model)
	}
	m.list.Refresh()
}

func (m *ModelList) download(model stt.Model) {
	m.mu.Lock()
	m.progress[model.ID] = stt.Progress{Model: model, Total: model.SizeBytes, Stage: stt.StageDownloading}
	m.mu.Unlock()
	m.list.Refresh()

	go func() {
		err := m.store.Download(context.Background(), model, func(p stt.Progress) {
			m.mu.Lock()
			m.progress[model.ID] = p
			m.mu.Unlock()
			fyne.Do(m.list.Refresh)
		})

		m.mu.Lock()
		delete(m.progress, model.ID)
		m.mu.Unlock()

		fyne.Do(func() {
			m.list.Refresh()
			if err != nil && !errorsIsCancelled(err) {
				dialog.ShowError(err, m.window)
				return
			}
			// First model downloaded becomes the one in use.
			if err == nil && m.selected == "" {
				m.choose(model)
			}
		})
	}()
}

func (m *ModelList) confirmDelete(model stt.Model) {
	dialog.ShowConfirm("Delete model",
		fmt.Sprintf("Remove %s? You can download it again later.", model.Name),
		func(confirmed bool) {
			if !confirmed {
				return
			}
			if err := m.store.Delete(model); err != nil {
				dialog.ShowError(err, m.window)
				return
			}
			if m.selected == model.ID {
				m.selected = ""
				m.updateActiveLabel()
			}
			m.list.Refresh()
		}, m.window)
}

func errorsIsCancelled(err error) bool {
	return err == context.Canceled || strings.Contains(err.Error(), "context canceled")
}

func (m *ModelList) CreateRenderer() fyne.WidgetRenderer {
	header := container.NewBorder(nil, nil, nil, m.filter, m.search)
	return widget.NewSimpleRenderer(container.NewBorder(
		container.NewVBox(m.active, header), nil, nil, nil, m.list))
}

func (m *ModelList) updateActiveLabel() {
	if m.active == nil {
		return
	}
	m.active.SetText(activeModelText(m.selected, stt.Catalogue()))
}

func activeModelText(selected string, models []stt.Model) string {
	for _, model := range models {
		if model.ID == selected {
			return "Active model: " + model.Name
		}
	}
	return "No active model selected"
}
