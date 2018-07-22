## 1.基础

css 选择器：
属性选择器： [attr=value]
子选择器： .table > thread
兄弟选择器：.vav-pills >li+li(临近兄弟)   .article h1 ~p(普通兄弟)
伪类：:hover :focus :first-child :last-child :nth-child  多个伪类可以一起用

## 2.整体架构
- css 12栅格系统
- 基础布局组件，比如排版、代码、表格、按钮、表单等
- jQuery
- 响应式设计
- css 组件

栅格系统：
平分 12 份。
栅格系统工作原理如下：
- row 必须包含在 .container 中，以便为其赋予合适的对齐方式和内边距
- 使用 row 在水平方向上创建一组 column
- 具体内容应该放置在 column 内，而只有column 可以作为 column 的直接子元素
- 使用 .row .col-sx-4 这种样式快速创建栅格布局
- 通过设置 padding 创见 column 之间的间隔
- 栅格系统中的列是通过指定 1 -12 的值表示其跨越的范围的

设计思想：
AO模式。A表示append，O是overwrite。不同名样式叠加，同名样式后面覆盖前面的。
css 八大类型样式：
- 基础样式
- 颜色样式 primary success info warning dangeer ，定义规则是 组件名称-颜色类型
- 尺寸样式：一般4种：超小（xs），小型（sm），普通，大型（lg）
- 状态样式：一些可以点击的元素： active 。注意嵌套元素样式
- 特殊样式：特定类型的组建一般只使用某一种或者几种固定的元素
- 并列元素样式
- 嵌套子元素样式
- 动画样式

先从高度等层面考虑这些问题

js 插件系统：
检测规则


## 3. css 布局



## 4. css 组件


## 5. JavaScript 插件
- 动画过度效果：transition.js
- 模态弹窗: modal.js modals.sass。 声明式用法和 javascript 用法
- 下拉式菜单：dropdown.js
- 滚动侦测：scrollspy.js
- 选项卡：tab.js
- 提示框：tooltip.js
- 弹出框：popover.js。默认触发事件是 click
- 按钮：button.js
- 折叠：collapse.js。手风琴风格(accortion)
- 轮播：carousel.js
- 自动定位浮标 affix.js


# 7. Win8 磁贴组件开发



