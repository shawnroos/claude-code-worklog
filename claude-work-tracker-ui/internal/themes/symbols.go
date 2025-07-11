package themes

// Symbol sets for different visual styles
type SymbolSet struct {
	// Status symbols
	Completed   string
	InProgress  string
	Pending     string
	Failed      string

	// Priority symbols
	High        string
	Medium      string
	Low         string

	// Navigation symbols
	Right       string
	Left        string
	Up          string
	Down        string

	// UI elements
	Bullet      string
	Arrow       string
	Branch      string
	Reference   string
	Group       string
	Item        string

	// View icons
	Dashboard   string
	WorkItems   string
	References  string
	FutureWork  string
}

var (
	// Unicode symbols (default)
	UnicodeSymbols = SymbolSet{
		Completed:  "✓",
		InProgress: "○",
		Pending:    "●",
		Failed:     "✗",

		High:       "🔴",
		Medium:     "🟡",
		Low:        "🟢",

		Right:      "→",
		Left:       "←",
		Up:         "↑",
		Down:       "↓",

		Bullet:     "•",
		Arrow:      "➤",
		Branch:     "├",
		Reference:  "🔗",
		Group:      "📦",
		Item:       "📄",

		Dashboard:  "📊",
		WorkItems:  "📋",
		References: "🔗",
		FutureWork: "🚀",
	}

	// ASCII symbols (terminal compatibility)
	ASCIISymbols = SymbolSet{
		Completed:  "[x]",
		InProgress: "[o]",
		Pending:    "[ ]",
		Failed:     "[!]",

		High:       "[H]",
		Medium:     "[M]",
		Low:        "[L]",

		Right:      "->",
		Left:       "<-",
		Up:         "^",
		Down:       "v",

		Bullet:     "*",
		Arrow:      ">",
		Branch:     "|",
		Reference:  "@",
		Group:      "#",
		Item:       "-",

		Dashboard:  "[D]",
		WorkItems:  "[W]",
		References: "[R]",
		FutureWork: "[F]",
	}

	// Nerdfont symbols (for developers)
	NerdfontSymbols = SymbolSet{
		Completed:  "",
		InProgress: "",
		Pending:    "",
		Failed:     "",

		High:       "",
		Medium:     "",
		Low:        "",

		Right:      "",
		Left:       "",
		Up:         "",
		Down:       "",

		Bullet:     "",
		Arrow:      "",
		Branch:     "",
		Reference:  "",
		Group:      "",
		Item:       "",

		Dashboard:  "",
		WorkItems:  "",
		References: "",
		FutureWork: "",
	}

	// Minimal symbols
	MinimalSymbols = SymbolSet{
		Completed:  "✓",
		InProgress: "○",
		Pending:    "●",
		Failed:     "!",

		High:       "!",
		Medium:     "-",
		Low:        "·",

		Right:      ">",
		Left:       "<",
		Up:         "^",
		Down:       "v",

		Bullet:     "·",
		Arrow:      ">",
		Branch:     "|",
		Reference:  "~",
		Group:      "+",
		Item:       "-",

		Dashboard:  "D",
		WorkItems:  "W",
		References: "R",
		FutureWork: "F",
	}
)

// Get symbol set by name
func GetSymbolSet(name string) SymbolSet {
	switch name {
	case "ascii":
		return ASCIISymbols
	case "nerdfont":
		return NerdfontSymbols
	case "minimal":
		return MinimalSymbols
	default:
		return UnicodeSymbols
	}
}