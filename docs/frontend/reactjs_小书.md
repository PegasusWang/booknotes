# http://huziketang.com/books/react/


### 使用 JSX 描述 UI 信息
- JSX 是 JavaScript 语言的一种语法扩展，长得像 HTML，但并不是 HTML。
- React.js 可以用 JSX 来描述你的组件长什么样的。
- JSX 在编译的时候会变成相应的 JavaScript 对象描述。
- react-dom 负责把这个用来描述 UI 信息的 JavaScript 对象变成 DOM 元素，并且渲染到页面上。

### 事件监听

```js
 class Title extends Component {
  handleClickOnTitle () {
    console.log('Click on title.')
  }

  render () {
    return (
      <h1 onClick={this.handleClickOnTitle}>React 小书</h1>
    )
  }
}
```

没有经过特殊处理的话，这些 on* 的事件监听只能用在普通的 HTML 的标签上，而不能用在组件标签上。也就是说，<Header onClick={…} /> 这样的写法不会有什么效果的。这一点要注意，但是有办法可以做到这样的绑定，以后我们会提及

### 高阶组件
高阶组件是一个函数（不是组件），传给他一个组件返回一个新的组件，新的组件使用传入的组件作为子组件。
是一个函数，接受一个组件，修改后返回。

```
import React, { Component } from "react";

export default WrappedComponent => {
  class NewComponent extends Component {
    render() {
      return <WrappedComponent />;
    }
  }
  return NewComponent;
};
```
为了组件之间代码复用，高阶组件内部的包装组件和被包装组件之间通过 props 传递数据。

### context

React.js 的 context 就是这么一个东西，某个组件只要往自己的 context 里面放了某些状态，这个组件之下的所有子组件都直接访问这个状态而不需要通过中间组件的传递。一个组件的 context 只有它的子组件能够访问，它的父组件是不能访问到的，你可以理解每个组件的 context 就是瀑布的源头，只能往下流不能往上飞。

### Redux

Redux 是一种架构模式（Flux 架构的一种变种），它不关注你到底用什么库，你可以把它应用到 React 和 Vue，甚至跟 jQuery 结合都没有问题。而 React-redux 就是把 Redux 这种架构模式和 React.js 结合起来的一个库，就是 Redux 架构在 React.js 中的体现。
“模块（组件）之间需要共享数据”，和“数据可能被任意修改导致不可预料的结果”之间的矛盾。

Pure function: 一个函数的返回结果只依赖于它的参数，并且在执行过程里面没有副作用，我们就把这个函数叫做纯函数
- 函数的返回结果只依赖于它的参数。参数确定值就确定，不会依赖于全局变量这种
- 函数执行过程里面没有副作用: 一个函数执行过程对产生了外部可观察的变化那么就说这个函数是有副作用的。比如修改了传入的对象

纯函数很严格，也就是说你几乎除了计算数据以外什么都不能干，计算的时候还不能依赖除了函数参数以外的数据。
为什么要煞费苦心地构建纯函数？因为纯函数非常“靠谱”，执行一个纯函数你不用担心它会干什么坏事，它不会产生不可预料的行为，也不会对外部产生影响。不管何时何地，你给它什么它就会乖乖地吐出什么。如果你的应用程序大多数函数都是由纯函数组成，那么你的程序测试、调试起来会非常方便。


### reducer

```
// http://huziketang.com/books/react/lesson34
function createStore (reducer) {
  let state = null
  const listeners = []
  const subscribe = (listener) => listeners.push(listener)
  const getState = () => state
  const dispatch = (action) => {
    state = reducer(state, action)
    listeners.forEach((listener) => listener())
  }
  dispatch({}) // 初始化 state
  return { getState, dispatch, subscribe }
}
```

createStore 接受一个叫 reducer 的函数作为参数，这个函数规定是一个纯函数，它接受两个参数，一个是 state，一个是 action。

如果没有传入 state 或者 state 是 null，那么它就会返回一个初始化的数据。
reducer 是不允许有副作用的。你不能在里面操作 DOM，也不能发 Ajax 请求，更不能直接修改 state，它要做的仅仅是 ——
初始化和计算新的 state。根据 state 和 action 计算具有共享结构的新的 state.

createStore 现在可以直接拿来用了，套路就是：

```
// 定一个 reducer
function reducer (state, action) {
  /* 初始化 state 和 switch case */
}

// 生成 store
const store = createStore(reducer)

// 监听数据变化重新渲染页面
store.subscribe(() => renderApp(store.getState()))

// 首次渲染页面
renderApp(store.getState())

// 后面可以随意 dispatch 了，页面自动更新
store.dispatch(...)
```


# Smart 组件 vs Dumb 组件

- 只会接受 props 并且渲染确定结果的组件我们把它叫做 Dumb 组件，这种组件只关心一件事情 —— 根据 props 进行渲染。 这个组件可能会在多处被使用到，那么我们就把它做成 Dumb 组件。
- 还有一种组件，它们非常聪明（smart），城府很深精通算计，我们叫它们 Smart 组件。它们专门做数据相关的应用逻辑，和各种数据打交道、和 Ajax 打交道，然后把数据通过 props 传递给 Dumb，它们带领着 Dumb 组件完成了复杂的应用程序逻辑。

  ![](http://huzidaha.github.io/static/assets/img/posts/25608378-BE07-4050-88B1-72025085875A.png)

一旦一个可复用的 Dumb 组件之下引用了一个 Smart 组件，就相当于污染了这个 Dumb 组件树。如果一个组件是 Dumb 的，那么它的子组件们都应该是 Dumb 的才对。
看到对复用性的需求不同，会导致我们划分组件的方式不同。
Smart 组件可以使用 Smart、Dumb 组件；而 Dumb 组件最好只使用 Dumb 组件，否则它的复用性就会丧失。
要根据应用场景不同划分组件，如果一个组件并不需要太强的复用性，直接让它成为 Smart 即可；否则就让它成为 Dumb 组件。
还有一点要注意，Smart 组件并不意味着完全不能复用，Smart 组件的复用性是依赖场景的，在特定的应用场景下是当然是可以复用 Smart 的。而 Dumb 则是可以跨应用场景复用，Smart 和 Dumb 都可以复用，只是程度、场景不一样。

connect 是一个高阶组件(函数)，把 Dump 组件和 context 连接(connect) 起来
