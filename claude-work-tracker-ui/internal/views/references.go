package views

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"claude-work-tracker-ui/internal/data"
	"claude-work-tracker-ui/internal/models"
)

type ReferenceNode struct {
	ID       string
	Content  string
	Type     string
	Status   string
	Priority string
	Children []ReferenceEdge
	X        int
	Y        int
}

type ReferenceEdge struct {
	TargetID         string
	SimilarityScore  float64
	RelationshipType string
	Confidence       float64
}

type ReferencesModel struct {
	dataClient   *data.Client
	workItems    []models.WorkItem
	nodes        map[string]*ReferenceNode
	selectedNode *ReferenceNode
	viewMode     string // "graph" or "list"
	width        int
	height       int
	loading      bool
	error        error
}

type ReferencesLoadedMsg struct {
	WorkItems []models.WorkItem
	Error     error
}

var (
	graphStyle = lipgloss.NewStyle().
			Padding(1, 2)

	nodeStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1).
			Margin(0, 1)

	selectedNodeStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("170")).
				Background(lipgloss.Color("235")).
				Padding(0, 1).
				Margin(0, 1).
				Bold(true)

	edgeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244"))

	strongEdgeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("170")).
			Bold(true)

	referenceHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")).
				Bold(true)
)

func NewReferencesModel(dataClient *data.Client) *ReferencesModel {
	return &ReferencesModel{
		dataClient: dataClient,
		nodes:      make(map[string]*ReferenceNode),
		viewMode:   "graph",
		loading:    true,
	}
}

func (m *ReferencesModel) Init() tea.Cmd {
	return m.loadReferences()
}

func (m *ReferencesModel) Update(msg tea.Msg) (*ReferencesModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case ReferencesLoadedMsg:
		m.loading = false
		m.workItems = msg.WorkItems
		m.error = msg.Error
		
		if m.error == nil {
			m.buildReferenceGraph()
		}
		
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("r", "R"))):
			m.loading = true
			return m, m.loadReferences()

		case key.Matches(msg, key.NewBinding(key.WithKeys("v", "V"))):
			if m.viewMode == "graph" {
				m.viewMode = "list"
			} else {
				m.viewMode = "graph"
			}
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("j", "down"))):
			m.selectNextNode()
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("k", "up"))):
			m.selectPreviousNode()
			return m, nil
		}
	}

	return m, nil
}

func (m *ReferencesModel) View() string {
	if m.loading {
		return "Loading references..."
	}

	if m.error != nil {
		return fmt.Sprintf("Error loading references: %s", m.error.Error())
	}

	if len(m.workItems) == 0 {
		return "No work items with references found"
	}

	switch m.viewMode {
	case "graph":
		return m.renderGraphView()
	case "list":
		return m.renderListView()
	default:
		return "Unknown view mode"
	}
}

func (m *ReferencesModel) renderGraphView() string {
	var sections []string

	// Header
	header := referenceHeaderStyle.Render("🔗 Reference Graph")
	sections = append(sections, header)

	// Graph visualization
	graph := m.renderASCIIGraph()
	sections = append(sections, graph)

	// Selected node details
	if m.selectedNode != nil {
		details := m.renderNodeDetails()
		sections = append(sections, details)
	}

	// Help
	help := m.renderHelp()
	sections = append(sections, help)

	return graphStyle.Render(strings.Join(sections, "\n"))
}

func (m *ReferencesModel) renderListView() string {
	var sections []string

	// Header
	header := referenceHeaderStyle.Render("🔗 Reference List")
	sections = append(sections, header)

	// List all nodes with their references
	for _, node := range m.nodes {
		nodeSection := m.renderNodeListItem(node)
		sections = append(sections, nodeSection)
	}

	// Help
	help := m.renderHelp()
	sections = append(sections, help)

	return graphStyle.Render(strings.Join(sections, "\n"))
}

