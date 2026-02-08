b站视频教程 https://www.bilibili.com/video/BV1jfiQBhEiH

# 01 认识大模型
- AI(Artificial Inteligence, AI) 研究、设计、构建具备智能、学习、推理和行动能力计算机和机器。
- GAI(Generative AI) 生成式人工智能。让机器能够产生复杂有结构的物件。

机器学习：监督学习、非监督学习、强化学习

深度学习: 机器学习一个分支。主要是使用神经网络模型（多个隐藏层组成）对数据进行学习和表示。
试图模拟大脑的工作方式，灵感来自于神经生物学，通过对大量数据学习，自动提取数据的高层次特征和模式，从而实现图像识别、语音
识别、自然语言处理等任务。神经网络 transformer

# 02 大模型应用场景

- 自然语言处理(NLP) 情感分析、文本归类、信息抽取、数学问题、角色扮演、编程问题、创作问题
- 语音处理(SLP)
- 图像视频处理

# 03 Gradio 快速入门 简易WebUI框架
Gradio: Build machine learning apps in Python. Create web interfaces for your ML models in minutes. Deploy anywhere, share with anyone.

- 可以为模型/API快速构建 demo 或者 web 应用
- 官网： https://www.gradio.app/
- 快速开始：https://www.gradio.app/guides/quickstart

一个简单的demo:

```
import gradio as dr


def reverse_text(text):
    return text[::-1]


demo = gr.Interface(
    fn=reverse_text,
    inputs="text",
    outputs="text")
)

demo.launch()
```

# 04 提示词工程入门

## 什么是提示工程
提示工程(Prompt EEnginerring)是一项通过优化提示词(Prompt)和生成策略，从而获得更好的模型返回结果的工程技术。
- 好的prompt 需要不断调优
- 说清楚自己到底想要什么，要具体
- 不要让机器猜测太多。告诉细节
- 提示工程有技巧，灵活掌握事半功倍。

提示词构成：
- 指示(Instruction): 描述让它做什么
- 上下文(context): 给出任务相关的背景信息
- 例子(Examples): 给出一些例子，让模型知道怎么回复
- 输入(Input): 任务的输入信息
- 输出(Output Format) : 输出格式，想要什么形式的输出

## 常见的提示工程方法

- Instruction = """根据下面的上下文回答问题。保持答案简短且准确"""
- context = ""Teplizumab起源于应该IE 位于新泽西的药品公司"
- query =  ""OKT3最初是从什么来源提取的"
- prompt = f""{Instruction} ### 上下文{contex} ### 问题: {query}"""

shot learning 样本学习
- one-shot learning: 只给一个 example
- few-shot learning: 多个 examples
- zero-shot learning: 不给任何 examplex

Chain-of-Thought 思维链(COT)
- 定义：把一个复杂任务，拆解成多个稍微简答的任务，让大语言模型分步思考问题
- 提示词：小明有 20 个苹果，吃了 2 个，又买了 5 个，小明现在一共多少个苹果？

# 05 提示词应用实战
## 应用案例

- 短视频脚本制作
- 小红书爆款文案
- 模拟面试

总结：
- 写清楚的指令，将复杂任务分解为更为简单的子任务。系统地测试更改
- 提供参考文本，使用外部工具
- 早期测试和学习
- 引入新信息，可靠地复制复杂的风格或方法

# 06 RAG 检索增强生成介绍
## 什么是 RAG
为什么需要RAG ? 大模型的知识来自于训练数据，存在以下局限
- 知识过时(无法知道训练后的事件)
- 幻觉问题(编造看似合理但是错误的信息)
- 缺乏特定领域知识(如公司内部文档、专业数据库)

RAG(检索增强生成): 结合信息检索与生成模型的新型架构，其核心思想是利用外部知识库或者文档集合为大模型提供实时、准备的背景
信息，从而弥补大模型局限性。

