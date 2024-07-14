# fasthttp samples

server programs that demonstrate fasthttp methods

# azulserver

server that uses a minimal html file and uploads js code to build the dom

## fasthttpServer.go


## fasthttpServerV2.go


## fasthttpServerV3.go

This program is a simple server. Has no cli interpreter

## fasthttpServerV5.go

This program is a derivative of fasthttpServerV4, since that program has CLI.
./fasthttpServerV5 /port=<porttno> [/idx=indexfile] [/dbg]

The program reads the index file and parses it to obtain all script files. It then imbeds the indexfile together with the script files for faster delivery.

## fasthttpServerV6.go


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


# Next Steps for Speed Optimization

 - break-up azulLib.js into 2 files
 - test minimizer for the js code
 - test compression
