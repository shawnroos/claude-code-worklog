# True Centralized Storage Design

## Problem Statement

The current "centralized" storage at project root still has fundamental issues:
- Different branches have different `.claude-work` directories
- Work items become part of git history causing merge conflicts  
- Each worktree/branch has isolated view of work
- No true single source of truth across all projects

## Solution: External Centralized Storage

### Storage Location
```
~/.claude/
├── work-data/                    # ALL work items for ALL projects
│   ├── projects/                 # Project registry
│   │   ├── project-index.json    # Maps project paths to IDs
│   │   └── {project-id}/         # Project metadata
│   ├── work/                     # Work items by project
│   │   └── {project-id}/
│   │       ├── now/
│   │       ├── next/
│   │       ├── later/
│   │       └── closed/
│   └── artifacts/                # Artifacts by project
│       └── {project-id}/
└── config/                       # User configuration
```

### Project Registration

When TUI starts in any repository:
1. Calculate project ID from git remote URL (or path hash)
2. Register project if not exists
3. All operations use external `~/.claude/work-data/`
4. Never create `.claude-work` in repository

### Benefits

1. **True Single Source**: One location for all work across all branches
2. **No Git Conflicts**: Work items never enter git history
3. **Cross-Project View**: Can see work across all projects
4. **Branch Independent**: Same work view from any branch/worktree
5. **Clean Repositories**: No `.claude-work` pollution

### Implementation Plan

#### Phase 1: External Storage Structure
```go
type ExternalStorage struct {
    BaseDir     string // ~/.claude/work-data
    ProjectsDir string // ~/.claude/work-data/projects
    WorkDir     string // ~/.claude/work-data/work
}

func NewExternalStorage() *ExternalStorage {
    homeDir, _ := os.UserHomeDir()
    baseDir := filepath.Join(homeDir, ".claude", "work-data")
    
    return &ExternalStorage{
        BaseDir:     baseDir,
        ProjectsDir: filepath.Join(baseDir, "projects"),
        WorkDir:     filepath.Join(baseDir, "work"),
    }
}
```

#### Phase 2: Project Registration
```go
type ProjectRegistry struct {
    Projects map[string]*Project `json:"projects"`
}

type Project struct {
    ID          string    `json:"id"`
    Path        string    `json:"path"`
    Name        string    `json:"name"`
    RemoteURL   string    `json:"remote_url"`
    CreatedAt   time.Time `json:"created_at"`
    LastAccess  time.Time `json:"last_access"`
}

func (r *ProjectRegistry) RegisterProject(path string) (*Project, error) {
    // Generate ID from remote URL or path
    id := generateProjectID(path)
    
    project := &Project{
        ID:         id,
        Path:       path,
        Name:       filepath.Base(path),
        RemoteURL:  getGitRemoteURL(path),
        CreatedAt:  time.Now(),
        LastAccess: time.Now(),
    }
    
    r.Projects[id] = project
    return project, r.Save()
}
```

#### Phase 3: Data Access Layer
```go
type CentralizedClient struct {
    storage  *ExternalStorage
    registry *ProjectRegistry
    project  *Project
}

func NewCentralizedClient() (*CentralizedClient, error) {
    storage := NewExternalStorage()
    registry := LoadProjectRegistry(storage.ProjectsDir)
    
    // Get current project
    cwd, _ := os.Getwd()
    projectRoot := findProjectRoot(cwd)
    project, err := registry.RegisterProject(projectRoot)
    
    return &CentralizedClient{
        storage:  storage,
        registry: registry,
        project:  project,
    }, nil
}

func (c *CentralizedClient) GetWorkDir() string {
    // Always return external storage location
    return filepath.Join(c.storage.WorkDir, c.project.ID)
}
```

### Migration Strategy

1. Move all existing work items to `~/.claude/work-data/`
2. Update TUI to use external storage exclusively
3. Add `.claude-work` to `.gitignore` in all projects
4. Remove all `.claude-work` directories from repositories

### Future Enhancements

1. **Multi-User Sync**: Share work items between team members
2. **Cloud Backup**: Optional sync to cloud storage
3. **Project Templates**: Reusable work templates
4. **Cross-Project Dependencies**: Link work across projects
5. **Global Search**: Search work across all projects