package tui

import (
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"kiri/internal/db"
	"kiri/internal/model"

	tea "github.com/charmbracelet/bubbletea"
)

type Overlay int

const (
	OverlayNone Overlay = iota
	OverlayAddPlant
	OverlayEditPlant
	OverlayDeleteConfirm
	OverlayBackupConfirm
)

const (
	plantFieldName = iota
	plantFieldLocation
	plantFieldInterval
	plantFieldOutdoor
	plantFieldCount
)

type plantForm struct {
	focus    int
	name     string
	location string
	interval string
	outdoor  bool
}

func newPlantForm(defaultDays int) plantForm {
	return plantForm{
		interval: strconv.Itoa(db.NormalizeDefaultIntervalDays(defaultDays)),
	}
}

func newPlantFormFrom(p *model.Plant) plantForm {
	return plantForm{
		name:     p.Name,
		location: p.Location,
		interval: strconv.Itoa(p.BaseIntervalDays),
		outdoor:  p.IsOutdoor,
	}
}

func (m *Model) cancelOverlay() {
	m.overlay = OverlayNone
	m.plantForm = plantForm{}
	m.confirmBtn = 0
	m.status = ""
}

func (m *Model) openBackupConfirm() {
	m.overlay = OverlayBackupConfirm
	m.confirmBtn = 0
	m.status = ""
}

func (m *Model) openAddPlant() {
	m.overlay = OverlayAddPlant
	m.plantForm = newPlantForm(m.defaultIntervalDays)
	m.status = ""
}

func (m *Model) openEditPlant() {
	p := m.selectedPlant()
	if p == nil {
		m.status = m.cat().StatusNoPlantSelected()
		return
	}
	m.overlay = OverlayEditPlant
	m.plantForm = newPlantFormFrom(p)
	m.status = ""
}

func (m *Model) openDeleteConfirm() {
	p := m.selectedPlant()
	if p == nil {
		m.status = m.cat().StatusNoPlantSelected()
		return
	}
	m.overlay = OverlayDeleteConfirm
	m.status = ""
}

func (m Model) handleOverlayKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.overlay {
	case OverlayDeleteConfirm:
		return m.handleDeleteConfirmKey(msg)
	case OverlayBackupConfirm:
		return m.handleBackupConfirmKey(msg)
	case OverlayAddPlant, OverlayEditPlant:
		return m.handlePlantFormKey(msg)
	default:
		return m, nil
	}
}

func (m Model) handleBackupConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "left", "h", "up":
		m.confirmBtn = 0
		return m, nil
	case "right", "l", "down":
		m.confirmBtn = 1
		return m, nil
	case "esc", "n", "N":
		m.cancelOverlay()
		m.status = m.cat().SettingsBackupCancelled()
		return m, nil
	case "y", "Y":
		return m.executeBackup()
	case "enter":
		if m.confirmBtn == 1 {
			return m.executeBackup()
		}
		m.cancelOverlay()
		m.status = m.cat().SettingsBackupCancelled()
		return m, nil
	}
	return m, nil
}

func (m *Model) executeBackup() (tea.Model, tea.Cmd) {
	c := m.cat()
	path, err := m.backupDatabase()
	if err != nil {
		m.err = err
		m.cancelOverlay()
		return m, nil
	}
	m.cancelOverlay()
	m.status = c.SettingsBackupDone(path)
	return m, nil
}

func (m Model) handleDeleteConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "n", "N":
		m.cancelOverlay()
		m.status = m.cat().StatusDeleteCancelled()
		return m, nil
	case "y", "Y", "enter":
		return m.deleteSelectedPlant()
	}
	return m, nil
}

func (m *Model) deleteSelectedPlant() (tea.Model, tea.Cmd) {
	p := m.selectedPlant()
	if p == nil {
		m.cancelOverlay()
		return m, nil
	}
	c := m.cat()
	name := p.Name
	id := p.ID
	if err := m.store.DeletePlant(id); err != nil {
		m.err = err
		m.cancelOverlay()
		return m, nil
	}
	m.cancelOverlay()
	m.status = c.StatusPlantDeleted(name)
	return m, tea.Batch(loadPlants(m.store), loadCareLog(m.store))
}

func (m Model) handlePlantFormKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyRunes {
		for _, r := range msg.Runes {
			m.plantForm.appendRune(r)
		}
		return m, nil
	}

	switch msg.String() {
	case "esc":
		editing := m.overlay == OverlayEditPlant
		m.cancelOverlay()
		if editing {
			m.status = m.cat().StatusEditCancelled()
		} else {
			m.status = m.cat().StatusAddCancelled()
		}
		return m, nil
	case "tab", "down":
		m.plantForm.focus = (m.plantForm.focus + 1) % plantFieldCount
		return m, nil
	case "shift+tab", "up":
		m.plantForm.focus = (m.plantForm.focus + plantFieldCount - 1) % plantFieldCount
		return m, nil
	case "enter":
		if m.overlay == OverlayEditPlant {
			return m.submitEditPlant()
		}
		return m.submitAddPlant()
	case " ":
		if m.plantForm.focus == plantFieldOutdoor {
			m.plantForm.outdoor = !m.plantForm.outdoor
		}
		return m, nil
	case "backspace", "ctrl+h":
		m.plantForm.backspace()
		return m, nil
	}
	return m, nil
}

