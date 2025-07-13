# Key Testing Instructions

The work tracker now has the following key bindings:

## In List View:
- `tab` or `→`: Switch to next tab (NOW/NEXT/LATER)
- `shift+tab` or `←`: Switch to previous tab
- `↑/↓`: Navigate list items
- `enter`: View full post with glamour rendering
- `d`: Toggle detail level in list
- `q` or `ctrl+c`: Quit application

## In Full Post View:
- `esc`: Return to list view
- `q` or `ctrl+c`: Quit application

## Expected Behavior:
1. Start with the fancy list view showing NOW/NEXT/LATER tabs
2. Use arrow keys to navigate list items
3. Press `enter` on any item to see full post with glamour markdown rendering
4. Press `esc` to return to list
5. Press `q` to quit from either view

The key issue was that app-level quit handling was intercepting keys before the view could handle them. This has been fixed by letting the fancy list view handle keys first.