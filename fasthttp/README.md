# fasthttp samples

server programs that demonstrate fasthttp methods

# azulserver

server that uses a minimal html file and uploads js code to build the dom

## fasthttpServer.go


## fasthttpServerV2.go


## fasthttpServerV3.go

This program is a simple server. Has no cli interpreter

## fasthttpServerV5.go

 - added cli interface
This program is a derivative of fasthttpServerV4, since that program has CLI.
./fasthttpServerV5 /port=<portno> [/idx=indexfile] [/dbg]

  - improve performance by reducing the azulLib
remove the material design from pageHtml class

  - no embedding



## fasthttpServerV6Old.go

change to fasthttpServerNV1.go

  - checked text and binary transmission with websocket
  - embedding of script files and index file to eliminate multiple roundtrips
  - reduced azulLibV8.js -> azulNLibV1.js eliminated most md code

The program reads the index file and parses it to obtain all script files. It then imbeds the indexfile together with the script files for faster delivery.

## fasthttpServerNV2.go

todo 1: parse index file to retrieve js files

 - added jsHandler to handle requests for js files
 - test reading js files with fetch from browser

todo:
 1. parse index file to retrieve js files
 2. test expanding the class HtmlPage with code from azulLibxp
 3. test minimizer for the js code
 4. test compression
 5. test webworkers

# ws for fasthttp
## fasthttpServerWSV1.go

creating a websocket upgrader for fasthttp

succeeded at upgrading by inserting js into browser by hand

## fasthttpServerWSV2.go

upgraded websocket with downloaded js
send simple message back and forth

## fasthttpServerWSV3.go

building upgrader
will submit to fasthttp

test bed for ws and html/js to test ws communication

## fasthttpServerWSV4.go

built separate package upgrader

 - upgrade.Upgrade
 - ctxHijack
 - hijack handler

# embed
A program that reads an html index file and embeds the quoted script files.
The default output file name is the input file name with the added "_embed" text.

# Next Steps for Speed Optimization

## fasthttpServerNV1.go

 - test binary number transmission and its interpretation 
 - dataview
 - mixed numbers and text

## fasthttpServerNV2.go

more websocket testing:
 - ping and pong
 - multiple frames

