## Testing
- test sixel with
	- wezterm
- test wez/iterm img with
	- wezterm
- test kitty with
  - wezterm
	- ghostty

## Notes

### terminal features matrix

| terminal | sixel | iTerm2 format | kitty format |
| :---     | :--:  | :--:          | :--:         |
| ghostty  |       |               | Y            |
| iterm2   | Y     | Y             |              |
| kitty    |       |               | Y            |
| rio      | Y     | Y             |              |
| mintty   | Y     | Y             |              |
| mlterm   | Y     | Y             |              |
| putty    |       |               |              |
| rlogin   | Y     | Y             |              |
| wezterm  | Y     | Y             | Y            |
| xterm    | Y     |               |              |

### known responses

#### CSI 0 c

| terminal       | response                                            |
| :----          | :----                                               |
| apple terminal | `\x1b[?1;2c`                                        |
| ghostty        | `\x1b[?62;22c`                                      |
| guake          | `\x1b[?65;1;9c`                                     |
| iterm2         | `\x1b[?62;4c`                                       |
| kitty          | `\x1b[?62;c`                                        |
| rio            | `\x1b[?62;4;6;22c`                                  |
| mintty         | `\x1b[?64;1;2;4;6;9;15;21;22;28;29c`                |
| mlterm         | `\x1b[?63;1;2;3;4;7;29c`                            |
| putty          | `\x1b[?6c`                                          |
| rlogin         | `\x1b[?65;1;2;3;4;6;7;8;9;15;18;21;22;29;39;42;44c` |
| st             | `\x1b[?6c`                                          |
| terminology    | `\x1b[?64;1;9;15;18;21;22c`                         |
| vimterm        | `\x1b[?1;2c`                                        |
| wez            | `\x1b[?65;4;6;18;22c`                               |
| xfce           | `\x1b[?65;1;9c`                                     |
| xterm          | `\x1b[?63;1;2;4;6;9;15;22c`                         |

#### CSI > 0 c

| terminal       | response            |
| :----          | :----               |
| apple terminal | `\x1b[>1;95;0c`     |
| ghostty        | `\x1b[>1;10;0c`     |
| guake          | `\x1b[>65;5402;1c`  |
| iterm2         | `\x1b[>0;95;0c`     |
| rio            | `\x1b[>0;95;0c`     |
| kitty          | `\x1b[>1;4000;19c`  |
| mintty         | `\x1b[>77;30104;0c` |
| mlterm         | `\x1b[>24;279;0c`   |
| putty          | `\x1b[>0;136;0c`    |
| rlogin         | `\x1b[>65;331;0c`   |
| st             | NO RESPONSE         |
| vimterm        | `\x1b[>0;100;0c`    |
| wez            | `\x1b[>0;0;0c`      |
| xfce           | `\x1b[>65;5402;1c`  |
| xterm          | `\x1b[>19;344;0c`   |

#### identifications

| terminal       | values                                      |
| :----          | :----                                       |
| apple terminal | `TERM_PROGRAM="Apple_Terminal"            ` |
| apple terminal | `__CFBundleIdentifier="com.apple.Terminal"` |
| ghostty        | `TERM="xterm-ghostty"                     ` |
| rio            | `TERM="rio"                               ` |
| guake          | `                                         ` |
| iterm2         | `LC_TERMINAL="iTerm2"                     ` |
| kitty          | `TERM="xterm-kitty"                       ` |
| mintty         | `TERM="mintty"                            ` |
| mlterm         | `                                         ` |
| putty          | `                                         ` |
| rlogin         | `                                         ` |
| st             | `                                         ` |
| terminology    | `TERM_PROGRAM=terminology`                  |
| vimterm        | `VIM_TERMINAL is set                      ` |
| wez            | `TERM_PROGRAM="wezterm"                   ` |
| xfce           | `                                         ` |
| xterm          | `                                         ` |

### opinions

- Sixel is a primitive and wasteful format.  Most sixel terminals also support the iTerm2 format--fewer bytes, full color instead of paletted, and no pixel re-processing required.  Much better!

### go stuff

```sh
go tool pprof -http=:8080 ./name.prof
godoc -http=:8099 -goroot="$HOME/go"
go test -v
go mod tidy
https://blog.golang.org/pprof
```

### more reading

- kitty inline images:  https://sw.kovidgoyal.net/kitty/graphics-protocol.html
- iterm2 inline images: https://iterm2.com/documentation-images.html
- xterm ctl seqs:       https://invisible-island.net/xterm/ctlseqs/ctlseqs.html
- sixel ctl seqs:       https://vt100.net/docs/vt3xx-gp/chapter14.html
- libsixel:             https://saitoha.github.io/libsixel/
