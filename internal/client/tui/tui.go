package tui

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/iurnickita/gophkeeper1/internal/client/model"
	"github.com/iurnickita/gophkeeper1/internal/client/service"
)

type modelState int

const (
	stateMenu modelState = iota
	stateTypeMenu
	stateForm
	stateLoading
	stateMessage
)

type actionType int

const (
	actionNone actionType = iota
	actionLogin
	actionRegister
	actionList
	actionRead
	actionDelete
	actionWriteLogin
	actionWriteText
	actionWriteBinary
	actionWriteCard
	actionBack
	actionQuit
)

type messageKind int

const (
	messageInfo messageKind = iota
	messageError
)

type menuItem struct {
	title        string
	action       actionType
	requiresAuth bool
}

type formConfig struct {
	title  string
	action actionType
	fields []formField
}

type formField struct {
	label       string
	placeholder string
	echo        textinput.EchoMode
}

type formState struct {
	config formConfig
	inputs []textinput.Model
	cursor int
}

type operationResultMsg struct {
	action  actionType
	err     error
	payload any
	message string
}

// Model описывает состояние TUI.
type Model struct {
	service service.Service

	state       modelState
	loggedIn    bool
	menuCursor  int
	writeCursor int

	form    *formState
	message string
	mkind   messageKind

	offlineNote string
}

// NewModel создает модель TUI.
func NewModel(service service.Service) Model {
	return Model{
		service: service,
		state:   stateMenu,
	}
}

// Run запускает Bubble Tea программу.
func Run(service service.Service) error {
	p := tea.NewProgram(NewModel(service), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case stateMenu:
			return m.updateMainMenu(msg)
		case stateTypeMenu:
			return m.updateWriteMenu(msg)
		case stateForm:
			return m.updateForm(msg)
		case stateMessage:
			return m.updateMessage(msg)
		case stateLoading:
			// Ignore keys while loading except quit.
			if msg.Type == tea.KeyCtrlC || msg.String() == "q" {
				return m, tea.Quit
			}
			return m, nil
		}
	case tea.WindowSizeMsg:
		// No-op: layout is simple.
		return m, nil
	case operationResultMsg:
		return m.handleResult(msg)
	}

	if m.state == stateForm && m.form != nil {
		return m.updateFormInputs(msg)
	}

	return m, nil
}

func (m Model) View() string {
	switch m.state {
	case stateMenu:
		return m.renderMainMenu()
	case stateTypeMenu:
		return m.renderWriteMenu()
	case stateForm:
		return m.renderForm()
	case stateLoading:
		return m.renderLoading()
	case stateMessage:
		return m.renderMessage()
	default:
		return ""
	}
}

func (m Model) mainMenuItems() []menuItem {
	items := []menuItem{
		{title: "Войти", action: actionLogin},
		{title: "Зарегистрироваться", action: actionRegister},
		{title: "Список данных", action: actionList, requiresAuth: true},
		{title: "Прочитать запись", action: actionRead, requiresAuth: true},
		{title: "Создать запись", action: actionBack, requiresAuth: true}, // navigation to type menu
		{title: "Удалить запись", action: actionDelete, requiresAuth: true},
		{title: "Выход", action: actionQuit},
	}
	return items
}

func (m Model) writeMenuItems() []menuItem {
	return []menuItem{
		{title: "Логин и пароль", action: actionWriteLogin, requiresAuth: true},
		{title: "Текстовая заметка", action: actionWriteText, requiresAuth: true},
		{title: "Бинарные данные (из файла)", action: actionWriteBinary, requiresAuth: true},
		{title: "Банковская карта", action: actionWriteCard, requiresAuth: true},
	}
}

func (m Model) updateMainMenu(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	}

	switch key.String() {
	case "up", "k":
		if m.menuCursor > 0 {
			m.menuCursor--
		}
	case "down", "j":
		if m.menuCursor < len(m.mainMenuItems())-1 {
			m.menuCursor++
		}
	case "enter":
		return m.handleMainMenuSelection()
	case "q", "esc":
		return m, tea.Quit
	}

	return m, nil
}

func (m Model) updateWriteMenu(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	}

	switch key.String() {
	case "up", "k":
		if m.writeCursor > 0 {
			m.writeCursor--
		}
	case "down", "j":
		if m.writeCursor < len(m.writeMenuItems())-1 {
			m.writeCursor++
		}
	case "enter":
		return m.handleWriteMenuSelection()
	case "esc":
		m.state = stateMenu
		m.writeCursor = 0
		return m, nil
	case "q":
		return m, tea.Quit
	}

	return m, nil
}

