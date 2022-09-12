package task

import (
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	runewidth "github.com/mattn/go-runewidth"
	"github.com/muesli/reflow/ansi"
	"github.com/muesli/reflow/truncate"
)

const (
	statusBarHeight      = 1
	noteCharacterLimit   = 256             // should match server
	statusMessageTimeout = time.Second * 2 // how long to show status messages like "stashed!"
	pagerStashIcon       = "🔒"
	ellipsis             = "…"
)

var (
	pagerHelpHeight int

	mintGreen = lipgloss.AdaptiveColor{Light: "#89F0CB", Dark: "#89F0CB"}
	darkGreen = lipgloss.AdaptiveColor{Light: "#1C8760", Dark: "#1C8760"}

	noteHeading = lipgloss.NewStyle().
			Foreground(cream).
			Background(green).
			Padding(0, 1).
			Render("Set Memo")

	statusBarNoteFg = lipgloss.AdaptiveColor{Light: "#656565", Dark: "#7D7D7D"}
	statusBarBg     = lipgloss.AdaptiveColor{Light: "#E6E6E6", Dark: "#242424"}

	statusBarScrollPosStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#949494", Dark: "#5A5A5A"}).
				Background(statusBarBg).
				Render

	statusBarNoteStyle = lipgloss.NewStyle().
				Foreground(statusBarNoteFg).
				Background(statusBarBg).
				Render

	statusBarHelpStyle = lipgloss.NewStyle().
				Foreground(statusBarNoteFg).
				Background(lipgloss.AdaptiveColor{Light: "#DCDCDC", Dark: "#323232"}).
				Render

	statusBarStashDotStyle = lipgloss.NewStyle().
				Foreground(green).
				Background(statusBarBg).
				Render

	statusBarMessageStyle = lipgloss.NewStyle().
				Foreground(mintGreen).
				Background(darkGreen).
				Render

	statusBarMessageStashIconStyle = lipgloss.NewStyle().
					Foreground(mintGreen).
					Background(darkGreen).
					Render

	statusBarMessageScrollPosStyle = lipgloss.NewStyle().
					Foreground(mintGreen).
					Background(darkGreen).
					Render

	statusBarMessageHelpStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#B6FFE4")).
					Background(green).
					Render

	helpViewStyle = lipgloss.NewStyle().
			Foreground(statusBarNoteFg).
			Background(lipgloss.AdaptiveColor{Light: "#f2f2f2", Dark: "#1B1B1B"}).
			Render

	pagerNoteInputPromptStyle = lipgloss.NewStyle().
					Foreground(darkGray).
					Background(yellowGreen).
					Padding(0, 1)

	pagerNoteInputStyle = lipgloss.NewStyle().
				Foreground(darkGray).
				Background(yellowGreen)

	pagerNoteInputCursorStyle = lipgloss.NewStyle().
					Foreground(fuschia)

	logoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ECFD65")).
			Background(fuschia).
			Bold(true)
)

type contentRenderedMsg string

type pagerState int

const (
	pagerStateBrowse pagerState = iota
	pagerStateSetNote
	pagerStateStashing
	pagerStateStashSuccess
	pagerStateStatusMessage
)

type pagerModel struct {
	common    *commonModel
	viewport  viewport.Model
	state     pagerState
	showHelp  bool
	textInput textinput.Model
	spinner   spinner.Model

	statusMessage      string
	statusMessageTimer *time.Timer

	// Current document being rendered, sans-glamour rendering. We cache
	// it here so we can re-render it on resize.
	currentDocument markdown
}

func newPagerModel(common *commonModel) pagerModel {
	// Init viewport
	vp := viewport.New(0, 0)
	vp.YPosition = 0
	vp.HighPerformanceRendering = common.cfg.HighPerformancePager

	// Text input for notes/memos
	ti := textinput.New()
	ti.Prompt = " > "
	ti.PromptStyle = pagerNoteInputPromptStyle
	ti.TextStyle = pagerNoteInputStyle
	ti.CursorStyle = pagerNoteInputCursorStyle
	ti.CharLimit = noteCharacterLimit
	ti.Focus()

	// Text input for search
	// sp := spinner.New()
	// sp.Style = spinnerStyle

	return pagerModel{
		common:    common,
		state:     pagerStateBrowse,
		textInput: ti,
		viewport:  vp,
		// spinner:   sp,
	}
}

