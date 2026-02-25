//kitty terminal config:

//folder caps letter fixing command : 
//run your terminal:
  echo "set completion-ignore-case on" >> ~/.inputrc

//terminal cursor customization :
//create a kitty.conf

# Enable cursor trail animation
cursor_trail 55

# Control trail fade timing (min_decay max_decay)
cursor_trail_decay 0.1 0.5

# Always show trail (0 = always active)
cursor_trail_start_threshold 0
