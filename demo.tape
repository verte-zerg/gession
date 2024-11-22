# Init
Output demo.gif
Set FontSize 24
Set PlaybackSpeed 0.5
Set Width 1920
Set Height 1080
Set Shell fish
Set Padding 10
Set WindowBar Colorful
Set Theme "Monokai Pro"
Env fish_autosuggestion_enabled "0"

# Prepare
Type "# Let's create a new tmux environment"
Enter
Type "tmux -L gession new -s demo-session"
Enter
Sleep 1s
Type "# Let's create some new sessions"
Enter
Sleep 0.5s
Type "make prepare-demo-sessions"
Enter
Sleep 1s
Type "tmux ls"
Enter
Sleep 2s

# Interface and new-session
Type "# Let's create a new session"
Sleep 0.5s
Enter
Type "gession"
Sleep 2s
Enter
Sleep 1s
Up
Sleep 0.5s
Up
Sleep 0.5s
Up
Sleep 0.5s
Right
Sleep 0.5s
Down
Sleep 0.5s
Down
Sleep 1s
Type "demo-another-session"
Sleep 1s
Enter
Sleep 1s
Type "# Now we are in the new session"
Enter

# Split and rename
Type "# Let's delete and rename some sessions"
Enter
Sleep 0.5s
Type "gession"
Sleep 0.5s
Enter
Sleep 0.5s
Type "demosession"
Sleep 1s
Ctrl+R
Sleep 0.1s
Type "first-session" 
Sleep 1s
Enter
Backspace 11
Sleep 1s
Up
Sleep 1s
Ctrl+E
Sleep 2s
Down
Sleep 0.5s
Enter

# Final
Type "# Have fun!"
Enter
Sleep 2s
Ctrl+D
Ctrl+D