func (m *ReferencesModel) renderASCIIGraph() string {
	if len(m.nodes) == 0 {
		return "No nodes to display"
	}

	var lines []string
	
	// Simple layout: arrange nodes in a grid
	nodesPerRow := 3
	row := 0
	col := 0
	
	nodeList := make([]*ReferenceNode, 0, len(m.nodes))
	for _, node := range m.nodes {
		nodeList = append(nodeList, node)
	}
	
	// Sort nodes by ID for consistent ordering
	sort.Slice(nodeList, func(i, j int) bool {
		return nodeList[i].ID < nodeList[j].ID
	})
	
	for _, node := range nodeList {
		node.X = col * 25
		node.Y = row * 6
		
		col++
		if col >= nodesPerRow {
			col = 0
			row++
		}
	}
	
	// Create a grid to draw on
	maxRows := row + 1
	maxCols := nodesPerRow
	grid := make([][]string, maxRows*6)
	for i := range grid {
		grid[i] = make([]string, maxCols*25)
		for j := range grid[i] {
			grid[i][j] = " "
		}
	}
	
	// Draw nodes
	for _, node := range nodeList {
		content := node.Content
		if len(content) > 15 {
			content = content[:12] + "..."
		}
		
		var style lipgloss.Style
		if m.selectedNode != nil && m.selectedNode.ID == node.ID {
			style = selectedNodeStyle
		} else {
			style = nodeStyle
		}
		
		nodeDisplay := style.Render(fmt.Sprintf("[%s] %s", node.Type, content))
		
		// Place node in grid (simplified)
		if node.Y < len(grid) && node.X < len(grid[0]) {
			for i, char := range nodeDisplay {
				if node.X+i < len(grid[node.Y]) {
					grid[node.Y][node.X+i] = string(char)
				}
			}
		}
	}
	
	// Draw edges (simplified - just show connections)
	for _, node := range nodeList {
		for _, edge := range node.Children {
			if targetNode, exists := m.nodes[edge.TargetID]; exists {
				// Draw a simple line indicator
				edgeDisplay := fmt.Sprintf("→ %s (%.2f)", edge.RelationshipType, edge.SimilarityScore)
				
				if edge.SimilarityScore > 0.7 {
					edgeDisplay = strongEdgeStyle.Render(edgeDisplay)
				} else {
					edgeDisplay = edgeStyle.Render(edgeDisplay)
				}
				
				// Place edge below the node
				edgeY := node.Y + 2
				if edgeY < len(grid) && node.X < len(grid[0]) {
					for i, char := range edgeDisplay {
						if node.X+i < len(grid[edgeY]) {
							grid[edgeY][node.X+i] = string(char)
						}
					}
				}
				
				_ = targetNode // Use targetNode to avoid unused variable
			}
		}
	}
	
	// Convert grid to string
	for i := range grid {
		line := strings.TrimRight(strings.Join(grid[i], ""), " ")
		lines = append(lines, line)
	}
	
	return strings.Join(lines, "\n")
}

func (m *ReferencesModel) renderNodeListItem(node *ReferenceNode) string {
	var sections []string
	
	// Node header
	header := fmt.Sprintf("[%s] %s (%s)", node.Type, node.Content, node.Status)
	if m.selectedNode != nil && m.selectedNode.ID == node.ID {
		header = selectedNodeStyle.Render(header)
	} else {
		header = nodeStyle.Render(header)
	}
	sections = append(sections, header)
	
	// References
	if len(node.Children) > 0 {
		sections = append(sections, "  References:")
		for _, edge := range node.Children {
			if targetNode, exists := m.nodes[edge.TargetID]; exists {
				refLine := fmt.Sprintf("    → %s: %s (%.2f, %s)", 
					edge.RelationshipType, 
					targetNode.Content, 
					edge.SimilarityScore, 
					edge.RelationshipType)
				
				if edge.SimilarityScore > 0.7 {
					refLine = strongEdgeStyle.Render(refLine)
				} else {
					refLine = edgeStyle.Render(refLine)
				}
				
				sections = append(sections, refLine)
			}
		}
	}
	
	return strings.Join(sections, "\n")
}

