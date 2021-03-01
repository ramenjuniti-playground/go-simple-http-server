package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/textproto"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("%+v", err)
	}
}

func run() error {
	fmt.Println("start tcp listen...")

	listen, err := net.Listen("tcp", "localhost:12345")
	if err != nil {
		return errors.WithStack(err)
	}
	defer listen.Close()

	conn, err := listen.Accept()
	if err != nil {
		return errors.WithStack(err)
	}
	defer conn.Close()

	fmt.Println(">>> start")

	reader := bufio.NewReader(conn)
	scanner := textproto.NewReader(reader)

	var method, path string
	header := make(map[string]string)

	isFirst := true
	for {
		line, err := scanner.ReadLine()
		if line == "" {
			break
		}
		if err != nil {
			return errors.WithStack(err)
		}

		if isFirst {
			isFirst = false
			headerLine := strings.Fields(line)
			header["Method"] = headerLine[0]
			header["Path"] = headerLine[1]
			fmt.Println(header["Method"], header["Path"])
			continue
		}

		headerFields := strings.SplitN(line, ": ", 2)
		fmt.Printf("%s: %s\n", headerFields[0], headerFields[1])
		header[headerFields[0]] = headerFields[1]
	}

	method, ok := header["Method"]
	if !ok {
		return errors.New("no method found")
	}
	if method == "POST" || method == "PUT" {
		len, err := strconv.Atoi(header["Content-Length"])
		if err != nil {
			return errors.WithStack(err)
		}
		buf := make([]byte, len)
		_, err = io.ReadFull(reader, buf)
		if err != nil {
			return errors.WithStack(err)
		}
		fmt.Println("Body:", string(buf))
	}

	var resp []byte
	if method == "GET" {
		path, ok = header["Path"]
		if !ok {
			return errors.New("no path found")
		}
		cwd, err := os.Getwd()
		if err != nil {
			return errors.WithStack(err)
		}
		p := filepath.Join(cwd, filepath.Clean(path))
		if err != nil {
			return errors.WithStack(err)
		}
		resp, err = ioutil.ReadFile(p)
	}

	io.WriteString(conn, "HTTP/1.1 200 OK\r\n")
	io.WriteString(conn, "Content-Type: text/html\r\n")
	io.WriteString(conn, "\r\n")
	io.WriteString(conn, string(resp))

	fmt.Println("<<< end")

	return nil
}
