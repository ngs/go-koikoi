// ローカルで dist/ を配信して WASM 版を動作確認するための開発用サーバ。
//
//	make serve-wasm
package main

import (
	"errors"
	"flag"
	"log"
	"net"
	"net/http"
	"syscall"
	"time"
)

func main() {
	dir := flag.String("dir", "dist", "配信するディレクトリ")
	addr := flag.String("addr", "localhost:8080", "リッスンアドレス (使用中なら空きポートに自動フォールバック)")
	flag.Parse()

	ln, err := net.Listen("tcp", *addr)
	if errors.Is(err, syscall.EADDRINUSE) {
		// 指定ポートが使用中の場合だけ空きポートへフォールバックする
		host, _, splitErr := net.SplitHostPort(*addr)
		if splitErr != nil || host == "" {
			host = "localhost"
		}
		log.Printf("%s は使用中のため空きポートを探します", *addr)
		ln, err = net.Listen("tcp", net.JoinHostPort(host, "0"))
	}
	if err != nil {
		log.Fatal(err)
	}

	fs := http.FileServer(http.Dir(*dir))
	srv := &http.Server{
		// 再ビルドした main.wasm がキャッシュから返らないようにする
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-store")
			fs.ServeHTTP(w, r)
		}),
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Printf("http://%s/ で %s を配信します", ln.Addr(), *dir)
	log.Fatal(srv.Serve(ln))
}
