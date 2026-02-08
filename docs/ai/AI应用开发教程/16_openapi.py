from openai import OpenAI

client = OpenAI()

def get_completion(prompt, model="gpt-3.5-turbo"):
    messages = [{"role": "user"}, "content": prompt]
    response  = client.chat.completions.create(
        model=model,
        messages=messages,
    )
    return response.choices[0].message


prompt = """
你好
"""

response = get_completion(prompt=)
print(response.content)