组成：
- 检索模块: 知识库中检索与当前输入问题相关的片段或者文档
- 生成模块：基于检索结果和原始输入，通过大模型生成准确、丰富的回答

应用场景： 电商智能客服系统的几种思路
- 传统 AI 1.0 + 检索
- 传统 AI 1.0 + 生成
- 传统 AI 2.0 + 检索生成

# 07 RAG 系统构建案例分析
基于文档的 LLM 回复系统搭建

- 用户 query
- 企业知识 -> 文档(chunk) -> split(chunk1、chunk2) -> (chunk->vector)向量数据库

# 08 文档分割
分割方式：
- 根据句子分割。句子段落，一个句子，一个 chunk  。 split_by_sentence 
- 按照字符数切分。 设置固定的字符数，不连贯 。 split_by_frixed_sentence_count
- 按照固定字符。 设置固定字符，结合一定的重复字符。  split_by_fixed_sentence_count
- 递归方法。设置固定字符，结合一定的重复字符 在加对应的语义。 langchain 库 RecusiveCharacterTextSplitter
- 根据语义进行分割 语义

视频中代码使用了re 正则模块示例。

# 09 文本向量化
## Vector Embedding Example

- Tokenization -> Create Embedding
阿里云百炼。 文本信息向量化 https://help.aliyun.com/zh/model-studio/embedding

## 向量相似度计算：
- 余弦距离Cosine: 基于两个向量夹角的余弦值来衡量近似度。可以用 python np 库计算。 np.dot/np.linalg.norm(A)


# 10 向量数据库和原生 RAG 项目实战
- 向量数据库(vector datebase)。也叫做适量数据库，主要用于存储和处理向量数据
- 检索方法：
  - 单独比较
  - index
  - Approximate search (近似搜索)。Approximate Nearest Neighbor Search (ANNS，近似最近邻搜索)
    - LSH 局部敏感哈希
    - IVF(倒排文档) + PQ(乘积量化)
    - HNSW (chroma)
    - DiskANN

视频里用 chromadb 的 client 存储到了本地作为示例。

文本 -> 向量化 -> 存储向量化数据库获取最相似结果 -> 传递大模型 -> 返回结果

# 11 倒排索引和KNN
倒排索引是一种常见的向量数据库索引结构，用于快速定位和查询向量相似的数据项
通过构建一个映射，将每个向量的特征与包含该特征的向量关联起来。当查询一个向量时，可以通过倒排索引快速找到相似特征值的向量。

分词：Tokenization。 大模型的应用再2025年会有哪些发展-> 大模型 的 应用 在 2025年 会 有哪些 发展

KNN 搜索：Knn-k Nearest Neighbor 。 k紧邻搜索，将查询语句转成向量，然后再求该向量与数据库中的向量相似度最高，距离最近的
向量集。

Brute force search 暴力搜索。 查询向量与数据库中每一个向量的距离，来评估向量之间的相似度，最终选择距离最近的k个数据点


# 12 ANN 近似最临近搜索和聚类索引
Approximate nearest neighbor search. 近似搜索
权衡检索的精度和效率。通过牺牲一部分精度，提升搜索速度。通过构建专门的索引结构index， ANN能够有效缩小搜索空间，
而不是对整个数据库进行全面的比较，从而快速定位到与查询向量近似的结果。

索引构建方式：
- 基于树的索引 tree-based index
- 基于聚类的索引 cluster-based index
- 基于图的索引 graph-based index

聚类索引-空间分块

欧氏距离: 欧几里得距离。最常见的距离度量，多为空间中两个点之间的绝对距离。

聚类索引-引用：
- 基于内容的推荐系统。根据过去看过的其他电影推荐一部电影
- 通过KNN，系统确实为用户推荐了相关电影，但是查询时间长
- 如果采用倒排索引，系统推荐 5 个最相关的电影，搜索时间比KNN快20 倍

