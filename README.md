# Freedxm

Freedom Port for Linux. Block Websites, Apps, and the Internet.

## What is Freedom?

Official website: [https://freedom.to/](https://freedom.to/)

> Easily block distracting websites and apps on any device. The original and best website blocker, Freedom helps you be more focused and productive.

On any device excepting Linux!

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