func (f *plantForm) appendRune(r rune) {
	if f.focus == plantFieldOutdoor {
		return
	}
	if f.focus == plantFieldInterval {
		if !unicode.IsDigit(r) {
			return
		}
		if len(f.interval) >= 3 {
			return
		}
		f.interval += string(r)
		return
	}
	if len(f.value()) >= 40 {
		return
	}
	switch f.focus {
	case plantFieldName:
		f.name += string(r)
	case plantFieldLocation:
		f.location += string(r)
	}
}

func (f *plantForm) backspace() {
	switch f.focus {
	case plantFieldName:
		f.name = trimLastRune(f.name)
	case plantFieldLocation:
		f.location = trimLastRune(f.location)
	case plantFieldInterval:
		f.interval = trimLastRune(f.interval)
	}
}

func (f *plantForm) value() string {
	switch f.focus {
	case plantFieldName:
		return f.name
	case plantFieldLocation:
		return f.location
	case plantFieldInterval:
		return f.interval
	default:
		return ""
	}
}

func trimLastRune(s string) string {
	if s == "" {
		return s
	}
	_, i := utf8.DecodeLastRuneInString(s)
	return s[:len(s)-i]
}

func (m *Model) parsePlantForm() (name, location string, days int, ok bool) {
	c := m.cat()
	name = strings.TrimSpace(m.plantForm.name)
	location = strings.TrimSpace(m.plantForm.location)
	if name == "" {
		m.status = c.StatusNameRequired()
		return "", "", 0, false
	}
	if location == "" {
		m.status = c.StatusLocationRequired()
		return "", "", 0, false
	}
	intervalStr := strings.TrimSpace(m.plantForm.interval)
	if intervalStr == "" {
		intervalStr = strconv.Itoa(db.NormalizeDefaultIntervalDays(m.defaultIntervalDays))
	}
	var err error
	days, err = strconv.Atoi(intervalStr)
	if err != nil || days < 1 || days > 365 {
		m.status = c.StatusInvalidInterval()
		return "", "", 0, false
	}
	return name, location, days, true
}

func (m *Model) submitAddPlant() (tea.Model, tea.Cmd) {
	c := m.cat()
	name, location, days, ok := m.parsePlantForm()
	if !ok {
		return m, nil
	}

	now := time.Now()
	p := model.Plant{
		Name:             name,
		Location:         location,
		BaseIntervalDays: days,
		IsOutdoor:        m.plantForm.outdoor,
		WaterLevel:       100,
		State:            model.StateNormal,
		LastUpdatedAt:    now,
		CreatedAt:        now,
	}
	model.RefreshPlantStateWithConfig(&p, m.effectiveModelCfg())
	if err := m.store.SavePlant(&p); err != nil {
		m.err = err
		return m, nil
	}
	m.logCare(p.ID, "added", c.LogPlantAdded())

	m.cancelOverlay()
	m.status = c.StatusPlantAdded(name)
	return m, tea.Batch(loadPlants(m.store), loadCareLog(m.store))
}

func (m *Model) submitEditPlant() (tea.Model, tea.Cmd) {
	c := m.cat()
	p := m.selectedPlant()
	if p == nil {
		m.cancelOverlay()
		return m, nil
	}

	name, location, days, ok := m.parsePlantForm()
	if !ok {
		return m, nil
	}

	oldDays := p.BaseIntervalDays
	p.Name = name
	p.Location = location
	p.BaseIntervalDays = days
	p.IsOutdoor = m.plantForm.outdoor
	if days > oldDays {
		p.ConsecutivePostpones = 0
	}
	model.RefreshPlantStateWithConfig(p, m.effectiveModelCfg())
	if err := m.persistPlant(p); err != nil {
		m.err = err
		return m, nil
	}
	m.logCare(p.ID, "updated", c.LogPlantUpdated(days))

	m.cancelOverlay()
	m.status = c.StatusPlantUpdated(name)
	return m, tea.Batch(loadPlants(m.store), loadCareLog(m.store), m.reloadPlantCareLog())
}

