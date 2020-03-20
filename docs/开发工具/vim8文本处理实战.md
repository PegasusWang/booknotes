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
