# Freedxm

Freedom Port for Linux. Block Websites, Apps, and the Internet.

## What is Freedxm?

This software is a Linux adaptation of [Freedom](https://freedom.to/):

> Easily block distracting websites and apps on any device. The original and best website blocker, Freedom helps you be more focused and productive.

The Freedom application, known for blocking distracting apps and websites to enhance productivity, was not available for Linux users. The challenge was taken on to port the Freedom application to Linux, enabling its functionality on this platform. This involved devising native methods to effectively block applications and websites on Linux, such as using Chrome's remote debugging mode and process termination techniques. As a result, Linux users can now benefit from increased productivity by reducing distractions through this ported application.

## Getting Started

Create the file `~/.config/freedxm.toml` with content like this:

```toml
[options]
log-date-time = true

[notification] # Optional
normal = ["i3-nagbar", "-t", "warning", "-m", "%title: %message"]
warning = ["dunstify", "-u", "critical", "%title", "%message", "--timeout", "2000"]

[[blocklist]]
name = "socialsites"
domains = ["facebook", "x.com", "youtube.com"]
processes = ["firefox", "opera"]

[[blocklist]]
name = "work"
domains = ["twitch.tv", "google.com", "chatgpt.com"]
processes = ["insomnia", "vlc"]
```

Compile and install the executable, making sure it is accessible by adding it to your `PATH` environment variable:

```sh
go install
```

Run the server. For an enhanced experience, consider using your preferred process manager, such as `systemd`:

```sh
freedxm serve
```

Once the server is running, you can execute the following commands to start blocking:

```sh
freedxm new -m 40 -b socialsites,work

# or

freedxm new -m 10 -b work
```

For more information, refer to the help command:

```sh
freedxm help
```

## How it Works

### Blocking Mechanism

- **Process and Domain Blocking**: Capable of blocking both processes and web domains. Process names and domains can include a wildcard `%` for partial matching.
- **Process Termination**: Utilizes `killall` with the `-TERM` signal initially, escalating to `-9` if the process does not terminate.

### Chrome Integration

- **WebSocket Control**: Chrome tabs/pages are controlled using WebSockets, eliminating the need for polling.
- **Remote Debugging**: Requires Chrome to operate in debugging mode. If a session is active and Chrome isn't in debugging mode, it will be terminated.
- **Domain Blocking**: Domains are blocked using Chrome's remote debugging mode, which allows for domain blocking without the need for direct network interaction such as using proxies or iptables.

### Process Management

- **Polling for Processes**: Uses polling for Linux processes to monitor and manage them effectively.
- **Browser Control**: Other browsers must be included in the process block list to ensure they are blocked. Future plans include controlling other browsers via remote debugging.

### Notification System Integration

- **Configurable Notifications**: Integrates with any notification system by configuring the TOML file as shown in the example above. Notifications can be customized using commands like `i3-nagbar` and `dunstify`, allowing users to receive alerts based on their preferences and system setup.

## Development

```sh
go test ./...
revive --formatter stylish ./...
protoc --go_out=. --go-grpc_out=. ./rpc/server.proto
```
