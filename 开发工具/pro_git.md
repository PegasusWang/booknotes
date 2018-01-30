# 3: 分之
提交的时候会保存一个 commit 对象：它包含一个指向暂存内容快照的指针，作者和相关信息，以及一定数量（也可能没有）指向该提交对象直接祖先的指针。
分之：本质上是一个指向 commit 对象的指针，新建一个分之就是创建一个新的分之指针。git 鼓励多用分之，创建和销毁非常廉价。
git 保存一个名为 HEAD 的特殊指针，它是一个指向你正在工作中的本地分之的指针（用来区分在哪个分之工作）
Fast Forward: 如果顺着一个分之可以到达另一个分之，git 在合并两者的时候只会简单地把指针前移，没有分歧需要解决，这个过程叫做快进(fast forward)。
merge commit: 如果有分叉，将合并的结果作为一个新的快照，然后创建一个指向它的commit
git push [远程名] [本地分之]:[远程分支]

合并(merge)与衍合(rebase):
rebase: 一个分之里提交的改变在另一个分之里重放一遍。
原理：回到两个分之（你所在的分之和你要rebase进去的分之）的"共同祖先"，提取你所在分之每次提交产生的差异(diff)，把这些差异分别保存在临时文件里，然后从当前分之
转换到你需要 rebase 的分之，依序施用每一个diff 文件。按照每行发生的顺序重演发生的改变。
git rebase [主分支] [特性分之] 会先检出特性分之，然后在主分支上重演。

rebase 黄金定律：永远不要rebase 那些已经推送到公共仓库的更新。(rebase实际上抛弃了一现存的commit 而创造了一些类似的不同的新的 commit)
如果把rebase 当成一种在推送之前清理提交历史的手段，而且仅仅rebase 那些永远不会公开的commit，就不会有任何问题。

# 4: 服务器上的 Git
四种协议：本地协议；ssh协议；git协议；http协议


# 5：分布式 Git
本章讲了不同大小团队使用 git 的工作流

为项目做贡献：
如何撰写提交说明，保持一行。
不要多余空白字符
每次提交限定与完成一次逻辑功能，适当分解为多次小更新。
rewind （回退）


# 6: 项目的管理

rebase  和 cherry-pick:
cherry-pick 类似于针对某次特定提交的 rebase


# 7: git 工具
交互式暂存： git add -i

stash(储藏)：git stash apply stash@{2}

改变最近一次提交说明：git commit --amend  修改提交信息，但是不要在最近一次提交被推送后还去修改
修改多个提交说明： git rebase -i   (注意千万不要涵盖你已经push到中心服务器的提交)

核弹级选项：filter-branch: 比如错误提交了二进制文件或者密码配置文件。会大面积修改提交历史，不要在公开项目搞。

使用 Git 调试：
git blame
git bisect 二分查找

子模块：
git submodule add git://github.com/XXXX XXXX


# 8: 自定义 Git

配置代码merge 和diff 工具


Git 挂钩：客户端和服务器端
在 .git/hooks 下的可执行文件，刻意激活该挂钩脚本


# 9: Git 与其他系统
git svn ： 桥接工具

# 10: Git 内部原理
本质上 git 是一套内容须知(content-addressable)文件系统
如何清理误假如的大文件
