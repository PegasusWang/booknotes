# 1. Getting Started

Type `:h` search (don't hit Enter yet) followed by Ctrl+D. This will give you a list of help tags containing the
substring search.


# 2. Advanced Editing and Navigation

### Buffer

`:b file` 如果出现了多个包含 file 的 buffer，可以用 tab 轮询每个文件。

### windows

Ctrl+w+o (or :only or :on command) will close all windows except for the current one.

### Fold

- zo: open current fold
- zc: close the fold
- zR/zM: open and close all folds

### Navigating text

word: 字母数字下划线。 WORD: 除了空白符（空格，tab，换行）

- `_` beginning of the line and `$` take you to the end of the line
- Shift + {  and Shift + } takes you to the beginning and the end of a paragraph
- Shift + (  and Shift + ) takes you to the beginning and the end a sentence

### Jumping into insert mode

- gi places you in insert mode were you last exited it

### Searching across files

`:vimgrep animal **/*.py` navigate use `:cn :cp` , open quickfix window by using `:copen`


# 3. Follow the Leader - Plugin Management
