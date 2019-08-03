# PegasusWang 的读书杂记

> Stay Hungry Stay Foolish.


### 本电子书制作和写作方式
使用 mkdocs 和 markdown 构建，使用  Python-Markdown-Math 完成数学公式。
markdown 语法参考：http://xianbai.me/learn-md/article/about/readme.html

安装依赖：
```sh
pip install mkdocs    # 制作电子书, http://markdown-docs-zh.readthedocs.io/zh_CN/latest/
# https://stackoverflow.com/questions/27882261/mkdocs-and-mathjax/31874157
pip install https://github.com/mitya57/python-markdown-math/archive/master.zip

# 或者直接
pip install -r requirements.txt
```

编写并查看：
```sh
mkdocs serve     # 修改自动更新，浏览器打开 http://localhost:8000 访问
# 数学公式参考 https://www.zybuluo.com/codeep/note/163962
mkdocs gh-deploy    # 部署到自己的 github pages, 如果是 readthedocs 会自动触发构建
```

### 访问:

[https://pegasuswang.readthedocs.io/zh/latest/](https://pegasuswang.readthedocs.io/zh/latest/)

[https://pegasuswang.github.io/booknotes/](https://pegasuswang.github.io/booknotes/)