func (m Model) updateForm(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	}

	switch key.String() {
	case "esc":
		m.resetForm()
		m.state = stateMenu
		return m, nil
	case "enter":
		if m.form == nil {
			return m, nil
		}
		if m.form.cursor < len(m.form.inputs)-1 {
			m.form.cursor++
			for i := range m.form.inputs {
				if i == m.form.cursor {
					m.form.inputs[i].Focus()
				} else {
					m.form.inputs[i].Blur()
				}
			}
			return m, nil
		}
		values := make([]string, len(m.form.inputs))
		for i := range m.form.inputs {
			values[i] = strings.TrimSpace(m.form.inputs[i].Value())
		}
		m.state = stateLoading
		cmd := m.executeAction(m.form.config.action, values)
		m.resetForm()
		return m, cmd
	}

	return m.updateFormInputs(key)
}

func (m Model) updateFormInputs(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.form == nil || len(m.form.inputs) == 0 {
		return m, nil
	}
	var cmds []tea.Cmd
	for i := range m.form.inputs {
		if i == m.form.cursor {
			var cmd tea.Cmd
			m.form.inputs[i], cmd = m.form.inputs[i].Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}
	return m, tea.Batch(cmds...)
}

func (m Model) updateMessage(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	}

	switch key.String() {
	case "enter", "esc", "q":
		m.state = stateMenu
		m.message = ""
		m.offlineNote = ""
		m.mkind = messageInfo
		return m, nil
	}
	return m, nil
}

func (m Model) handleMainMenuSelection() (tea.Model, tea.Cmd) {
	items := m.mainMenuItems()
	if len(items) == 0 {
		return m, nil
	}
	item := items[m.menuCursor]

	if item.action == actionQuit {
		return m, tea.Quit
	}

	if item.requiresAuth && !m.loggedIn {
		//return m.showMessage("Требуется выполнить вход", messageError), nil
		return m.showMessage("Требуется выполнить вход", messageError)
	}

	switch item.action {
	case actionLogin:
		return m.startForm(loginForm())
	case actionRegister:
		return m.startForm(registerForm())
	case actionList:
		m.state = stateLoading
		return m, m.executeAction(actionList, nil)
	case actionRead:
		return m.startForm(readForm())
	case actionBack:
		m.state = stateTypeMenu
		m.writeCursor = 0
		return m, nil
	case actionDelete:
		return m.startForm(deleteForm())
	}
	return m, nil
}

func (m Model) handleWriteMenuSelection() (tea.Model, tea.Cmd) {
	items := m.writeMenuItems()
	if len(items) == 0 {
		return m, nil
	}
	item := items[m.writeCursor]

	if item.requiresAuth && !m.loggedIn {
		//return m.showMessage("Требуется выполнить вход", messageError), nil
		return m.showMessage("Требуется выполнить вход", messageError)
	}

	var cfg formConfig
	switch item.action {
	case actionWriteLogin:
		cfg = writeLoginForm()
	case actionWriteText:
		cfg = writeTextForm()
	case actionWriteBinary:
		cfg = writeBinaryForm()
	case actionWriteCard:
		cfg = writeCardForm()
	default:
		return m, nil
	}

	m.state = stateForm
	form := newForm(cfg)
	m.form = &form
	return m, form.init()
}

func (m Model) startForm(cfg formConfig) (tea.Model, tea.Cmd) {
	m.state = stateForm
	form := newForm(cfg)
	m.form = &form
	return m, form.init()
}

