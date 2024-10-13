package main

import (
    "fmt"
    "io"
    "log"
    "net"
    "os"
)

const (
    HOST = "0.0.0.0"
    PORT = "8085"
    TYPE = "tcp"
)

func main() {
    listen, err := net.Listen(TYPE, HOST+":"+PORT)
    if err != nil {
        log.Fatal(err)
        os.Exit(1)
    }
    defer listen.Close()

    log.Printf("Proxy server is listening on %s:%s...\n", HOST, PORT)

    for {
        conn, err := listen.Accept()
        if err != nil {
            log.Fatal(err)
            os.Exit(1)
        }
        go handleRequest(conn)
    }
}

func handleRequest(clientConn net.Conn) {
    defer clientConn.Close()

    buf := make([]byte, 4096)
    _, err := clientConn.Read(buf)
    if err != nil {
        log.Printf("Error reading: %v", err)
        return
    }

    if buf[0] != 0x05 {
        log.Println("Unsupported SOCKS version")
        return
    }

    // Ответ о поддерживаемых методах аутентификации (без аутентификации)
    _, err = clientConn.Write([]byte{0x05, 0x00})
    if err != nil {
        log.Printf("Error writing method response: %v", err)
        return
    }

    // Получение желания соединения от клиента
    _, err = clientConn.Read(buf)
    if err != nil {
        log.Printf("Error reading connection request: %v", err)
        return
    }

    if buf[0] != 0x05 {
        log.Println("Unsupported SOCKS version")
        return
    }

    // Получение адреса и порта целевого сервера
    addr := net.IP(buf[4:8]).String()
    port := (int(buf[8]) << 8) | int(buf[9])
    
    targetConn, err := net.Dial("tcp", net.JoinHostPort(addr, fmt.Sprintf("%d", port)))
    if err != nil {
        log.Printf("Error connecting to target: %v", err)
        return
    }
    defer targetConn.Close()

    // Ответ клиенту о разрешении соединения
    _, err = clientConn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
    if err != nil {
        log.Printf("Error writing connection response: %v", err)
        return
    }

    go copyData(targetConn, clientConn)
    copyData(clientConn, targetConn)
}

func copyData(dst net.Conn, src net.Conn) {
    _, err := io.Copy(dst, src)
    if err != nil {
        log.Printf("Error copying data: %v", err)
    }
}
