# 5 Vim Concepts

常用模式(Modes):

- Normal mdoe
- insert mode
- command mode
- visual mode

# 11. Registers

`"ayy`  拷贝行到 a 寄存器 ，然后使用 `"ap` 粘贴，如果是大写的寄存器名称，会把内容 append 进去而不是替换

# 12. Buffers

buffer 是加载文件的内存区，直到保存文件才会从内存写入硬盘。
可以用 `:ls` 现实所有的 buffer，然后使用 `:5b` 打开序号为 5 的 buffer。
使用 `ctrl-6` 可以在两个 buffer 之间切换(依然是同一个窗口)。（当然不清楚为啥是 6 这个数字）


# 13. Windows, Tabs and Sessions

A window  is a viewport on a buffer, a Tab page is a collections of windows.
可以使用 `-o` 参数为每个打开的文件显示一个 window，比如 `vim -O file1.txt file2.txt`

使用 `:sp` `:vs` 可以横分屏和竖分屏

使用 `ctrl-w h` 也就是 h,j,k,l 可以在不同窗口之间切换。如果觉得麻烦可以映射

`nnoremap <C-H> <C-W><C-H>` 这样就可以使用 `ctrl+h` 来移动到左边的窗口了。

使用 `ctrl-w J` 大写字母可以移动当前窗口到最下方。
使用 `:only` 可以关闭所有其他的分屏

Session: 用来保存 vim 会话。 `:mksession ~/mysession.vim`，可以 `vim -S ~/mysession.vim` 或者 进入vim之后 `:source ~/mysession.vim` 启动。

# 14. Macros

宏：用来记录一系列操作行为，并且能够重放。

- 使用 qa 开始记录内容到寄存器 a 中
- 再次按下 q 退出宏的录制
- 使用 @a 使用宏
- 使用 @@ 重复上次执行的宏

宏每次执行都会重绘，如果觉得慢可以使用 set lazyredraw

# 15. The power of Visual modes

块选配合命令模式。比如 块选多行之后执行 `:normal A;` 可以给选中行都加上分号，或者执行 dot 重复。

# 16. Mappings

设置映射(递归的)

```
:nmap v :version<cr>
:nunmap v
```
使用非递归版本 nnoremap

禁用某个键可以使用 `:noremap <left> <nop>`

# 17. Folding

使用 z 前缀折叠，foldmethod 常用折叠选项有 manual, indent, syntax

# 18. Effective multiple file editing

vim 批量操作：

- `:argdo`: for argument list, 比如 vim file1 file2，参数就是 file1, file2
- `:bufdo`: for buffer list，对所有的缓冲区起作用,比如保存所有文件后推出。 `:bufdo wq`
- `:windo`: for window list，只对可见区域的buffer 起作用

使用命令：

- `:normal[al]` - for running commands in Normal mode
- `:exe[cute]` - for executing commands， 把字符串解释成vim命令， `:execute "echom 'Hello world!'"`

`bufdo exe ":normal Gp" | update` 向所有buffer 文件末尾粘贴内容。update 命令如果文件内容没有变不会更新文件时间戳（w会）

`bufdo exe ":normal! @a" | w`  向所有活动的buffer 执行宏

# 19. Productivity Tips

- Relative numbers: 相对行号，比如向下移动5行，5j，这个时候相对行号就比较直观。`set relativenumber`. 还可以让相对行号只在normal 模式起作用

```
augroup toggle_relative_number
autocmd InsertEnter * :setlocal norelativenumber
autocmd InsertLeave * :setlocal relativenumber
```

- Using the Leader Key: leader 提供了给vim映射定义命名空间的方式，比如使用 leader + w 来保存文件

```
let mapleader  = "\<Space>"
nnoremap <Leader>w :w<CR>
```

- Automatic Completion:  ctrl + n/p, ctrl+x ctrl+f 补全文件名

- Using File Template: 创建一个文件模板放到 `~/.vim/templates/html.tpl`，然后 `:autocmd BufNewFile *.html 0r
  ~/.vim/templates/html.tpl`

- Repeat The last Ex command: 可以用 `@:` 在mormal 模式下重复上一个 ex command

- Paste text while in insert mode: `ctrl-r 0`

- Delete in Insert mode: ctrl+h 删除一个字符， ctrl-w 删一个单词， ctrl-u 删除一行

- Repeatable operations on search matches: 比如想查询后一个个修改 `/string` 之后 cw 修改单词然后n找到下一个用 `.`
  重复命令。可以使用 cgn 替代， c 执行修改，gn 向前找到下一个匹配并块选中。

- Copy lines without cursor movement: `:20,25co10` 把 20-25行copy到 10 行下

- Move lines without cursor movement:  `:6m28` 第6行移到第28行，然后用 `''` 移动回去之前的指针位置

- Delete lines without cursor movement: `5,10d`

- Generating numbered lists: 生成序列 `:put =range(1,10)` ，甚至 ip 序列 `:for i in range(1,100) | put ='192.168.0.'.i | endfor`

- Increasing or descreasing numbers: ctrl-a/x 可以递增或者递减数字
批量递增数字：

```
a[0] = 1
a[0] = 2
a[0] = 3
a[0] = 4
```

变成

```
a[1] = 1
a[2] = 2
a[3] = 3
a[4] = 4
```

使用 ctrl-v 块选中所有的 0，然后执行 `g ctrl+a` 完美解决。

- Clear highlighted searches: `nmap <silent> ,/ :nohlsearch<CR>`

- Execute multiple commands at once: `%s/Atom/Vim/c | %s/bad/good/c`

- External program intergration: `%!<command>`

- Auto remove trailing whitespace: 笔者用了一个 StripWhitespace 插件来解决的
```
match ErrorMsg '\s\s+$'
autocmd BufWritePre * :%s/\s\+$//e
```

- Navigation through cursor history:
    - Jumplist : 存储光标跳转的每个位置。ctrl-o ctrl-i 分别是向之前（backwards） 和之后(forwards) 跳转
    - Changelist: 存储每个修改的可被 revert 的位置(with undo)。使用 g; 和 g, 向之前(backwards) 和之后(forwards) 跳转
    - `.  command will bring you to your last change.
    - `` which will bring you back to where the cursor was before you made your last jump.
    - `^ this is the position where the cursor was the last time when insert mode was stopped

上边几个方便的跳转命令(https://vi.stackexchange.com/questions/2001/how-do-i-jump-to-the-location-of-my-last-edit):
撇也可以用单引号替代(` -> ')

- Invert selection: 使用 `:v` 命令

```
:g/127.0.0.1/d    删除包含 127.0.0.1 的行
:v/127.0.0.1/d    删除不包含 127.0.0.1 的行
```

- Quickly switching buffers: vim 有个快捷键切换最后一个编辑的buffer。 ctrl-^ ,不是很方便。
```
" Jump back to last edited buffer
nnoremap <C-[> <C-^>
inoremap <C-[> <esc><C-^>
```

# 20. Plugins

使用插件管理器，书中推荐的是 Vundle，不过使用 dein 或者 vim-plug
支持并发管理，速度要快很多。安装过程看下github文档就好

https://vimawesome.com
