import requests, json

token = 'AxZTwBZXs2YK1J5IZa0fyCcfHz9cYM9VjaO627eMSbAXAMdwFluGZSOTB4Vk3RHY'

url = 'https://push.remote.alexanderebhart.de/message?token=' + token
headers = {'content-type': 'application/json'}

messages = [
    {
        "message": "<hr>",
        "title": "NEW TEST",
        "extras": 
        {
            "client::display": 
            {
                 "contentType": "text/html"
            }
        }
    },
    {
        "message": "line <br> break",
        "title": "Linebreak single br",
        "extras": 
        {
            "client::display": 
            {
                 "contentType": "text/html"
            }
        }
    },
    {
        "message": "line \n break",
        "title": "Linebreak single n",
        "extras": 
        {
            "client::display": 
            {
                 "contentType": "text/html"
            }
        }
    },
    {
        "message": """line 
         break""",
        "title": "Linebreak native",
        "extras": 
        {
            "client::display": 
            {
                 "contentType": "text/html"
            }
        }
    },
    {
        "message": "# Heading 1\n\n###### Heading 6\n\nAlt-H1\n=====\n\n**bold**\n\n_curly_\n\n~~Strike through~~\n\n[Link](google.com)\n\n* list1 \n* list2\n\n1. another list\n2. numerated\n\n![alt text](https://github.com/adam-p/markdown-here/raw/master/src/common/images/icon48.png)\n\n`some inline code`\n\n```\nA code block\n```\n\n> Blockquote",
        "title": "Markdown styling",
        "extras": 
        {
            "client::display": 
            {
                 "contentType": "text/markdown"
            }
        }
    },
    {
        "message": "<font data-mx-color='#AA0000'>Red</font>",
        "title": "HTML",
        "extras": 
        {
            "client::display": 
            {
                 "contentType": "text/html"
            }
        }
    },
]

priorities = [
        {
        "message": "Some text",
        "title": "Prio 0",
        "extras": 
        {
            "client::display": 
            {
                 "contentType": "text/html"
            }
        },
        "priority": 0
    },
    {
        "message": "Some text",
        "title": "Prio 3",
        "extras": 
        {
            "client::display": 
            {
                 "contentType": "text/html"
            }
        },
        "priority": 3
    },
    {
        "message": "Some text",
        "title": "Prio 4",
        "extras": 
        {
            "client::display": 
            {
                 "contentType": "text/html"
            }
        },
        "priority": 4
    },
    {
        "message": "Some text",
        "title": "Prio 15",
        "extras": 
        {
            "client::display": 
            {
                 "contentType": "text/html"
            }
        },
        "priority": 15
    },
    {
        "message": "Some text",
        "title": "Prio 21",
        "extras": 
        {
            "client::display": 
            {
                 "contentType": "text/html"
            }
        },
        "priority": 21
    },
    {
        "message": "Some text",
        "title": "Prio -5",
        "extras": 
        {
            "client::display": 
            {
                 "contentType": "text/html"
            }
        },
        "priority": -5
    },
    {
        "message": "Some text",
        "title": "No Prio",
        "extras": 
        {
            "client::display": 
            {
                 "contentType": "text/html"
            }
        },

    },
]


for data in priorities:
    x = requests.post(url, data = json.dumps(data), headers=headers)
    print(x.text)