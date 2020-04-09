# 1. Getting Started

https://github.com/PacktPublishing/Mastering-Vim

Type `:h` search (don't hit Enter yet) followed by Ctrl+D. This will give you a list of help tags containing the
substring search.


# 2. Advanced Editing and Navigation

### Buffer

`:b file` 如果出现了多个包含 file 的 buffer，可以用 tab 轮询每个文件。

### windows

Ctrl+w+o (or :only or :on command) will close all windows except for the current one.

### Fold

set foldmethod=indent

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

### vim-plug 
可以懒加载:

```vim
" 使用 NERDTreeToggle 命令时候加载插件：
Plug 'scrooloose/nerdtreee', { 'on': 'NERDTreeToggle' }
" 指定文件类型加载
Plug 'junegunn/goyo.vim', { 'for': 'markdown' }
```

### Profiling

Profiling slow plugins : `vim --startuptime startuptime.log`

Profiling spedific actions:

```
:profile start profile.log
:profile func *
:profile file *
```

### Mode

在visual 模式下，可以用 o 从选择的开始和结束跳转

可以用 term 模式运行一个命令输出到 buffer

`:term python3 animal.py cat dog`

### Remapping commands

- `:map` for recursive mapping, `:unmap`
- `:noremap` for non-recursive mapping
- `:map g` will display every mapping starting with g

特殊键位映射：

- `<c-_>` ctrl + 字母（下划线这里指的是字母)
- `<a-_>` or `<m-_>` 表示 Alt
- `<s-_>` 表示 Shift
- `<cr>` Enter

### Mode - aware remapping

- :nmap and :nnoremap : Normal mode
- :vmap and :vnoremap : Visual and select modes
- :xmap and :xnoremap : Visual mode
- :smap and :snoremap : Select mode
- :omap and :onoremap : Operator-pending mode
- :map! and :noremap! : Insert and Command-line mode
- :cmap and :cnoremap: Command-line mode

### The leader key

The leader key is essentially a namespace for a user or plugin defined shortcuts.

```
" Map the leader key to a comma.
let mapleader = ','
let mapleader = "\<space>"
```


# 4. Understanding the Text

### Code autocomplete

YouCompleteMe

### Navigating the code base with tags

Exuberant Ctags

```
$ ctags -R .
" vimrc
set tags=tags;

" regerate tags when saving python file
autocmd BufWritePost *.py silent! !ctags -R &
```

- ctrl+] ctrl+t jump and to back in the tag stack
- :tn :tp  next tag and previous tag
- :ts tag select
- g]  select tag menu instead jump to the tag under the cursor

### Undo tree and Gundo


# 5. Build, Test and Execute

### Integrating Git with Vim (vim-fugitive)

'tpope/vim-fugitive'

- Gstatus
- Glog. :copen :cprevious
- Gblame
- Gread checks out the file straight into a buffer for a preview
- Ggrep wraps around git grep
- Gmove moves the files
- Gdelete wraps git remove commands

### Resolving conflicts with vimdiff

`vimdiff a.py b.py`

- `]c [c` move forward and backward
- do or :diffget (do stands for diff obtain) move the change to the active window
- dp or :diffput (stands for diff put) will push the change from the active window

```
git config --global merge.tool vimdiff
git config --global merge.conflictstyle diff3
git config --global mergetool.promptt false
```

git mergetool 进入合并页面。

- LOCAL: current branch
- BASE: common ancestor
- REMOTE: remote branch
- MERGED: merge result

```
<<<<<<< [LOCAL commit/branch]
[LOCAL change]
|||||||| merged common ancestors
[BASE - clostest common ancestor]
========
[REMOTE change]
>>>>>>> [REMOTE commit/branch]
```

- Get a REMOTE change using :diffg R
- Get a BASE change using :diffg B
- Get a LOCAL change using :diffg L

when you are done, use :wqa quit and git commit your merge results.

### Tmux, screen and vim terminal mode

'christoomey/vim-tmux-navigator'

tmux use tpm plugin manager install

### Building and testing
