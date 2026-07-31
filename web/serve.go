// ローカルで dist/ を配信して WASM 版を動作確認するための開発用サーバ。
//
//	make serve-wasm
package main

import (
	"flag"
	"log"
	"net"
	"net/http"
)

func main() {
	dir := flag.String("dir", "dist", "配信するディレクトリ")
	addr := flag.String("addr", "localhost:8080", "リッスンアドレス (使用中なら空きポートに自動フォールバック)")
	flag.Parse()

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		// 指定ポートが使用中などの場合は空きポートに任せる
		host, _, splitErr := net.SplitHostPort(*addr)
		if splitErr != nil || host == "" {
			host = "localhost"
		}
		ln, err = net.Listen("tcp", net.JoinHostPort(host, "0"))
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Printf("http://%s/ で %s を配信します", ln.Addr(), *dir)
	log.Fatal(http.Serve(ln, http.FileServer(http.Dir(*dir))))
}
