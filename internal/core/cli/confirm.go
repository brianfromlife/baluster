package cli

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	environment      string
	confirmationCode string
	input            textinput.Model
	message          string
	err              error
	confirmed        bool
	cancelled        bool
}

func initialModel(env string, code string) model {
	ti := textinput.New()
	ti.Placeholder = code
	ti.Focus()
	ti.CharLimit = 50
	ti.Width = 50

	message := fmt.Sprintf("\nIf the above looks correct, please enter the code to deploy the infrastructure:\n\n%s", code)

	return model{
		environment:      env,
		confirmationCode: code,
		input:            ti,
		message:          message,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.cancelled = true
			return m, tea.Quit

		case tea.KeyEnter:
			value := strings.TrimSpace(m.input.Value())

			expected := m.confirmationCode
			if value == expected {
				m.confirmed = true
				return m, tea.Quit
			} else {
				m.err = fmt.Errorf("confirmation code does not match")
				m.input.SetValue("")
				return m, nil
			}
		}
	}

	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m model) View() string {
	var s strings.Builder

	s.WriteString("\n")
	s.WriteString(m.message)
	s.WriteString("\n\n")

	if m.err != nil {
		fmt.Fprintf(&s, "❌ %s\n\n", m.err.Error())
	} else {
		s.WriteString("→ ")
	}

	s.WriteString(m.input.View())
	s.WriteString("\n\n")
	s.WriteString("(press Ctrl+C or Esc to cancel)\n")

	return s.String()
}

// Run starts the confirmation prompt and returns true if confirmed, false if cancelled
// It always generates an 8-digit numeric code for confirmation
func Run(environment string) (bool, error) {
	code, err := generateNumericCode(8)
	if err != nil {
		return false, fmt.Errorf("failed to generate confirmation code: %w", err)
	}

	p := tea.NewProgram(initialModel(environment, code))
	m, err := p.Run()
	if err != nil {
		return false, err
	}

	model := m.(model)
	if model.cancelled {
		return false, nil
	}

	return model.confirmed, nil
}

// generateNumericCode generates a random numeric code of the specified length
func generateNumericCode(length int) (string, error) {
	code := make([]byte, length)
	for i := range code {
		num, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		code[i] = byte('0' + num.Int64())
	}
	return string(code), nil
}

// envPromptModel is the model for prompting for environment
type envPromptModel struct {
	input     textinput.Model
	err       error
	value     string
	cancelled bool
	done      bool
}

func initialEnvPromptModel() envPromptModel {
	ti := textinput.New()
	ti.Placeholder = "dev or prod"
	ti.Focus()
	ti.CharLimit = 10
	ti.Width = 50

	return envPromptModel{
		input: ti,
	}
}

func (m envPromptModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m envPromptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.cancelled = true
			return m, tea.Quit

		case tea.KeyEnter:
			value := strings.ToLower(strings.TrimSpace(m.input.Value()))
			if value == "" {
				m.err = fmt.Errorf("environment cannot be empty")
				return m, nil
			}
			if value != "dev" && value != "prod" {
				m.err = fmt.Errorf("environment must be either 'dev' or 'prod', got: %s", value)
				m.input.SetValue("")
				return m, nil
			}
			m.value = value
			m.done = true
			return m, tea.Quit
		}
	}

	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m envPromptModel) View() string {
	var s strings.Builder

	s.WriteString("\n")
	s.WriteString("Environment (dev/prod)\n\n")

	if m.err != nil {
		fmt.Fprintf(&s, "❌ %s\n\n", m.err.Error())
	} else {
		s.WriteString("→ ")
	}

	s.WriteString(m.input.View())
	s.WriteString("\n\n")
	s.WriteString("(press Enter to continue, Ctrl+C or Esc to cancel)\n")

	return s.String()
}

// PromptEnvironment prompts the user for the deployment environment (dev or prod)
func PromptEnvironment() (string, error) {
	p := tea.NewProgram(initialEnvPromptModel())
	m, err := p.Run()
	if err != nil {
		return "", err
	}

	model := m.(envPromptModel)
	if model.cancelled {
		return "", fmt.Errorf("cancelled")
	}

	if !model.done {
		return "", fmt.Errorf("prompt incomplete")
	}

	return model.value, nil
}