func (m Model) executeAction(action actionType, values []string) tea.Cmd {
	switch action {
	case actionLogin:
		login := values[0]
		password := values[1]
		return func() tea.Msg {
			err := m.service.Login(login, password)
			if err != nil {
				return operationResultMsg{action: action, err: err}
			}
			return operationResultMsg{action: action, message: "Успешный вход"}
		}
	case actionRegister:
		login := values[0]
		password := values[1]
		return func() tea.Msg {
			err := m.service.Register(login, password)
			if err != nil {
				return operationResultMsg{action: action, err: err}
			}
			return operationResultMsg{action: action, message: "Регистрация завершена"}
		}
	case actionList:
		return func() tea.Msg {
			list, err := m.service.List()
			return operationResultMsg{action: action, err: err, payload: list}
		}
	case actionRead:
		unitName := values[0]
		target := ""
		if len(values) > 1 {
			target = values[1]
		}
		return func() tea.Msg {
			result, err := m.service.Read(unitName, target)
			return operationResultMsg{action: action, err: err, payload: result}
		}
	case actionDelete:
		unitName := values[0]
		return func() tea.Msg {
			err := m.service.Delete(unitName)
			if err != nil {
				return operationResultMsg{action: action, err: err}
			}
			return operationResultMsg{action: action, message: "Удалено"}
		}
	case actionWriteLogin:
		unitName := values[0]
		login := values[1]
		password := values[2]
		return func() tea.Msg {
			unitValue := fmt.Sprintf("%s %s", login, password)
			err := m.service.Write(unitName, model.UnitTypeLogin, unitValue, "")
			if err != nil {
				return operationResultMsg{action: action, err: err}
			}
			return operationResultMsg{action: action, message: "Запись сохранена"}
		}
	case actionWriteText:
		unitName := values[0]
		text := ""
		if len(values) > 1 {
			text = values[1]
		}
		return func() tea.Msg {
			err := m.service.Write(unitName, model.UnitTypeText, text, "")
			if err != nil {
				return operationResultMsg{action: action, err: err}
			}
			return operationResultMsg{action: action, message: "Запись сохранена"}
		}
	case actionWriteBinary:
		unitName := values[0]
		source := values[1]
		return func() tea.Msg {
			err := m.service.Write(unitName, model.UnitTypeBinary, "", source)
			if err != nil {
				return operationResultMsg{action: action, err: err}
			}
			return operationResultMsg{action: action, message: "Запись сохранена"}
		}
	case actionWriteCard:
		unitName := values[0]
		number := values[1]
		expiry := values[2]
		name := values[3]
		surname := values[4]
		cvv := values[5]
		return func() tea.Msg {
			unitValue := fmt.Sprintf("%s %s %s %s %s", number, expiry, name, surname, cvv)
			err := m.service.Write(unitName, model.UnitTypeCard, unitValue, "")
			if err != nil {
				return operationResultMsg{action: action, err: err}
			}
			return operationResultMsg{action: action, message: "Запись сохранена"}
		}
	default:
		return nil
	}
}

func (m Model) handleResult(msg operationResultMsg) (tea.Model, tea.Cmd) {
	m.state = stateMessage
	m.mkind = messageInfo
	m.offlineNote = ""

	if msg.err != nil {
		if errors.Is(msg.err, service.ErrOffline) {
			m.mkind = messageInfo
			m.offlineNote = "Оффлайн режим: данные из кэша"
		} else {
			return m.showMessage(msg.err.Error(), messageError)
		}
	}

	switch msg.action {
	case actionLogin, actionRegister:
		m.loggedIn = true
		text := msg.message
		if text == "" {
			text = "Готово"
		}
		return m.showMessage(text, messageInfo)
	case actionList:
		var builder strings.Builder
		items, _ := msg.payload.([]string)
		if len(items) == 0 {
			builder.WriteString("Нет данных")
		} else {
			for idx, item := range items {
				builder.WriteString(fmt.Sprintf("%d. %s\n", idx+1, item))
			}
		}
		if m.offlineNote != "" {
			builder.WriteString("\n" + m.offlineNote)
		}
		return m.showMessage(strings.TrimSpace(builder.String()), messageInfo)
	case actionRead:
		value, _ := msg.payload.(string)
		message := value
		if message == "" {
			message = "Пустое значение"
		}
		if m.offlineNote != "" {
			message += "\n\n" + m.offlineNote
		}
		return m.showMessage(message, messageInfo)
	case actionDelete, actionWriteLogin, actionWriteText, actionWriteBinary, actionWriteCard:
		text := msg.message
		if text == "" {
			text = "Готово"
		}
		return m.showMessage(text, messageInfo)
	default:
		return m.showMessage("Операция завершена", messageInfo)
	}
}

func (m Model) showMessage(text string, kind messageKind) (tea.Model, tea.Cmd) {
	m.state = stateMessage
	m.message = text
	m.mkind = kind
	return m, nil
}

func (m *Model) resetForm() {
	m.form = nil
}

func (m Model) renderMainMenu() string {
	var b strings.Builder
	header := lipgloss.NewStyle().Bold(true).Render("GophKeeper TUI")
	b.WriteString(header)
	b.WriteString("\n\n")

	items := m.mainMenuItems()
	for i, item := range items {
		cursor := " "
		if i == m.menuCursor {
			cursor = ">"
		}
		title := item.title
		if item.requiresAuth && !m.loggedIn {
			title = fmt.Sprintf("%s (требуется вход)", title)
		}
		b.WriteString(fmt.Sprintf("%s %s\n", cursor, title))
	}
	b.WriteString("\nUp/Down для выбора - Enter подтвердить - q для выхода")
	return b.String()
}

