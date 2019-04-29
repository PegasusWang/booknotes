import requests
url = 'http://127.0.0.1:9000/api/comments'
resp = requests.post(
    url,
    data={
        "name": "wnn",
        "email": "wnn@qq.com",
        "comments": "comment",
        "page_id":"2"
    }
)
print(resp.text)
