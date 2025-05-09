package tui

import (
	"fmt"
	"os"
	"strconv"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ScanMode represents different scanning modes
type ScanMode int

const (
	SinglePackageMode ScanMode = iota
	DirectoryScanMode
)

// AppConfig holds the configuration captured via the TUI
type AppConfig struct {
	// Scanning mode
	Mode ScanMode

	// Single package scan options
	PackageName      string
	PackageVersion   string
	PackageEcosystem string

	// Directory scan options
	DirectoryPath string
	FileExtension string
	Concurrency   int

	// Database options
	UseDB      bool
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// Logging options
	LogToFile     bool
	LogFilePath   string
	LogMaxSize    int
	LogMaxBackups int
	LogMaxAge     int
	LogCompress   bool
	LogLevel      string
	LogFormat     string
}

// keyMap defines the keybindings for the application
type keyMap struct {
	Next          key.Binding
	Prev          key.Binding
	Submit        key.Binding
	ToggleMode    key.Binding
	ToggleOptions key.Binding
	Quit          key.Binding
}

// defaultKeyMap returns the default keybindings
func defaultKeyMap() keyMap {
	return keyMap{
		Next: key.NewBinding(
			key.WithKeys("tab", "down", "ctrl+n"),
			key.WithHelp("tab/↓", "next field"),
		),
		Prev: key.NewBinding(
			key.WithKeys("shift+tab", "up", "ctrl+p"),
			key.WithHelp("shift+tab/↑", "previous field"),
		),
		Submit: key.NewBinding(
			key.WithKeys("enter", "ctrl+s"),
			key.WithHelp("enter", "submit"),
		),
		ToggleMode: key.NewBinding(
			key.WithKeys("ctrl+t"),
			key.WithHelp("ctrl+t", "toggle scan mode"),
		),
		ToggleOptions: key.NewBinding(
			key.WithKeys("ctrl+o"),
			key.WithHelp("ctrl+o", "toggle advanced options"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c", "esc"),
			key.WithHelp("ctrl+c/esc", "quit"),
		),
	}
}

// Model represents the state of the TUI application
type Model struct {
	config          AppConfig
	keys            keyMap
	help            help.Model
	showAdvanced    bool
	activeInput     int
	inputs          []textinput.Model
	dbInputs        []textinput.Model
	logInputs       []textinput.Model
	width           int
	height          int
	ready           bool
	err             error
	quitting        bool
	inputsSubmitted bool
}

// ShortHelp returns keybindings to be shown in the mini help view.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Next, k.Prev, k.Submit, k.ToggleMode, k.ToggleOptions, k.Quit}
}

// FullHelp returns keybindings for the expanded help view.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Next, k.Prev},
		{k.Submit, k.ToggleMode, k.ToggleOptions},
		{k.Quit},
	}
}

var (
	titleStyle         = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FAFAFA")).Background(lipgloss.Color("#7D56F4")).Padding(0, 2)
	activeInputStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4")).BorderForeground(lipgloss.Color("#7D56F4"))
	inactiveInputStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).BorderForeground(lipgloss.Color("241"))
	statusMessageStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575")).Bold(true)
	errorMessageStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Bold(true)
	helpStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	sectionStyle       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FAFAFA")).Background(lipgloss.Color("#3B3B98")).Padding(0, 1)
	labelStyle         = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FAFAFA"))
)

