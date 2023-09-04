# wrapper for chisel server for ease-of-use in code-server

## Features
* Runs server as libary, configured for code-server port-forwarding
* Listens for initial "Open in Browser" connection and returns configured `chisel client` command with cookie header

## Usage

### In code-server Terminal
1. Run local command that needs to be forwarded

    Ex: netcat listening on UDP 9999
    ```sh
    netcat -l -k -u localhost 9999  # or any command
    ```

1. Build and run with tunnel port

    ```sh
    go build \
        -ldflags="-s -w -X github.com/jpillora/chisel/share.BuildVersion=1.9.1" \
        -o code-server-chisel \
        github.com/micahyoung/code-server-chisel

    ./code-server-chisel server -p 8082
    ```

1. Click "Open in Browser" and copy command to clipboard

### On client workstation terminal
1. Paste the copied command, replacing the `# add local port` comment with the server-side local command port

    Ex: pasted command with "9999/udp" appended
    ```sh
    chisel client --header 'Cookie: code-server-session=...' https://localhost:8080/proxy/8082 9999/udp
    ```

1. Run the local client command to send data
    ```sh
    echo "hello" | nc -u localhost 9999
    ```
