from datetime import datetime
import sys
import json

if __name__ == '__main__':
    fname = sys.argv[1]
    result = json.loads(open(fname).read())
    for r in sorted(result["records"], key=lambda x: int(x["time"])):
        ts = datetime.fromtimestamp(int(r["time"]))
        author = r["source"]["sender_id"]
        print(f'[{ts}] {author}:\n{r["text_content"]}\n')