// NewModel creates a new TUI model with default values
func NewModel() Model {
	// Read environment variables for default settings
	logFormat := os.Getenv("LOG_FORMAT")
	if logFormat == "" {
		logFormat = "json" // Default to JSON if not specified
	}

	// Read other environment variables with defaults
	logLevel := getEnvWithDefault("LOG_LEVEL", "info")
	logFilePath := getEnvWithDefault("LOG_FILE_PATH", "logs/package-scanner.log")
	logMaxSize, _ := strconv.Atoi(getEnvWithDefault("LOG_MAX_SIZE", "10"))
	logMaxBackups, _ := strconv.Atoi(getEnvWithDefault("LOG_MAX_BACKUPS", "5"))
	logMaxAge, _ := strconv.Atoi(getEnvWithDefault("LOG_MAX_AGE", "30"))

	// Database defaults
	dbHost := getEnvWithDefault("DB_HOST", "localhost")
	dbPortStr := getEnvWithDefault("DB_PORT", "5432")
	dbPort, _ := strconv.Atoi(dbPortStr)
	dbUser := getEnvWithDefault("DB_USER", "postgres")
	dbPassword := getEnvWithDefault("DB_PASSWORD", "")
	dbName := getEnvWithDefault("DB_NAME", "package_scanner")
	dbSSLMode := getEnvWithDefault("DB_SSL_MODE", "disable")

	// Other defaults
	packageName := getEnvWithDefault("PACKAGE_NAME", "Microsoft.AspNetCore.Identity")
	packageVersion := getEnvWithDefault("PACKAGE_VERSION", "2.3.0")
	packageEcosystem := getEnvWithDefault("PACKAGE_ECOSYSTEM", "NuGet")
	directoryPath := getEnvWithDefault("DIRECTORY_PATH", "./packages")
	fileExtension := getEnvWithDefault("FILE_EXTENSION", "nupkg")
	concurrency, _ := strconv.Atoi(getEnvWithDefault("CONCURRENCY", "5"))

	m := Model{
		config: AppConfig{
			Mode:             SinglePackageMode,
			PackageName:      packageName,
			PackageVersion:   packageVersion,
			PackageEcosystem: packageEcosystem,
			DirectoryPath:    directoryPath,
			FileExtension:    fileExtension,
			Concurrency:      concurrency,
			UseDB:            false,
			DBHost:           dbHost,
			DBPort:           dbPort,
			DBUser:           dbUser,
			DBPassword:       dbPassword,
			DBName:           dbName,
			DBSSLMode:        dbSSLMode,
			LogToFile:        true,
			LogFilePath:      logFilePath,
			LogMaxSize:       logMaxSize,
			LogMaxBackups:    logMaxBackups,
			LogMaxAge:        logMaxAge,
			LogCompress:      true,
			LogLevel:         logLevel,
			LogFormat:        logFormat,
		},
		keys:         defaultKeyMap(),
		help:         help.New(),
		showAdvanced: false,
		activeInput:  0,
	}

	// Initialize inputs with default values
	m.initializeInputs()

	return m
}

// getEnvWithDefault reads an environment variable and returns a default value if not set
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// RunTUI starts the TUI application
func RunTUI() (*AppConfig, error) {
	p := tea.NewProgram(NewModel())
	model, err := p.Run()
	if err != nil {
		return nil, err
	}

	finalModel, ok := model.(Model)
	if !ok {
		return nil, nil
	}

	if finalModel.inputsSubmitted {
		return &finalModel.config, nil
	}

	return nil, nil
}

