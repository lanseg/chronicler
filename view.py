from datetime import datetime
import re
import sys
import json

if __name__ == '__main__':
    fname = sys.argv[1]
    result = json.loads(open(fname).read())

    users = {}
    for u in result["userMetadata"]:
        users[u["id"]] = u

    for r in sorted(result["records"], key=lambda x: int(x["time"])):
        ts = datetime.fromtimestamp(int(r["time"]))
        authorId = r["source"]["sender_id"]
        author = []
        if authorId in users:
            author.append(users[authorId]["username"])
            if users[authorId]["quotes"]:
                author.append("(" + " ".join(users[authorId]["quotes"]) + ")")
        author.append(authorId)

        text = re.sub('[\n]+', '\n', r["text_content"])
        print(f'[{ts}] {" ".join(author)}:\n{text}\n\n')
