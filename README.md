# Freedxm

Freedom Port for Linux. Block Websites, Apps, and the Internet.

## What is Freedom?

Official website: [https://freedom.to/](https://freedom.to/)

> Easily block distracting websites and apps on any device. The original and best website blocker, Freedom helps you be more focused and productive.

The Freedom application, known for blocking distracting apps and websites to enhance productivity, was not available for Linux users. The challenge was taken on to port the Freedom application to Linux, enabling its functionality on this platform. This involved devising native methods to effectively block applications and websites on Linux, such as using Chrome's remote debugging mode and process termination techniques. As a result, Linux users can now benefit from increased productivity by reducing distractions through this newly ported application.

## How it Works

Both the program and this document are under construction.

Mechanism:

* Capable of blocking both processes (by specifying their names) and web domains.
* Both process names and domains can include a wildcard `%` for partial matching.
* Utilizes polling (if you think polling is inefficient, consider how inefficient you are with all those distractions anyway lol).
* Requires Chrome to operate in debugging mode. If a session is active and Chrome isn't in debugging mode, it will be terminated.
* Domains are blocked using Chrome's remote debugging mode, avoiding direct network interaction (as it's challenging to do reliably).
* Other browsers must be included in the process block list to ensure they are blocked.
* Future plans include controlling other browsers via remote debugging as well.
* Terminates processes using `killall` with the `-TERM` signal initially, and escalates to `-9` if the process does not terminate.

## Development

```sh
go test ./...
revive --formatter stylish ./...
```