// initializeInputs sets up all the text inputs with their initial values
func (m *Model) initializeInputs() {
	// Single package scan inputs
	singlePackageInputs := []struct {
		placeholder string
		value       string
		label       string
	}{
		{placeholder: "Package Name", value: m.config.PackageName, label: "Package Name"},
		{placeholder: "Version", value: m.config.PackageVersion, label: "Version"},
		{placeholder: "Ecosystem (npm, NuGet, PyPI, etc.)", value: m.config.PackageEcosystem, label: "Ecosystem"},
	}

	// Directory scan inputs
	dirScanInputs := []struct {
		placeholder string
		value       string
		label       string
	}{
		{placeholder: "Directory Path", value: m.config.DirectoryPath, label: "Directory Path"},
		{placeholder: "File Extension (nupkg, tgz, etc.)", value: m.config.FileExtension, label: "File Extension"},
		{placeholder: "Concurrency (1-20)", value: fmt.Sprintf("%d", m.config.Concurrency), label: "Concurrency"},
	}

	// Database inputs
	dbInputs := []struct {
		placeholder string
		value       string
		label       string
	}{
		{placeholder: "Host", value: m.config.DBHost, label: "Database Host"},
		{placeholder: "Port", value: fmt.Sprintf("%d", m.config.DBPort), label: "Database Port"},
		{placeholder: "User", value: m.config.DBUser, label: "Database User"},
		{placeholder: "Password", value: m.config.DBPassword, label: "Database Password"},
		{placeholder: "Database Name", value: m.config.DBName, label: "Database Name"},
		{placeholder: "SSL Mode", value: m.config.DBSSLMode, label: "SSL Mode"},
	}

	// Log inputs
	logInputs := []struct {
		placeholder string
		value       string
		label       string
	}{
		{placeholder: "Log File Path", value: m.config.LogFilePath, label: "Log File Path"},
		{placeholder: "Max Size (MB)", value: fmt.Sprintf("%d", m.config.LogMaxSize), label: "Max Size"},
		{placeholder: "Max Backups", value: fmt.Sprintf("%d", m.config.LogMaxBackups), label: "Max Backups"},
		{placeholder: "Max Age (days)", value: fmt.Sprintf("%d", m.config.LogMaxAge), label: "Max Age"},
		{placeholder: "Log Level (debug, info, warn, error)", value: m.config.LogLevel, label: "Log Level"},
		{placeholder: "Log Format (json, text)", value: m.config.LogFormat, label: "Log Format"},
	}

	// Initialize single package inputs
	m.inputs = make([]textinput.Model, 0)

	// Set up inputs based on the current mode
	var inputs []struct{ placeholder, value, label string }
	if m.config.Mode == SinglePackageMode {
		inputs = singlePackageInputs
	} else {
		inputs = dirScanInputs
	}

	// Create and style text inputs
	for i, inp := range inputs {
		ti := textinput.New()
		ti.Placeholder = inp.placeholder
		ti.SetValue(inp.value)
		ti.CharLimit = 64
		ti.Width = 40

		// Set initial focus
		if i == 0 {
			ti.Focus()
			ti.PromptStyle = activeInputStyle
			ti.TextStyle = activeInputStyle
		}

		m.inputs = append(m.inputs, ti)
	}

	// Initialize database inputs
	m.dbInputs = make([]textinput.Model, len(dbInputs))
	for i, inp := range dbInputs {
		ti := textinput.New()
		ti.Placeholder = inp.placeholder
		ti.SetValue(inp.value)
		ti.CharLimit = 64
		ti.Width = 40

		// Hide password input
		if inp.label == "Database Password" {
			ti.EchoMode = textinput.EchoPassword
			ti.EchoCharacter = '•'
		}

		m.dbInputs[i] = ti
	}

	// Initialize logging inputs
	m.logInputs = make([]textinput.Model, len(logInputs))
	for i, inp := range logInputs {
		ti := textinput.New()
		ti.Placeholder = inp.placeholder
		ti.SetValue(inp.value)
		ti.CharLimit = 64
		ti.Width = 40

		m.logInputs[i] = ti
	}
}