func (m *pagerModel) setSize(w, h int) {
	m.viewport.Width = w
	m.viewport.Height = h - statusBarHeight
	m.textInput.Width = w -
		ansi.PrintableRuneWidth(noteHeading) -
		ansi.PrintableRuneWidth(m.textInput.Prompt) - 1

	if m.showHelp {
		if pagerHelpHeight == 0 {
			pagerHelpHeight = strings.Count(m.helpView(), "\n")
		}
		m.viewport.Height -= (statusBarHeight + pagerHelpHeight)
	}
}

func (m *pagerModel) setContent(s string) {
	m.viewport.SetContent(s)
}

func (m *pagerModel) toggleHelp() {
	m.showHelp = !m.showHelp
	// TODO: setting size
	m.setSize(m.common.width, m.common.height)
	if m.viewport.PastBottom() {
		m.viewport.GotoBottom()
	}
}

// Perform stuff that needs to happen after a successful markdown stash. Note
// that the the returned command should be sent back the through the pager
// update function.
func (m *pagerModel) showStatusMessage(statusMessage string) tea.Cmd {
	// Show a success message to the user
	m.state = pagerStateStatusMessage
	m.statusMessage = statusMessage
	if m.statusMessageTimer != nil {
		m.statusMessageTimer.Stop()
	}
	m.statusMessageTimer = time.NewTimer(statusMessageTimeout)

	return waitForStatusMessageTimeout(pagerContext, m.statusMessageTimer)
}

func (m pagerModel) Update(msg tea.Msg) (pagerModel, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case pagerStateSetNote:
			switch msg.String() {
			case "esc":
				m.state = pagerStateBrowse
				return m, nil
			case "enter":
				var cmd tea.Cmd
				if m.textInput.Value() != m.currentDocument.Note { // don't update if the note didn't change
					m.currentDocument.Note = m.textInput.Value() // update optimistically
					// cmd = saveDocumentNote(m.common.cc, m.currentDocument.ID, m.currentDocument.Note)
				}
				m.state = pagerStateBrowse
				m.textInput.Reset()
				return m, cmd
			}
		default:
			switch msg.String() {
			case "q", "esc":
				if m.state != pagerStateBrowse {
					m.state = pagerStateBrowse
					return m, nil
				}
			case "home", "g":
				m.viewport.GotoTop()
				if m.viewport.HighPerformanceRendering {
					cmds = append(cmds, viewport.Sync(m.viewport))
				}
			case "end", "G":
				m.viewport.GotoBottom()
				if m.viewport.HighPerformanceRendering {
					cmds = append(cmds, viewport.Sync(m.viewport))
				}
			case "m":
				isStashed := m.currentDocument.docType == StashedDoc ||
					m.currentDocument.docType == ConvertedDoc

				// Users can only set the note on user-stashed markdown
				if !isStashed {
					break
				}

				m.state = pagerStateSetNote

				// Stop the timer for hiding a status message since changing
				// the state above will have cleared it.
				if m.statusMessageTimer != nil {
					m.statusMessageTimer.Stop()
				}

				// Pre-populate note with existing value
				if m.textInput.Value() == "" {
					m.textInput.SetValue(m.currentDocument.Note)
					m.textInput.CursorEnd()
				}

				return m, textinput.Blink
			case "?":
				m.toggleHelp()
				if m.viewport.HighPerformanceRendering {
					cmds = append(cmds, viewport.Sync(m.viewport))
				}
			}
		}

	case spinner.TickMsg:
		if m.state == pagerStateStashing {
			// If we're still stashing, or if the spinner still needs to
			// finish, spin it along.
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		} else if m.state == pagerStateStashSuccess {
			// NOTE: we shouldn't reach here
			log.Println("we shouldn't reach here")
			// If the spinner's finished and we haven't told the user the
			// stash was successful, do that.
			m.state = pagerStateBrowse
			// m.currentDocument = *m.stashedDocument
			// m.stashedDocument = nil
			cmds = append(cmds, m.showStatusMessage("Stashed!"))
		}

	// Glow has rendered the content
	case contentRenderedMsg:
		m.setContent(string(msg))
		if m.viewport.HighPerformanceRendering {
			cmds = append(cmds, viewport.Sync(m.viewport))
		}

	// We've reveived terminal dimensions, either for the first time or
	// after a resize
	case tea.WindowSizeMsg:
		return m, renderWithGlamour(m, m.currentDocument.Body)

	// case stashSuccessMsg:
	// 	// Stashing was successful. Convert the loaded document to a stashed
	// 	// one and show a status message. Note that we're also handling this
	// 	// message in the main update function where we're adding this stashed
	// 	// item to the stash listing.
	// 	m.state = pagerStateStashSuccess
	// 	if !m.spinner.Visible() {
	// 		// The spinner has finished spinning, so tell the user the stash
	// 		// was successful.
	// 		m.state = pagerStateBrowse
	// 		m.currentDocument = markdown(msg)
	// 		cmds = append(cmds, m.showStatusMessage("Stashed!"))
	// 	} else {
	// 		// The spinner is still spinning, so just take note of the newly
	// 		// stashed document for now.
	// 		md := markdown(msg)
	// 		m.stashedDocument = &md
	// 	}
	//
	// case stashFailMsg:
	// 	delete(m.common.filesStashed, msg.markdown.stashID)

	case statusMessageTimeoutMsg:
		m.state = pagerStateBrowse
	}

	switch m.state {
	case pagerStateSetNote:
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)
	default:
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m pagerModel) View() string {
	var b strings.Builder
	fmt.Fprint(&b, m.viewport.View()+"\n")

	// Footer
	switch m.state {
	case pagerStateSetNote:
		m.setNoteView(&b)
	default:
		m.statusBarView(&b)
	}

	if m.showHelp {
		fmt.Fprint(&b, "\n"+m.helpView())
	}

	return b.String()
}

