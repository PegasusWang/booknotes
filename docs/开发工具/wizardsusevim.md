

## 技巧：如何映射命令自动编辑完成配置之后 source

    " Make easy editing and sourcing vimrc
    command! RefreshConfig source $MYVIMRC <bar> echo "Refreshed vimrc!"
    " <Leader>ev for editing vimrc
    :nnoremap <leader>ev :vsplit $MYVIMRC<cr>
    " <Leader>sv for sourcing vimrc
    :nnoremap <leader>ev :RefreshConfig<cr>


## copy, pasting and vim Registers

- yank and put
- using resgisters

    set clipboard+=unnamedplus  "neovim use system clipboard
    set clipboard=unnamed "use system clicpboard


## undo and redo


undo: u
redo: ctrl + r


## Inserting Text
aio

gi: puts you in the insert mode at the last place you left insert mode

- ctrl+h 删除上一个字符
- ctrl+w 删除 上一个 word
- c-u 删除上一行
- <C-R>" paste from the unnamed register


## Visual Mode

- v : character wise
- V : line wise
- ctrl+v: block wise


## More command Mode

:s (substitute)
:{range}s/{pattern}/{replace}/{flags}


## Splitting Windows

如何映射快捷键

nnoremap <c-h> <c-h><c-w>h


## Tabs

gt/gT/tabonly

## Moving Around Files

- <C-o> go back to the jumplist
- <TAB>or<C-i> move forward within the jumplist
- g;   to go back the changelist
- g,   to go forward the changelist
- %: bewteen matching brackets
- gg/G


## Multifile Editing

args
multiple cursors


## Configuring Vim

#### how do you configure Vim
- use Ex commands in command line mode
- Using a vim congiguration file
- Through a set of limited command line options


## 持久化配置

    ".vimrc demo

    set nocompatible
    filetype plugin indent on
    syntax enable

map: recursive mapping.
noremap: non-recursive mapping。几乎大部分场景都是使用非递归映射


## 定义自己的命令

    command! RefreshConfig source $MYVIMRC
    " then use :RefreshConfig  ex command souce vimrcfile


## Leader Key

为什么需要 leder key？键盘上按键就那么多，要防止快捷键冲突，可以加上前缀 leader。

    " default leader key is \, hard to type。通常用空格或者逗号替换
    let mapleader = ","
    let mapleader = "\<Space>"


## Extending Vim With Plugins

- 安装插件管理器
- 增加必要配置到vimrc
- 添加需要的插件配置
- 使用命令安装、更新或者删除插件

- 常用插件推荐


## 自定义colorscheme
colorschem 或者插件