func (m Model) renderWriteMenu() string {
	var b strings.Builder
	header := lipgloss.NewStyle().Bold(true).Render("Тип создаваемой записи")
	b.WriteString(header)
	b.WriteString("\n\n")

	items := m.writeMenuItems()
	for i, item := range items {
		cursor := " "
		if i == m.writeCursor {
			cursor = ">"
		}
		b.WriteString(fmt.Sprintf("%s %s\n", cursor, item.title))
	}
	b.WriteString("\nUp/Down для выбора - Enter подтвердить - Esc назад")
	return b.String()
}

func (m Model) renderForm() string {
	if m.form == nil {
		return ""
	}
	var b strings.Builder
	header := lipgloss.NewStyle().Bold(true).Render(m.form.config.title)
	b.WriteString(header)
	b.WriteString("\n\n")
	for i := range m.form.inputs {
		input := m.form.inputs[i]
		b.WriteString(input.View())
		b.WriteString("\n")
	}
	b.WriteString("\nEnter - далее/сохранить - Esc - отмена")
	return b.String()
}

func (m Model) renderLoading() string {
	style := lipgloss.NewStyle().Bold(true)
	return style.Render("Выполняется операция...")
}

func (m Model) renderMessage() string {
	style := lipgloss.NewStyle()
	if m.mkind == messageError {
		style = style.Foreground(lipgloss.Color("9"))
	} else {
		style = style.Foreground(lipgloss.Color("10"))
	}
	var b strings.Builder
	header := style.Render(m.message)
	b.WriteString(header)
	b.WriteString("\n\nEnter или Esc - вернуться к меню")
	return b.String()
}

func newForm(cfg formConfig) formState {
	form := formState{config: cfg}
	form.inputs = make([]textinput.Model, len(cfg.fields))
	for i, field := range cfg.fields {
		input := textinput.New()
		input.Prompt = field.label + ": "
		input.Placeholder = field.placeholder
		input.CharLimit = 256
		input.EchoMode = field.echo
		if field.echo == textinput.EchoPassword {
			input.EchoCharacter = '*'
		}
		form.inputs[i] = input
	}
	form.cursor = 0
	return form
}

func (f *formState) init() tea.Cmd {
	if len(f.inputs) == 0 {
		return nil
	}
	f.inputs[0].Focus()
	return nil
}

func loginForm() formConfig {
	return formConfig{
		title:  "Вход",
		action: actionLogin,
		fields: []formField{
			{label: "Логин", placeholder: "user"},
			{label: "Пароль", placeholder: "password", echo: textinput.EchoPassword},
		},
	}
}

func registerForm() formConfig {
	return formConfig{
		title:  "Регистрация",
		action: actionRegister,
		fields: []formField{
			{label: "Логин", placeholder: "user"},
			{label: "Пароль", placeholder: "password", echo: textinput.EchoPassword},
		},
	}
}

func readForm() formConfig {
	return formConfig{
		title:  "Чтение записи",
		action: actionRead,
		fields: []formField{
			{label: "Имя записи", placeholder: "unit_name"},
			{label: "Целевой файл", placeholder: "(опционально)"},
		},
	}
}

func deleteForm() formConfig {
	return formConfig{
		title:  "Удаление записи",
		action: actionDelete,
		fields: []formField{
			{label: "Имя записи", placeholder: "unit_name"},
		},
	}
}

func writeLoginForm() formConfig {
	return formConfig{
		title:  "Создание логина/пароля",
		action: actionWriteLogin,
		fields: []formField{
			{label: "Имя записи", placeholder: "unit_name"},
			{label: "Логин", placeholder: "login"},
			{label: "Пароль", placeholder: "password", echo: textinput.EchoPassword},
		},
	}
}

func writeTextForm() formConfig {
	return formConfig{
		title:  "Создание текстовой заметки",
		action: actionWriteText,
		fields: []formField{
			{label: "Имя записи", placeholder: "unit_name"},
			{label: "Текст", placeholder: "введите текст"},
		},
	}
}

func writeBinaryForm() formConfig {
	return formConfig{
		title:  "Создание бинарной записи",
		action: actionWriteBinary,
		fields: []formField{
			{label: "Имя записи", placeholder: "unit_name"},
			{label: "Путь к файлу", placeholder: "/path/to/file"},
		},
	}
}

func writeCardForm() formConfig {
	return formConfig{
		title:  "Создание банковской карты",
		action: actionWriteCard,
		fields: []formField{
			{label: "Имя записи", placeholder: "unit_name"},
			{label: "Номер", placeholder: "0000-0000-0000-0000"},
			{label: "Срок (YYYYMM)", placeholder: "203012"},
			{label: "Имя", placeholder: "IMYA"},
			{label: "Фамилия", placeholder: "FAMILIYA"},
			{label: "CVV", placeholder: "000"},
		},
	}
}
