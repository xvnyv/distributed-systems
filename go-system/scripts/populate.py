"""
fetch(`http://localhost:8080/write-request`, {
    method: "POST",
    headers: {
      "Content-type": "application/json",
    },
    body: JSON.stringify(item),
  })
  
{
    UserID: "SAMPLE",
    Item: {
        1: {Id: 1, Name: "pencil", Quantity: 3},
        3: {Id: 3, Name: "paper", Quantity: 1}
    },
    VectorClock: [1, 2, 3234, 4],
    Guid: "a257b9ea-3af4-4a8b-b6e8-cfd9b890eadf"
}
"""
import random
import requests

ITEMS = [
    {"Id": 1, "Name": "pencil", "Quantity": 1},
    {"Id": 2, "Name": "pen", "Quantity": 1},
    {"Id": 3, "Name": "paper", "Quantity": 1},
    {"Id": 4, "Name": "notebook", "Quantity": 1},
    {"Id": 5, "Name": "backpack", "Quantity": 1},
    {"Id": 6, "Name": "water bottle", "Quantity": 1},
    {"Id": 7, "Name": "eraser", "Quantity": 1},
    {"Id": 8, "Name": "glue", "Quantity": 1},
    {"Id": 9, "Name": "tape", "Quantity": 1},
    {"Id": 10, "Name": "highlighter", "Quantity": 1},
]

# fmt: off
USERIDS = ["123", "108", "129", "188", "150", "121", "102", "127", "100", 
           "133", "143", "144", "132", "111", "125", "154", "191"]

ENDPOINT = "http://localhost:8080/write-request"

for userid in USERIDS:
    num_items = random.randint(1, 5)
    start = random.randint(0, len(ITEMS)-1)
    items = {}
    for i in range(num_items):
        cur_item = ITEMS[(start+i)%len(ITEMS)]
        cur_item["Quantity"] = random.randint(1,10)
        items[cur_item["Id"]] = cur_item
    data = {"UserID": userid, "Item": items}
    res = requests.post(ENDPOINT, json=data)
    
    if res.status_code != 200:
        print(res.text)