// Init initializes the Bubble Tea model
func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles events and updates the model accordingly
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, m.keys.Submit):
			// Update config with current values
			m.updateConfig()
			m.inputsSubmitted = true
			return m, tea.Quit

		case key.Matches(msg, m.keys.ToggleMode):
			// Toggle between single package and directory scan modes
			if m.config.Mode == SinglePackageMode {
				m.config.Mode = DirectoryScanMode
			} else {
				m.config.Mode = SinglePackageMode
			}
			// Reinitialize inputs for the new mode
			m.activeInput = 0
			m.initializeInputs()

		case key.Matches(msg, m.keys.ToggleOptions):
			// Toggle advanced options visibility
			m.showAdvanced = !m.showAdvanced

		case key.Matches(msg, m.keys.Next):
			m.nextInput()

		case key.Matches(msg, m.keys.Prev):
			m.prevInput()
		}

	case tea.WindowSizeMsg:
		// Set the window size
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width

		if !m.ready {
			// This is the first time we're receiving a WindowSizeMsg, so we're ready to render
			m.ready = true
		}
	}

	// Handle active input updates
	if m.activeInput >= 0 && m.activeInput < len(m.inputs) {
		// Update active input in main inputs
		newInput, cmd := m.inputs[m.activeInput].Update(msg)
		m.inputs[m.activeInput] = newInput
		cmds = append(cmds, cmd)
	} else if m.showAdvanced {
		// If we're showing advanced options and beyond main inputs
		dbInputCount := len(m.dbInputs)
		if m.activeInput >= len(m.inputs) && m.activeInput < len(m.inputs)+dbInputCount {
			// Update active input in DB inputs
			dbIdx := m.activeInput - len(m.inputs)
			newInput, cmd := m.dbInputs[dbIdx].Update(msg)
			m.dbInputs[dbIdx] = newInput
			cmds = append(cmds, cmd)
		} else if m.activeInput >= len(m.inputs)+dbInputCount {
			// Update active input in logging inputs
			logIdx := m.activeInput - len(m.inputs) - dbInputCount
			if logIdx < len(m.logInputs) {
				newInput, cmd := m.logInputs[logIdx].Update(msg)
				m.logInputs[logIdx] = newInput
				cmds = append(cmds, cmd)
			}
		}
	}

	return m, tea.Batch(cmds...)
}

// nextInput focuses the next input field
func (m *Model) nextInput() {
	totalInputs := len(m.inputs)
	if m.showAdvanced {
		totalInputs += len(m.dbInputs) + len(m.logInputs)
	}

	// Reset the currently active input
	if m.activeInput >= 0 && m.activeInput < len(m.inputs) {
		m.inputs[m.activeInput].Blur()
		m.inputs[m.activeInput].PromptStyle = inactiveInputStyle
		m.inputs[m.activeInput].TextStyle = inactiveInputStyle
	} else if m.showAdvanced {
		dbInputCount := len(m.dbInputs)
		if m.activeInput >= len(m.inputs) && m.activeInput < len(m.inputs)+dbInputCount {
			// Blurring a DB input
			dbIdx := m.activeInput - len(m.inputs)
			m.dbInputs[dbIdx].Blur()
			m.dbInputs[dbIdx].PromptStyle = inactiveInputStyle
			m.dbInputs[dbIdx].TextStyle = inactiveInputStyle
		} else if m.activeInput >= len(m.inputs)+dbInputCount {
			// Blurring a log input
			logIdx := m.activeInput - len(m.inputs) - dbInputCount
			if logIdx < len(m.logInputs) {
				m.logInputs[logIdx].Blur()
				m.logInputs[logIdx].PromptStyle = inactiveInputStyle
				m.logInputs[logIdx].TextStyle = inactiveInputStyle
			}
		}
	}

	// Move to the next input
	m.activeInput = (m.activeInput + 1) % totalInputs

	// Focus the new active input
	if m.activeInput < len(m.inputs) {
		m.inputs[m.activeInput].Focus()
		m.inputs[m.activeInput].PromptStyle = activeInputStyle
		m.inputs[m.activeInput].TextStyle = activeInputStyle
	} else if m.showAdvanced {
		dbInputCount := len(m.dbInputs)
		if m.activeInput >= len(m.inputs) && m.activeInput < len(m.inputs)+dbInputCount {
			// Focusing a DB input
			dbIdx := m.activeInput - len(m.inputs)
			m.dbInputs[dbIdx].Focus()
			m.dbInputs[dbIdx].PromptStyle = activeInputStyle
			m.dbInputs[dbIdx].TextStyle = activeInputStyle
		} else {
			// Focusing a log input
			logIdx := m.activeInput - len(m.inputs) - dbInputCount
			if logIdx < len(m.logInputs) {
				m.logInputs[logIdx].Focus()
				m.logInputs[logIdx].PromptStyle = activeInputStyle
				m.logInputs[logIdx].TextStyle = activeInputStyle
			}
		}
	}
}

