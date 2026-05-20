Output scripts/demo.gif

Set FontSize 16
Set Width 1200
Set Height 720
Set Padding 10
Set Framerate 30
Set TypingSpeed 60ms

# Build the demo binary out of frame; the GIF starts after `Show`.
Hide
Type "cd /home/hech/Documents/opensouce/mash && go build -o /tmp/mash-demo ./scripts/demo && cd /tmp && clear"
Enter
Sleep 5s

Show
Type "./mash-demo"
Sleep 300ms
Enter

# Let the user take in the full 14-connection list.
Sleep 1500ms

# Walk down through SSH, EC2 and into GCP rows so each colour-coded
# type is visible in turn. Minimum 250ms between keystrokes.
Down
Sleep 250ms
Down
Sleep 250ms
Down
Sleep 250ms
Down
Sleep 250ms
Down
Sleep 250ms
Down
Sleep 250ms
Down
Sleep 250ms
Down
Sleep 400ms

# Select the highlighted connection to open the detail panel.
Type "l"
Sleep 4s

# Quit and clean up out of frame.
Type "q"
Sleep 300ms

Hide
Type "rm -f /tmp/mash-demo"
Enter