# 13 PQ乘积量化
一个10242 维的向量有多大呢？
- 向量一般使用单精度浮点数表示
- 一个单精度浮点数32位(bit)
- 1byte = 8bit
- 1024*32/8 = 4096 个字节  = 4kb

向量压缩：Product Quantization 乘积量化

originnal vector ->  subvectors of equal size -> clustering via k-means -> Quantized subvectors -> PQ code (8 个数字)

1024个浮点数变成了 8 个数字(pq code)。 4kb->8byte

距离计算：
- 查询字段 q 分割为相同的子段
- 对于每个q， 提前计算和所有中心点的欧氏距离
- 距离存储在 距离表(distance table)


# 14 什么是提示工程
提示工程(Prompt Engineering)也叫做指令工程。就是探讨如何设计出最佳提示词，用于指导语言模型帮助我们高效完成某项任务。
Prompt(提示词) 即发送给大模型的指令，比如“讲个笑话”。

# 15 Prompt 的组成元素
prompt 的组成主要包括 指令(Instruction)， 输入数据(input data)、上下文(context)和输出指示器(output indicator)

- 指令：想要模型执行的特定任务或指令
- 上下文：包含外部信息或额外的上下文信息，引导语言模型更好地响应。
- 输入数据： 用户输入的内容或者问题
- 输出指示：指定输出的类型或者格式

从 prompt 的内容和形式，可以分为：
- 零样本提示(zero-shot prompts): 用户仅提供了一个任务描述
- 少样本提示(few-shot promots): 用户提供如何完成任务的示例

区别本质在于上下文的多寡，上下文越多，得到的回答越准确。

# 16 OpenAI 调用



# 17 少样本提示


# 18 思维链 COT
通过让大模型逐步参与，将一个复杂问题分解为一步一步的子问题并依次求解的过程可以显著提升大模型的性能。
这一系列推理的中间步骤成为思维链(Chain of Thought)。

# 19 Ltm 提示方法
Least-to-most 从最少到最多。

`To solve __, we need to first sovle`


# 20 思维树 TOT
self-consistency 自我一致性。
首先利用 cot 生成多个推理路径和答案，最终选择答案出现最多的作为最终答案输出。

处理大规模或者复杂任务时，将问题或者任务分解成一系列子问题或者子任务，这些子问题或者子问题进一步细化，形成树状结构，从而
使得复杂问题变得容易理解和管理。

- 思维分解: 拆分任务。
- 思维生成：为下一个思维步骤生成 k 个候选者
- 状态评估：评估候选者解决问题的进展
- 搜索算法: 根据状态，探索最有希望完成任务的分支

# 21 思维树算24数代码落地
https://www.bilibili.com/video/BV1jfiQBhEiH?vd_source=7ccfa1fd47ec9e99147d0cdae6f1d1a7&spm_id_from=333.788.player.switch&p=21
game24_tot.py

# 22 Prompt 的攻击与防护
劫持语言模型的输出过程，它允许黑客使模型说出任何他们想要的话。在提示词注入攻击中，攻击人尝试通过提供包含恶意内容的输入，
来操纵语言模型的输出。


# 23 什么是 RAG
结合信息检索和文本生成的技术。

# 24 RAG 的原理

# 25. RAG应用案例：阿里云AI助理

# 26. 动手实验1: 创建 RAG 应用

在阿里云百炼上创建 RAG 智能体。

# 27. 动手实验2: 连接钉钉机器人

创建连接流

钉钉和百炼的连接到一起

# 28. 提升索引准确率

# 29. 让问题更好理解

# 30. 改造信息抽取途径


# 31. Langchain 介绍

# 32. Langchain 核心组件

# 33. Langchain 的输入、输出封装

# 34. Langchain 的数据连接封装

# 35. 对话历史管理


# 36. Langchain Expression Language (LCEL)

# 37. 智能体架构 Agent


# 38 LangServe 和 LangChain.js


# 39 认识大模型 Agent

# 40 Agent Prompt 模板设计

# 41 Agent Tuning
