# Chronicler

A extensible framework and a telegram bot that helps you to expand the long posts and save all valuable data you've found.

Telegram bot receives a message, provides it to the tool which saves and tries to find all supported links or social network
references. 

## How it works

// TODO: add plans here

## Details

### Record protobuf

// TODO: Some description of the protobuf format

### Adapters

Adapter is an interface that allows to find social network, webpage or other resource
references in the text and loads information by that reference in a [RecordSet](#Record protobuf) format. 

Currently supported:
* Telegram (Messages forwarded to the bot)
* Twitter (Links to twitter threads)
* Pikabu (Link to pikabu stories)
* Web pages (Just a regular http/https link)

### Webdriver

Reinventing the wheel. Managed Firefox (marionette) in a headless mode that loads a link
and runs some user javascript to prepare the page. 

Why not plain http client? Well, even simpliest modern web pages rely on scripting a lot 
and instead of investigating queries to the backend or lazy loading mechanisms for each
website I decided to let Firefox do the job. 

// Here goes diagram [request -> loading page -> running userscript -> doing the job]

### Storage

Storage for records.
// TODO: add storage description

### Frontend

Web page.
// TODO: add frontend description

### Monitoring

// TODO: add information about monitoring