基于Go语言的大规模微服务框架设计



标志性框架：

- Web 服务框架：MVC 架构 (ASP.Net, Rails)
- Web 服务框架：SaaS 与 RESTful (Sinatra)
- 微服务框架: RPC服务(Thrift)
- 微服务架构：容器化与FaaS (Serveless, Istio)


大型微服务框架设计要点：

- 屏蔽业务无关通用技术细节
	- 功能：服务治理，虚拟化，水平扩容，问题定位，性能压测，系统监控，兼容遗留系统。。。
  - 工具链：项目模板，代码生成器，文档生成器，发布打包脚本。。
  - 设计风格：interceptors，组合模式，依赖注入

- 让不可靠调用变得可靠
	- rpc调用 - 函数调用
	- 访问基础服务 - 访问本地存储
	- 服务拆分/合并- 类拆分/合并


- https://www.youtube.com/redirect?redir_token=M8Xqv9W1kEn6KadyOmxGSnQcLrh8MTU2NjE4NTkxMUAxNTY2MDk5NTEx&q=https%3A%2F%2Fgithub.com%2Fgopherchina%2Fconference&event=comments&stzid=UgwebnM0VQWV5JiWpgd4AaABAg.8wLkBHZICb78wM1YSOcPap
- https://github.com/gopherchina/conference