// prevInput focuses the previous input field
func (m *Model) prevInput() {
	totalInputs := len(m.inputs)
	if m.showAdvanced {
		totalInputs += len(m.dbInputs) + len(m.logInputs)
	}

	// Reset the currently active input
	if m.activeInput >= 0 && m.activeInput < len(m.inputs) {
		m.inputs[m.activeInput].Blur()
		m.inputs[m.activeInput].PromptStyle = inactiveInputStyle
		m.inputs[m.activeInput].TextStyle = inactiveInputStyle
	} else if m.showAdvanced {
		dbInputCount := len(m.dbInputs)
		if m.activeInput >= len(m.inputs) && m.activeInput < len(m.inputs)+dbInputCount {
			// Blurring a DB input
			dbIdx := m.activeInput - len(m.inputs)
			m.dbInputs[dbIdx].Blur()
			m.dbInputs[dbIdx].PromptStyle = inactiveInputStyle
			m.dbInputs[dbIdx].TextStyle = inactiveInputStyle
		} else if m.activeInput >= len(m.inputs)+dbInputCount {
			// Blurring a log input
			logIdx := m.activeInput - len(m.inputs) - dbInputCount
			if logIdx < len(m.logInputs) {
				m.logInputs[logIdx].Blur()
				m.logInputs[logIdx].PromptStyle = inactiveInputStyle
				m.logInputs[logIdx].TextStyle = inactiveInputStyle
			}
		}
	}

	// Move to the previous input
	m.activeInput = (m.activeInput - 1 + totalInputs) % totalInputs

	// Focus the new active input
	if m.activeInput < len(m.inputs) {
		m.inputs[m.activeInput].Focus()
		m.inputs[m.activeInput].PromptStyle = activeInputStyle
		m.inputs[m.activeInput].TextStyle = activeInputStyle
	} else if m.showAdvanced {
		dbInputCount := len(m.dbInputs)
		if m.activeInput >= len(m.inputs) && m.activeInput < len(m.inputs)+dbInputCount {
			// Focusing a DB input
			dbIdx := m.activeInput - len(m.inputs)
			m.dbInputs[dbIdx].Focus()
			m.dbInputs[dbIdx].PromptStyle = activeInputStyle
			m.dbInputs[dbIdx].TextStyle = activeInputStyle
		} else {
			// Focusing a log input
			logIdx := m.activeInput - len(m.inputs) - dbInputCount
			if logIdx < len(m.logInputs) {
				m.logInputs[logIdx].Focus()
				m.logInputs[logIdx].PromptStyle = activeInputStyle
				m.logInputs[logIdx].TextStyle = activeInputStyle
			}
		}
	}
}

// updateConfig updates the configuration with values from the input fields
func (m *Model) updateConfig() {
	// Update config from main inputs
	if m.config.Mode == SinglePackageMode {
		if len(m.inputs) >= 3 {
			m.config.PackageName = m.inputs[0].Value()
			m.config.PackageVersion = m.inputs[1].Value()
			m.config.PackageEcosystem = m.inputs[2].Value()
		}
	} else {
		if len(m.inputs) >= 3 {
			m.config.DirectoryPath = m.inputs[0].Value()
			m.config.FileExtension = m.inputs[1].Value()

			// Parse concurrency as int
			if concurrency, err := strconv.Atoi(m.inputs[2].Value()); err == nil {
				m.config.Concurrency = concurrency
			}
		}
	}

	// Update DB config
	if m.showAdvanced {
		if len(m.dbInputs) >= 6 {
			m.config.UseDB = true // If they're in advanced mode and submitting, assume DB usage
			m.config.DBHost = m.dbInputs[0].Value()
			if port, err := strconv.Atoi(m.dbInputs[1].Value()); err == nil {
				m.config.DBPort = port
			}
			m.config.DBUser = m.dbInputs[2].Value()
			m.config.DBPassword = m.dbInputs[3].Value()
			m.config.DBName = m.dbInputs[4].Value()
			m.config.DBSSLMode = m.dbInputs[5].Value()
		}

		// Update log config
		if len(m.logInputs) >= 6 {
			m.config.LogToFile = true
			m.config.LogFilePath = m.logInputs[0].Value()
			if size, err := strconv.Atoi(m.logInputs[1].Value()); err == nil {
				m.config.LogMaxSize = size
			}
			if backups, err := strconv.Atoi(m.logInputs[2].Value()); err == nil {
				m.config.LogMaxBackups = backups
			}
			if age, err := strconv.Atoi(m.logInputs[3].Value()); err == nil {
				m.config.LogMaxAge = age
			}
			m.config.LogLevel = m.logInputs[4].Value()
			m.config.LogFormat = m.logInputs[5].Value()
		}
	} else {
		// If not showing advanced options, disable DB usage
		m.config.UseDB = false
	}
}

