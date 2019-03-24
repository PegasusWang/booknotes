

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
