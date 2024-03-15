package client

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	viewport viewport.Model
	textarea textarea.Model
	messages []string
	wsclient *client
}

type chatMsg struct {
	msg  string
	from string
}

type errMsg error

var (
    borderStyle = lipgloss.NewStyle().
        BorderStyle(lipgloss.RoundedBorder()).Padding(0, 1)
)

// --------------------------
//   Init
// --------------------------
func NewModel(wsclient *client) model {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()

	ta.Prompt = ""
	ta.CharLimit = 280

	ta.SetWidth(40)
	ta.SetHeight(3)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	vp := viewport.New(40, 20)
	vp.SetContent(`Welcome to the chat room!
Type a message and press Enter to send.
Type ':q' to exit`)

	ta.KeyMap.InsertNewline.SetEnabled(false)

	return model{
		textarea: ta,
		viewport: vp,
		messages: []string{},
		wsclient: wsclient,
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

// --------------------------
//   Update
// --------------------------
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {

	// key pressed event
	case tea.KeyMsg:
		switch msg.Type {

		case tea.KeyCtrlC, tea.KeyEsc:
			return m.exit()

		case tea.KeyEnter:
			// when user presses enter

			input := strings.TrimSpace(m.textarea.Value())
			if input == ":q" {
				// exiting...
				return m.exit()
			}

			// send the message to the server
			cmds = append(
				cmds,
				func() tea.Msg {
					err := m.wsclient.Send(input)
					if err != nil {
						return errMsg(err)
					}
					return nil
				},
			)

			yourMsg := "You:\n\t" + input
			m.messages = append(m.messages, yourMsg)
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.textarea.Reset()
			m.viewport.GotoBottom()
		}
	case errMsg:
		log.Println("ui: error message received \n", msg)
		return m.exit()

	case chatMsg:
		log.Printf("ui: message received %s from %s\n", msg.msg, msg.from)

		otherMsg := fmt.Sprintf("%s:\n\t%s", msg.from, msg.msg)
		m.messages = append(m.messages, otherMsg)
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()
	}

	return m, tea.Batch(cmds...)
}

func (m model) exit() (tea.Model, tea.Cmd) {
	m.viewport.SetContent("Exiting chat room...")
	m.wsclient.Close()
	log.Println("ui: exiting program")
	return m, tea.Quit
}

// --------------------------
//   View
// --------------------------
func (m model) View() string {
	return fmt.Sprintf(
		"\n%s\n%s",
		m.viewportView(),
		m.textareaView(),
	) + "\n\n"
}

func (m model) textareaView() string {
	return borderStyle.Render(m.textarea.View())
}

func (m model) viewportView() string {
	return borderStyle.Render(m.viewport.View())
}

// --------------------------
//   Util
// --------------------------
func HandleOnRead(p *tea.Program) OnRead {
	return func(msg interface{}) {
		p.Send(msg)
	}
}
