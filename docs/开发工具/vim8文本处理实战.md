# 3. 插件管理

分析 vim 性能： 使用 `-startuptime` 和 :profile 来分析性能

`:map g` 将显示所有 g 开头映射

`noremap <c-u> :w<cr>` 使用 ctrl-s 保存文件


```vim
let mapleader = "\<space>"
noremap <leader>w :w<cr>
noremap <leader>n :NERDTreeToggle<cr>
```

# 4. 理解文本

本章介绍了几个插件

- YouCompelteMe
- Exuberant Ctags
- Gundo 撤销树浏览

# 5. 构建、测试和运行

- tpope/vim-fugitive

使用 vimdiff 进行代码冲突合并。

```
<<<<<<<< [LOCAL commit/branch]
[LOCAL change]
||||||| merged commong ancestors
[BASE - closest common ancestor]
========
[REMOTE change]
>>>>>>>> [REMOTE commit/branch]
```

- vim-tmux-navigator
- vim-test 执行单测。 :TestNearest :TestFile :TestSuite :TestLast

# 6. 用正则表达式和宏重构代码

介绍了使用正则和 宏录制来进行重构

# 7. 定制自己的 vim

- airline 状态栏

vimrc 中手动折叠记号 `{{{` 。使用 zM 关闭所有折叠。

# 8. Vimscript

```
" 使用 set 为内部选项赋值
set background=dark
" 非内部变量使用 let
let animal_name = 'Miss'
let is_cat = 1 " bool 使用1/0
"vim 中变量和作用域是通过前缀实现的
let g:animal_name = 'miss'
```
