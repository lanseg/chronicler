#!/usr/bin/env python3

import re
from enum import Enum
import collections
import urllib.request
from html import parser

SELF_CLOSING = set("area, base, br, col, embed, hr, img, input, keygen, link, menuitem, meta, param, source, track, wbr".split(", "))
ParamDef = collections.namedtuple('ParamDef', ['name', 'param_type', 'optional', 'comment'])
TypeDef = collections.namedtuple('TypeDef', ['name', 'comment', 'params', 'is_function'])

UNKNOWN_TYPES = [
    'ChatMember',
    'InputFile',
    'CallbackGame',
    'InlineQueryResult',
    'InputMessageContent',
    'VoiceChatStarted',
    'InputMedia',
    'BotCommandScope',
    'PassportElementError',
]

TG_GO_TYPES = {
    'Integer': 'int64',
    'Float': 'float32',
    'Float number': 'float32',
    'String': 'string',
    'Boolean': 'bool',
    'True': 'bool',
    'False': 'bool'
}

# Utility code
def formatWord(word):
    if word.lower() in ['id', 'url', 'ip']:
        return word.upper()
    return word[0].upper() + word[1:]
    
def toCamelCase(usstr, capFirst=True):
    result = "".join(map(formatWord, usstr.split('_')))
    if not capFirst and result.upper() != result: 
        return result[0].lower() + result[1:]
    return result


def formatComment(comment, offset, maxWidth = 95): 
    prefix = ' ' * offset + '// '
    currentLines = list(reversed(comment.split('\n')))
    commentLines = []
     
    while currentLines:
        line = currentLines.pop()
        if len(line) <= maxWidth:
            commentLines.append(line)
            continue
        splitAt = maxWidth
        while splitAt >= 0 and not line[splitAt].isspace():
            splitAt -= 1
        commentLines.append(line[:splitAt].strip())
        currentLines.append(line[splitAt:].strip())
    return prefix + ('\n' + prefix).join(commentLines)

# Main code
class Node:
    
    def __init__(self, name, params):
        self.name = name
        self.params = params
        self.children = []
        self.data = []
        
    def find(self, name = "", params = {}): 
        result = []
        toVisit = [self]
        while toVisit:
            node = toVisit.pop()
            if (name == '' or node.name == name) and (params == {} or node.params == params):
                result.append(node)
            toVisit.extend(node.children)
        return result
        
    def expand_data(self):
        result = []
        for d in self.data:
            if isinstance(d, str):
                result.append(d)
            else:
                result.append(d.expand_data())
        return ''.join(result)
        
    def __str__(self):
        return self.name

def nodesToTypedef(title, prefix, table, suffix):
    params = []
    funcdef = title.data[1][0].islower()

    for typedef in table.find('tr')[:-1] if table else []:
        paramNodes = list(map(lambda x: x.expand_data(), typedef.children))
        params.append(ParamDef(
            paramNodes[0],
            paramNodes[1], 
            paramNodes[2] if funcdef else '',
            paramNodes[3] if funcdef else paramNodes[2]))
    return TypeDef(title.data[1], 
                      " ".join(map(lambda x: str(x.expand_data()).strip(), prefix)) + 
                      " ".join(map(lambda x: str(x.expand_data()).strip(), suffix)),
                      params,
                      funcdef)

# Golang formatting params
def formatGoParamType(typename):
    dimensions = typename.count('Array of')
    result = typename.replace('Array of', '').strip()
    result = TG_GO_TYPES.get(result, result)
    if result in UNKNOWN_TYPES or ' or ' in result:
        result = 'interface{}'
    if result[0].isupper():
        result = "*" + result
    for i in range(dimensions):
        result = "[]" + result
    return result

def formatGolangType(typedef):  
    result = ''
    for param in typedef.params:
        paramComment = formatComment(param.comment, 2)
        paramType = formatGoParamType(param.param_type)        
        if (" or "  in paramType) or (" and " in paramType):
            paramType = "interface{}"
        result += "\n%s\n  %s %s `json:\"%s\"`\n" % (paramComment, toCamelCase(param.name), paramType, param.name)
    return "%s\ntype %s struct {%s}" % (formatComment(typedef.comment, 0), typedef.name, result)


class TypedefCollector(parser.HTMLParser):
    nodes = [Node("root", {})]

    def handle_starttag(self, tag, attrs):
        node = Node(tag, dict(attrs))
        parent = self.nodes[-1]
        parent.children.append(node)
        parent.data.append(node)
        if tag not in SELF_CLOSING:
          self.nodes.append(node)

    def handle_endtag(self, tag):
        if tag not in SELF_CLOSING:
          self.nodes.pop()
        
    def handle_data(self, data):
        self.nodes[-1].data.append(data)
    
def parseNodes(data):
   i = 0
   
   while i < len(data):
     while i < len(data) and data[i].name != 'h4':
       i += 1
     if i == len(data): 
         break
     
     title = data[i]
     prefix = []
     table = None
     suffix = []
     i += 1
     
     prefix = []
     while i < len(data) and data[i].name != 'table' and data[i].name != 'h4':
         prefix.append(data[i])
         i += 1
     if i == len(data): 
         break
     
     if data[i].name == 'table':
         table = data[i]
         i += 1
         
     while i < len(data) and data[i].name != 'h4':
         suffix.append(data[i])
         i += 1
     yield nodesToTypedef (title, prefix, table, suffix)
     
     

with urllib.request.urlopen('https://core.telegram.org/bots/api') as response:
    parser = TypedefCollector()
    parser.feed(response.read().decode('utf-8'))
    nodes = parser.nodes[0].find('div', {'id': 'dev_page_content'})[0].children        
    while nodes[0].expand_data() != 'Update':
        nodes = nodes[1:]
    print('package telegram')
    for node in parseNodes(nodes):
        if node.is_function:
            continue
        print(formatGolangType(node))
        print()
     

   
   
