#!/usr/bin/env python3

import re
import collections
import urllib.request
from html import parser

TypeDef = collections.namedtuple('TypeDef', ['name', 'description', 'params'])
TypeParamDef = collections.namedtuple('TypeParamDef', ['name', 'type', 'description'])

FuncDef = collections.namedtuple('FuncDef', ['name', 'description', 'params', 'returns'])
FuncParamDef = collections.namedtuple('FuncParamDef', ['name', 'type', 'required', 'description'])

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

TG_PROTO_TYPES = {
    'Integer': 'int64',
    'Float': 'float',
    'Float number': 'float',
    'String': 'string',
    'Boolean': 'bool',
    'True': 'bool',
    'False': 'bool'
}

TG_GO_TYPES = {
    'Integer': 'int64',
    'Float': 'float32',
    'Float number': 'float32',
    'String': 'string',
    'Boolean': 'bool',
    'True': 'bool',
    'False': 'bool'
}

RETURN_TYPE_PARSER = re.compile('[Aa]rray of (\w+)|(\w+) is returned|(\w+) object is returned')

# Common functions
def toCamelCase(usstr, capFirst=True):
    result = "".join(map(lambda x: x[0].upper() + x[1:], usstr.split('_')))
    if not capFirst:
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


# Protobuf formatting
def formatProtoParamType(typename):
    if typename.count(' or') > 0:
        return 'bytes'
    
    dimensions = typename.count('Array of')
    if dimensions > 1:
        return 'bytes'

    result = typename.replace('Array of', '').strip()
    result = TG_PROTO_TYPES.get(result, result)
    if dimensions == 1: 
        result = 'repeated ' + result
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

def formatAsProto(types):
   result = []
   for parsedType in sorted(parser.types, key=getSortingKey):
       result.append(formatProtobufType(parsedType))
   return '\n'.join(result)


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
    typeComment = typedef.description
    
    result = ''
    for param in typedef.params:
        paramComment = formatComment(param.description, 2)
        paramType = formatGoParamType(param.type)
        if (" or "  in paramType) or (" and " in paramType):
            paramType = "interface{}"
        result += "\n%s\n  %s %s `json:\"%s\"`\n" % (paramComment, toCamelCase(param.name), paramType, param.name)

    return "%s\ntype %s struct {%s}" % (formatComment(typeComment, 0), typedef.name, result)

def formatGolangFunc(typedef):
    returnType = ""
    params = []
    for param in typedef.params:
        actualType = formatGoParamType(param.type)
        params.append(toCamelCase(param.name, False) + " " + actualType)
    if typedef.returns:
        returnType = "(%s, error)" % formatGoParamType(typedef.returns)
    else: 
        returnType = "error"
    return (" // func %s (%s) %s {}") % (toCamelCase(typedef.name), ", ".join(params), returnType)

def formatAsGoModule(types):
   types = []
   funcs = []
   for parsedType in sorted(parser.types, key=getSortingKey):
       if parsedType.name[0].islower():
         funcs.append(formatGolangFunc(parsedType))
       else: 
        types.append(formatGolangType(parsedType))
   return "package telegram\n\n%s\n\n%s" % ('\n\n'.join(types), '\n'.join(funcs))
       
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
            isTypeDef = False
            isFuncDef = False
            i = 0
            for i in range(len(self.typedef) - 3):
                if self.typedef[i] == 'Field' and self.typedef[i+1] == 'Type' and self.typedef[i+2] == 'Description':
                    isTypeDef = True
                    break
                if self.typedef[i] == 'Parameter' and self.typedef[i+1] == 'Type' and self.typedef[i+2] == 'Required' and self.typedef[i+3] == 'Description':
                    isFuncDef = True
                    break

            endOfDef = i
            params = []
            if isTypeDef:
                i += 3
                while i < len(self.typedef) - 2:
                    params.append(TypeParamDef(*self.typedef[i:i+3]))
                    i += 3                    
            elif isFuncDef:
                i += 4
                while i < len(self.typedef) - 3:
                    params.append(FuncParamDef(*self.typedef[i:i+4]))
                    i += 4
            
            if not params:
                self.typedef = []
                return
    
            description = "\n".join(self.typedef[1:endOfDef])
            if isTypeDef: 
              self.types.append(TypeDef(self.typedef[0], description, params))
            if isFuncDef:
              returnType = ''
              m = RETURN_TYPE_PARSER.search(description)
              if m:
                  returnType=list(filter(None, m.groups()))[0]
              self.types.append(FuncDef(self.typedef[0], description, params, returnType))
              
            self.typedef = []

def getSortingKey(typedef):
    firstCapital = 0
    for firstCapital in range(len(typedef.name)):
        if typedef.name[firstCapital].isupper():
            break
    return typedef.name[firstCapital:]
    
with urllib.request.urlopen('https://core.telegram.org/bots/api') as response:
   parser = TypedefCollector()
   parser.feed(response.read().decode('utf-8'))
   print (formatAsGoModule(parser.types))

