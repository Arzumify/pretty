# TODO

- mintty:
	- detection for iTerm format: https://github.com/mintty/mintty/issues/881
- perhaps query tmux directly: TMUX=/tmp/tmux-1000/default,3218,4
- improve terminal identification
	19:VT340
	ESC[>0c = 19;344:0c
	https://invisible-island.net/xterm/ctlseqs/ctlseqs-contents.html
- animation (probably via ffmpeg)
    - gif
    - h264
- clear, d=
    a all
    i id (i=, and optionally p=)
    n newest (optionallay I= and p=)
    c cursor position
    f animation framse
    p specific cell (x=, y=)
    q cell at z-index (x, y, z)
    r id in range (x <= id <= y)
    x placements at column (x)
    y placements at row (y)
    z placementz at z index (z)
