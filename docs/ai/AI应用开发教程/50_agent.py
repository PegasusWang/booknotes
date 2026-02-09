from langchain_openapi import ChatOpenAI
from langchain.agents import initialize_agent, AgentType
from langchain.community.agent_toolkits.load_tools import load_tools
from dotenv import load_dotenv
load_dotenv()

# pip install -U langchain

llm = ChatOpenAI(
    temperature=0,
    model="gpt-3.5-turbo"
)

# google search
tools = load_tools(["serpapi", "llm-math"], llm=llm)

agent = initialize_agent(
    tools,
    llm,
    agent=AgentType.ZERO_SHOT_REACT_DESCRIPTION,
    verbose=True,
)


print(agent.run("请问现任美国总统谁？他的年龄的平方是多少?"))
