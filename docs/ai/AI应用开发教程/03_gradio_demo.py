import gradio as gr
from dotenv import load_dotenv
import os
import requests
import gradio as dr


def text_demo():
    def reverse_text(text):
        return text[::-1]

    demo = dr.Interface(
        fn=reverse_text,
        inputs="text",
        outputs="text",
    )


# 加载环境变量（从.env文件读取API Key）
# load_dotenv()

# 配置Claude API信息
ANTHROPIC_API_KEY = os.getenv("ANTHROPIC_API_KEY")
API_URL = "https://api.anthropic.com/v1/messages"
# Claude Code推荐使用claude-3-5-sonnet-20240620模型，对代码理解/生成更优
MODEL_NAME = "claude-3-5-sonnet-20240620"


def call_claude_code_api(prompt, max_tokens=1024, temperature=0.7):
    """
    调用Claude Code API处理代码相关请求
    :param prompt: 用户输入的代码需求/问题
    :param max_tokens: 生成的最大token数
    :param temperature: 生成随机性（0-1，越低越精准）
    :return: Claude返回的代码/回答内容
    """
    # 校验API Key
    if not ANTHROPIC_API_KEY:
        return "错误：未配置ANTHROPIC_API_KEY，请在.env文件中设置"

    # 构建请求头
    headers = {
        "x-api-key": ANTHROPIC_API_KEY,
        "Content-Type": "application/json",
        "anthropic-version": "2023-06-01"
    }

    # 构建请求体（符合Claude API格式）
    data = {
        "model": MODEL_NAME,
        "max_tokens": max_tokens,
        "temperature": temperature,
        "messages": [
            {
                "role": "user",
                "content": f"""你是专业的代码助手Claude Code，请解决以下代码相关问题：
{prompt}
要求：代码需可运行，附带清晰注释，说明核心逻辑"""
            }
        ]
    }

    try:
        # 发送请求
        response = requests.post(
            API_URL, headers=headers, json=data, timeout=30)
        response.raise_for_status()  # 抛出HTTP错误

        # 解析响应
        result = response.json()
        return result["content"][0]["text"]

    except requests.exceptions.Timeout:
        return "错误：请求超时，请检查网络或稍后重试"
    except requests.exceptions.HTTPError as e:
        return f"HTTP错误：{e.response.status_code} - {e.response.text}"
    except Exception as e:
        return f"未知错误：{str(e)}"


# 构建Gradio界面
with gr.Blocks(title="Claude Code 代码助手") as demo:
    gr.Markdown("""
    # Claude Code 代码助手
    基于Claude 3.5 Sonnet模型的代码生成/调试/解释工具
    """)

    # 输入区域
    with gr.Row():
        with gr.Column(scale=3):
            prompt_input = gr.Textbox(
                label="输入你的代码需求",
                placeholder="例如：写一个Python函数实现快速排序，附带注释；调试这段代码的bug：xxx；解释这段代码的逻辑：xxx",
                lines=6
            )
        with gr.Column(scale=1):
            max_tokens = gr.Slider(
                label="生成最大Token数",
                minimum=256,
                maximum=4096,
                value=1024,
                step=256
            )
            temperature = gr.Slider(
                label="生成随机性（0=精准，1=创意）",
                minimum=0.0,
                maximum=1.0,
                value=0.7,
                step=0.1
            )

    # 按钮和输出区域
    submit_btn = gr.Button("生成/解答", variant="primary")
    output = gr.Textbox(
        label="Claude Code 回答",
        lines=15,
        interactive=False
    )

    # 绑定事件
    submit_btn.click(
        fn=call_claude_code_api,
        inputs=[prompt_input, max_tokens, temperature],
        outputs=output
    )

# 运行界面
if __name__ == "__main__":
    # demo.launch()
    # share=True可生成公共链接（临时），server_port指定端口
    demo.launch(server_port=7860, share=False)
