Output scripts/demo.gif

Set FontSize 18
Set Width 120
Set Height 40
Set Padding 0
Set Framerate 30

Hide
Type "cd /home/hech/Documents/opensouce/mash && go build -o scripts/demo/mash-demo ./scripts/demo && clear"
Enter
Sleep 2s

Show
Type "scripts/demo/mash-demo"
Sleep 500ms
Enter

Sleep 1s

# Navigate down 5 rows to reach ec2-prod-web-us-east (same as cloud_browser_first_ec2 test)
Down
Down
Down
Down
Down

Sleep 2s

# Quit
Ctrl+C

Sleep 500ms

Hide
Type "rm -f scripts/demo/mash-demo"
Enter