func glowLogoView(text string) string {
	return logoStyle.Render(text)
}

func (m pagerModel) statusBarView(b *strings.Builder) {
	const (
		minPercent               float64 = 0.0
		maxPercent               float64 = 1.0
		percentToStringMagnitude float64 = 100.0
	)
	var (
		isStashed         bool = m.currentDocument.docType == StashedDoc || m.currentDocument.docType == ConvertedDoc
		showStatusMessage bool = m.state == pagerStateStatusMessage
	)

	// Logo
	logo := glowLogoView(" Glow ")

	// Scroll percent
	percent := math.Max(minPercent, math.Min(maxPercent, m.viewport.ScrollPercent()))
	scrollPercent := fmt.Sprintf(" %3.f%% ", percent*percentToStringMagnitude)
	if showStatusMessage {
		scrollPercent = statusBarMessageScrollPosStyle(scrollPercent)
	} else {
		scrollPercent = statusBarScrollPosStyle(scrollPercent)
	}

	// "Help" note
	var helpNote string
	if showStatusMessage {
		helpNote = statusBarMessageHelpStyle(" ? Help ")
	} else {
		helpNote = statusBarHelpStyle(" ? Help ")
	}

	// Status indicator; spinner or stash dot
	var statusIndicator string
	if m.state == pagerStateStashing || m.state == pagerStateStashSuccess {
		statusIndicator = statusBarNoteStyle(" ") + m.spinner.View()
	} else if isStashed && showStatusMessage {
		statusIndicator = statusBarMessageStashIconStyle(" " + pagerStashIcon)
	} else if isStashed {
		statusIndicator = statusBarStashDotStyle(" " + pagerStashIcon)
	}

	// Note
	var note string
	if showStatusMessage {
		note = m.statusMessage
	} else {
		note = m.currentDocument.Note
		if len(note) == 0 {
			note = "(No memo)"
		}
	}
	note = truncate.StringWithTail(" "+note+" ", uint(max(0,
		m.common.width-
			ansi.PrintableRuneWidth(logo)-
			ansi.PrintableRuneWidth(statusIndicator)-
			ansi.PrintableRuneWidth(scrollPercent)-
			ansi.PrintableRuneWidth(helpNote),
	)), ellipsis)
	if showStatusMessage {
		note = statusBarMessageStyle(note)
	} else {
		note = statusBarNoteStyle(note)
	}

	// Empty space
	padding := max(0,
		m.common.width-
			ansi.PrintableRuneWidth(logo)-
			ansi.PrintableRuneWidth(statusIndicator)-
			ansi.PrintableRuneWidth(note)-
			ansi.PrintableRuneWidth(scrollPercent)-
			ansi.PrintableRuneWidth(helpNote),
	)
	emptySpace := strings.Repeat(" ", padding)
	if showStatusMessage {
		emptySpace = statusBarMessageStyle(emptySpace)
	} else {
		emptySpace = statusBarNoteStyle(emptySpace)
	}

	fmt.Fprintf(b, "%s%s%s%s%s%s",
		logo,
		statusIndicator,
		note,
		emptySpace,
		scrollPercent,
		helpNote,
	)
}

