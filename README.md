# pritty

Encodes images to iTerm / Kitty / SIXEL (terminal) inline graphics protocols.

[![Go Reference](https://pkg.go.dev/badge/github.com/arzumify/pritty.svg)](https://pkg.go.dev/github.com/arzumify/pritty)
![rasterm sample output](screenshot.png)

## Supported Image Encodings

- **Kitty**
- **iTerm2 / WezTerm**
- **Sixel**

## Compatibility matrix

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

