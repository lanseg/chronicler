#!/usr/bin/env python3

import collections
import urllib.request
from html import parser

TypeParamDef = collections.namedtuple('TypeParamDef', ['name', 'type', 'description'])
FuncParamDef = collections.namedtuple('FuncParamDef', ['name', 'type', 'required', 'description'])
TypeDef = collections.namedtuple('TypeDef', ['name', 'description', 'params'])

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

def formatProtoParamType(typename):
    if typename.count(' or') > 0:
        return 'bytes'
    dimensions = typename.count('Array of')
    if dimensions > 1:
        return 'bytes'
    result = typename.replace('Array of', '').strip()
    result = {
        'Integer': 'int64',
        'Float': 'float',
        'Float number': 'float',
        'String': 'string',
        'Boolean': 'bool',
        'True': 'bool',
        'False': 'bool'
    }.get(result, result)
    if dimensions == 1: 
        result = 'repeated ' + result
    return result

def formatGoParamType(typename):
    dimensions = typename.count('Array of')
    result = typename.replace('Array of', '').strip()
    result = {
        'Integer': 'int64',
        'Float': 'float32',
        'Float number': 'float32',
        'String': 'string',
        'Boolean': 'bool',
        'True': 'bool',
        'False': 'bool'
    }.get(result, result)
    if result in UNKNOWN_TYPES:
        result = 'interface{}'
    if result[0].isupper():
        result = "*" + result
    for i in range(dimensions):
        result = "[]" + result
    return result

def formatProtobufType(typedef):
    typeComment = formatComment(typedef.description, 0)
    
    result = ''
    for i in range(len(typedef.params)):
        param = typedef.params[i]
        paramComment = formatComment(param.description, 2)
        paramType = formatProtoParamType(param.type)
        result += "\n%s\n  %s %s = %d;\n" % (paramComment, paramType, param.name, i + 1)

    return "message %s {\n%s}" % (typedef.name, resuvlt)


def formatGolangFieldName(field):
    result = ""
    for p in field.split('_'):
        if p == '':
            continue
        result = result + (p[0].upper() + p[1:])
    return result
    
def formatGolangType(typedef):
    typeComment = formatComment(typedef.description, 0)
    
    result = ''
    for param in typedef.params:
        paramComment = formatComment(param.description, 2)
        paramType = formatGoParamType(param.type)
        if (" or "  in paramType) or (" and " in paramType):
            paramType = "interface{}"            
        result += "\n%s\n  %s %s\n" % (paramComment, formatGolangFieldName(param.name), paramType)

    return "%s\ntype %s struct {%s\n}" % (typeComment, typedef.name, result)


def formatAsProto(types):
   result = []
   for parsedType in sorted(parser.types, key=getSortingKey):
       result.append(formatProtobufType(parsedType))
   return '\n'.join(result)


def formatAsGoModule(types):
   result = ['package telegram', '']
   for parsedType in sorted(parser.types, key=getSortingKey):
       result.append(formatGolangType(parsedType))
   return '\n'.join(result)
       
class TypedefCollector(parser.HTMLParser):
    
    buf = ""
    typedef = []
    types = []
    should_collect = False
    
    def handle_starttag(self, tag, attrs):
        if tag in ["h4", "p", "td", "th", "table"]:
            if tag == "h4":
                self.submit_type()
            self.should_collect = True
            self.buf = ""
                

    def handle_endtag(self, tag):
        if tag in ["h4", "p", "td", "th", "table"]:
            self.typedef.append(self.buf)
            if tag == "table":
                self.submit_type()            
            self.should_collect = False
        
    def handle_data(self, data):
        if self.should_collect:
            self.buf += data
            
    def submit_type(self):
        if len(self.typedef) > 0:
            isTypedef = False
            isFuncdef = False
            i = 0
            for i in range(len(self.typedef) - 3):
                if self.typedef[i] == 'Field' and self.typedef[i+1] == 'Type' and self.typedef[i+2] == 'Description':
                    isTypedef = True
                    break
                if self.typedef[i] == 'Parameter' and self.typedef[i+1] == 'Type' and self.typedef[i+2] == 'Required' and self.typedef[i+3] == 'Description':
                    isFuncdef = True
                    break

            endOfDef = i
            params = []
            if isTypedef:
                i += 3
                while i < len(self.typedef) - 2:
                    params.append(TypeParamDef(*self.typedef[i:i+3]))
                    i += 3                    
            elif isFuncdef:
                i += 4
                while i < len(self.typedef) - 3:
                    params.append(FuncParamDef(*self.typedef[i:i+4]))
                    i += 4
            if params: 
              self.types.append(TypeDef(self.typedef[0], "\n".join(self.typedef[1:endOfDef]), params))
            self.typedef = []

def getSortingKey(typedef):
    firstCapital = 0
    for firstCapital in range(len(typedef.name)):
        if typedef.name[firstCapital].isupper():
            break
    return typedef.name[firstCapital:]
    
# with urllib.request.urlopen('http://python.org/') as response:
with open('apiout', 'r') as response:
   parser = TypedefCollector()
   parser.feed(response.read())
   
   print (formatAsGoModule(parser.types))