func (m *ReferencesModel) renderNodeDetails() string {
	node := m.selectedNode
	if node == nil {
		return ""
	}
	
	var sections []string
	
	sections = append(sections, "Selected Node Details:")
	sections = append(sections, fmt.Sprintf("ID: %s", node.ID))
	sections = append(sections, fmt.Sprintf("Type: %s", node.Type))
	sections = append(sections, fmt.Sprintf("Status: %s", node.Status))
	sections = append(sections, fmt.Sprintf("Priority: %s", node.Priority))
	sections = append(sections, fmt.Sprintf("Content: %s", node.Content))
	
	if len(node.Children) > 0 {
		sections = append(sections, fmt.Sprintf("References: %d", len(node.Children)))
		for _, edge := range node.Children {
			if targetNode, exists := m.nodes[edge.TargetID]; exists {
				refDetail := fmt.Sprintf("  → %s (%.2f confidence)", 
					targetNode.Content, edge.Confidence)
				sections = append(sections, refDetail)
			}
		}
	}
	
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1).
		Render(strings.Join(sections, "\n"))
}

func (m *ReferencesModel) renderHelp() string {
	help := []string{
		"j/k: Navigate nodes",
		"v: Toggle view mode",
		"r: Refresh",
	}
	
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Render(strings.Join(help, " • "))
}

func (m *ReferencesModel) buildReferenceGraph() {
	m.nodes = make(map[string]*ReferenceNode)
	
	// Create nodes for all work items
	for _, item := range m.workItems {
		node := &ReferenceNode{
			ID:       item.ID,
			Content:  item.Content,
			Type:     item.Type,
			Status:   item.Status,
			Priority: item.GetPriority(),
			Children: []ReferenceEdge{},
		}
		
		// Add references if they exist
		if item.HasSmartReferences() {
			for _, ref := range item.Metadata.SmartReferences {
				edge := ReferenceEdge{
					TargetID:         ref.TargetID,
					SimilarityScore:  ref.SimilarityScore,
					RelationshipType: ref.RelationshipType,
					Confidence:       ref.Confidence,
				}
				node.Children = append(node.Children, edge)
			}
		}
		
		m.nodes[item.ID] = node
	}
	
	// Set initial selected node
	if len(m.nodes) > 0 && m.selectedNode == nil {
		for _, node := range m.nodes {
			m.selectedNode = node
			break
		}
	}
}

func (m *ReferencesModel) selectNextNode() {
	if len(m.nodes) == 0 {
		return
	}
	
	nodeList := make([]*ReferenceNode, 0, len(m.nodes))
	for _, node := range m.nodes {
		nodeList = append(nodeList, node)
	}
	
	sort.Slice(nodeList, func(i, j int) bool {
		return nodeList[i].ID < nodeList[j].ID
	})
	
	if m.selectedNode == nil {
		m.selectedNode = nodeList[0]
		return
	}
	
	for i, node := range nodeList {
		if node.ID == m.selectedNode.ID {
			if i+1 < len(nodeList) {
				m.selectedNode = nodeList[i+1]
			} else {
				m.selectedNode = nodeList[0] // Wrap around
			}
			break
		}
	}
}

func (m *ReferencesModel) selectPreviousNode() {
	if len(m.nodes) == 0 {
		return
	}
	
	nodeList := make([]*ReferenceNode, 0, len(m.nodes))
	for _, node := range m.nodes {
		nodeList = append(nodeList, node)
	}
	
	sort.Slice(nodeList, func(i, j int) bool {
		return nodeList[i].ID < nodeList[j].ID
	})
	
	if m.selectedNode == nil {
		m.selectedNode = nodeList[len(nodeList)-1]
		return
	}
	
	for i, node := range nodeList {
		if node.ID == m.selectedNode.ID {
			if i-1 >= 0 {
				m.selectedNode = nodeList[i-1]
			} else {
				m.selectedNode = nodeList[len(nodeList)-1] // Wrap around
			}
			break
		}
	}
}

func (m *ReferencesModel) loadReferences() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		workItems, err := m.dataClient.GetAllWorkItems()
		
		// Filter to only items with references
		var itemsWithRefs []models.WorkItem
		for _, item := range workItems {
			if item.HasSmartReferences() {
				itemsWithRefs = append(itemsWithRefs, item)
			}
		}
		
		return ReferencesLoadedMsg{
			WorkItems: itemsWithRefs,
			Error:     err,
		}
	})
}