func (m Model) renderOverlay() string {
	switch m.overlay {
	case OverlayDeleteConfirm:
		return m.renderDeleteConfirm()
	case OverlayBackupConfirm:
		return m.renderBackupConfirm()
	case OverlayAddPlant:
		return m.renderPlantForm(false, "")
	case OverlayEditPlant:
		p := m.selectedPlant()
		titleName := ""
		if p != nil {
			titleName = p.Name
		}
		return m.renderPlantForm(true, titleName)
	default:
		return ""
	}
}

func (m Model) renderModal(content string) string {
	tw := m.modalTextWidth()
	return StyleModal.Width(m.modalBoxWidth()).Render(joinModalLines(strings.Split(content, "\n"), tw))
}

func (m Model) renderBackupModal(content string) string {
	tw := m.modalTextWidth()
	return StyleModalBackup.Width(m.modalBoxWidth()).Render(joinModalLines(strings.Split(content, "\n"), tw))
}

func (m Model) renderBackupConfirm() string {
	c := m.cat()
	tw := m.modalTextWidth()
	lines := []string{
		modalPadLine(modalOpaque(StyleSectionTitle).Render(truncateTerm(c.BackupConfirmTitle(), tw)), tw),
		"",
		modalStyleWrapped(StyleRowText, c.BackupConfirmBody(), tw),
		"",
		modalPadLine(m.renderConfirmButtons(), tw),
		"",
		modalStyleWrapped(StyleRowMuted, c.BackupConfirmHint(), tw),
	}
	return m.renderBackupModal(strings.Join(lines, "\n"))
}

func (m Model) renderConfirmButtons() string {
	c := m.cat()
	noLabel := c.SettingsNo()
	yesLabel := c.SettingsYes()

	var noBtn, yesBtn string
	if m.confirmBtn == 0 {
		noBtn = renderToggleOption(true, noLabel)
		yesBtn = renderToggleOption(false, yesLabel)
	} else {
		noBtn = renderToggleOption(false, noLabel)
		yesBtn = renderToggleOption(true, yesLabel)
	}
	sep := fillSpaces(2)
	return noBtn + sep + yesBtn
}

func (m Model) renderDeleteConfirm() string {
	c := m.cat()
	tw := m.modalTextWidth()
	p := m.selectedPlant()
	name := ""
	if p != nil {
		name = p.Name
	}
	lines := []string{
		modalPadLine(modalOpaque(StyleStatusUrgent).Render(truncateTerm(c.DeleteConfirmTitle(name), tw)), tw),
		"",
		modalStyleWrapped(StyleRowText, c.DeleteConfirmBody(), tw),
		"",
		modalStyleWrapped(StyleRowMuted, c.DeleteConfirmHint(), tw),
	}
	return m.renderModal(strings.Join(lines, "\n"))
}

func (m Model) renderPlantForm(editing bool, plantName string) string {
	c := m.cat()
	f := m.plantForm
	tw := m.modalTextWidth()

	title := c.AddPlantTitle()
	if editing {
		title = c.EditPlantTitle(plantName)
	}

	lines := []string{
		modalPadLine(modalOpaque(StyleSectionTitle).Render(truncateTerm(title, tw)), tw),
		m.renderFormField(c.FieldName(), f.name, f.focus == plantFieldName, tw),
		m.renderFormField(c.FieldLocation(), f.location, f.focus == plantFieldLocation, tw),
		m.renderFormField(c.FieldInterval(), f.interval, f.focus == plantFieldInterval, tw),
		m.renderFormOutdoor(c.FieldOutdoor(), f.outdoor, f.focus == plantFieldOutdoor, tw),
	}
	if m.status != "" {
		lines = append(lines, "", modalStyleWrapped(StyleStatusMsg, m.status, tw))
	}
	lines = append(lines, "", modalStyleWrapped(StyleRowMuted, c.AddPlantHint(), tw))
	return m.renderModal(strings.Join(lines, "\n"))
}

func (m Model) renderFormField(label, value string, focused bool, tw int) string {
	display := value
	if focused {
		display += "▌"
	}
	if display == "" && !focused {
		prefix := modalOpaque(StyleRowText).Render(truncateTerm(label+": ", tw))
		ellipsis := modalOpaque(StyleRowMuted).Render("…")
		return modalPadLine(prefix+ellipsis, tw)
	}
	if focused && value == "" {
		display = "▌"
	}
	line := truncateTerm(label+": "+display, tw)
	if focused {
		return modalPadLine(modalOpaque(StyleRowNameSelected).Render(line), tw)
	}
	return modalPadLine(modalOpaque(StyleRowText).Render(line), tw)
}

func (m Model) renderFormOutdoor(label string, outdoor, focused bool, tw int) string {
	val := m.cat().No()
	if outdoor {
		val = m.cat().Yes()
	}
	line := truncateTerm(label+": "+val, tw)
	if focused {
		return modalPadLine(modalOpaque(StyleRowNameSelected).Render(line), tw)
	}
	return modalPadLine(modalOpaque(StyleRowText).Render(line), tw)
}