func (m pagerModel) setNoteView(b *strings.Builder) {
	fmt.Fprint(b, noteHeading)
	fmt.Fprint(b, m.textInput.View())
}

func (m pagerModel) helpView() (s string) {
	memoOrStash := "m       set memo"
	// NOTE: no stashing
	// if m.common.authStatus == authOK && m.currentDocument.docType == LocalDoc {
	// 	memoOrStash = "s       stash this document"
	// }

	col1 := []string{
		"g/home  go to top",
		"G/end   go to bottom",
		"",
		memoOrStash,
		"esc     back to files",
		"q       quit",
	}

	if m.currentDocument.docType == NewsDoc {
		deleteFromStringSlice(col1, 3)
	}

	s += "\n"
	s += "k/↑      up                  " + col1[0] + "\n"
	s += "j/↓      down                " + col1[1] + "\n"
	s += "b/pgup   page up             " + col1[2] + "\n"
	s += "f/pgdn   page down           " + col1[3] + "\n"
	s += "u        ½ page up           " + col1[4] + "\n"
	s += "d        ½ page down         "

	if len(col1) > 5 {
		s += col1[5]
	}

	s = indent(s, 2)

	// Fill up empty cells with spaces for background coloring
	if m.common.width > 0 {
		lines := strings.Split(s, "\n")
		for i := 0; i < len(lines); i++ {
			l := runewidth.StringWidth(lines[i])
			n := max(m.common.width-l, 0)
			lines[i] += strings.Repeat(" ", n)
		}

		s = strings.Join(lines, "\n")
	}

	return helpViewStyle(s)
}

// COMMANDS

func renderWithGlamour(m pagerModel, md string) tea.Cmd {
	return func() tea.Msg {
		s, err := glamourRender(m, md)
		if err != nil {
			// if debug {
			log.Printf("error rendering with Glamour: %v\n", err)
			log.Println("errored text is ", s)
			// }
			return err
		}
		log.Printf("rendered the following text with Glamour: %s\n", s)
		return contentRenderedMsg(s)
	}
}

// This is where the magic happens.
func glamourRender(m pagerModel, markdown string) (string, error) {
	// TODO: global config (not m.cfg)
	// if !m.common.cfg.GlamourEnabled {
	// 	return markdown, nil
	// }

	log.Println("initialize glamour")
	// initialize glamour
	var gs glamour.TermRendererOption
	if m.common.cfg.GlamourStyle == "auto" {
		gs = glamour.WithAutoStyle()
	} else {
		gs = glamour.WithStylePath(m.common.cfg.GlamourStyle)
	}

	width := max(0, min(int(m.common.cfg.GlamourMaxWidth), m.viewport.Width))
	r, err := glamour.NewTermRenderer(
		gs,
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return "", err
	}

	out, err := r.Render(markdown)
	if err != nil {
		return "", err
	}

	// trim lines
	lines := strings.Split(out, "\n")

	var content string
	for i, s := range lines {
		content += strings.TrimSpace(s)

		// don't add an artificial newline after the last split
		if i+1 < len(lines) {
			content += "\n"
		}
	}

	return content, nil
}

// ETC
func waitForStatusMessageTimeout(appCtx applicationContext, t *time.Timer) tea.Cmd {
	return func() tea.Msg {
		<-t.C
		return statusMessageTimeoutMsg(appCtx)
	}
}

// Note: this runs in linear time; O(n).
func deleteFromStringSlice(a []string, i int) []string {
	copy(a[i:], a[i+1:])
	a[len(a)-1] = ""
	return a[:len(a)-1]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func indent(s string, n int) string {
	if n <= 0 || s == "" {
		return s
	}
	l := strings.Split(s, "\n")
	b := strings.Builder{}
	i := strings.Repeat(" ", n)
	for _, v := range l {
		fmt.Fprintf(&b, "%s%s\n", i, v)
	}
	return b.String()
}

func errorView(err error, fatal bool) string {
	exitMsg := "press any key to "
	if fatal {
		exitMsg += "exit"
	} else {
		exitMsg += "return"
	}
	s := fmt.Sprintf("%s\n\n%v\n\n%s",
		errorTitleStyle.Render("ERROR"),
		err,
		subtleStyle.Render(exitMsg),
	)
	return "\n" + indent(s, 3)
}