// View renders the current UI state
func (m Model) View() string {
	if !m.ready {
		return "\nInitializing..."
	}

	if m.quitting {
		return "Goodbye!\n"
	}

	// Build the view
	var s string

	// Add header
	header := titleStyle.Render(" Package Scanner ")
	s += lipgloss.PlaceHorizontal(m.width, lipgloss.Center, header) + "\n\n"

	// Add mode indicator and toggle instruction
	modeText := "Single Package Mode"
	if m.config.Mode == DirectoryScanMode {
		modeText = "Directory Scan Mode"
	}
	modeToggle := fmt.Sprintf("Current: %s (Press Ctrl+T to toggle mode)", modeText)
	s += statusMessageStyle.Render(modeToggle) + "\n\n"

	// Add form inputs
	s += renderInputGroup(m)

	// Show advanced options if toggled
	if m.showAdvanced {
		// Database section
		s += "\n" + sectionStyle.Render(" Database Configuration ") + "\n\n"

		// Define labels for database fields
		dbLabels := []string{"Host:", "Port:", "User:", "Password:", "Database Name:", "SSL Mode:"}

		// Add database inputs with descriptive labels
		for i, input := range m.dbInputs {
			if i < len(dbLabels) {
				s += labelStyle.Render(fmt.Sprintf("%-15s", dbLabels[i])) + " " + input.View() + "\n"
			}
		}

		// Logging section
		s += "\n" + sectionStyle.Render(" Logging Configuration ") + "\n\n"

		// Define labels for logging fields
		logLabels := []string{"Log File Path:", "Max Size (MB):", "Max Backups:", "Max Age (days):", "Log Level:", "Log Format:"}

		// Add logging inputs with descriptive labels
		for i, input := range m.logInputs {
			if i < len(logLabels) {
				s += labelStyle.Render(fmt.Sprintf("%-15s", logLabels[i])) + " " + input.View() + "\n"
			}
		}
	} else {
		s += "\n" + statusMessageStyle.Render("Press Ctrl+O to show advanced options") + "\n"
	}

	// Help view
	helpView := m.help.View(m.keys)
	s += "\n" + helpStyle.Render(helpView)

	// Center everything
	return lipgloss.PlaceHorizontal(m.width, lipgloss.Center, s)
}

// renderInputGroup renders the appropriate input group based on the current mode
func renderInputGroup(m Model) string {
	var s string

	// Add section header based on mode
	if m.config.Mode == SinglePackageMode {
		s += sectionStyle.Render(" Package Details ") + "\n\n"

		// Define labels for single package mode
		labels := []string{"Package Name:", "Version:", "Ecosystem:"}

		// Add inputs with descriptive labels
		for i, input := range m.inputs {
			if i < len(labels) {
				s += labelStyle.Render(fmt.Sprintf("%-15s", labels[i])) + " " + input.View() + "\n"
			}
		}
	} else {
		s += sectionStyle.Render(" Directory Scan ") + "\n\n"

		// Define labels for directory scan mode
		labels := []string{"Directory Path:", "File Extension:", "Concurrency:"}

		// Add inputs with descriptive labels
		for i, input := range m.inputs {
			if i < len(labels) {
				s += labelStyle.Render(fmt.Sprintf("%-15s", labels[i])) + " " + input.View() + "\n"
			}
		}
	}

	return s
}
