# go-concurrent-downloader

This project is simple usage of concurrency in golang.
You can download your file faster than single thread apps.

## How to use
To use this project only install golang on your machine and run below command:

`go run main.go --url=url-your-interested-file --path=/path/filename --n=number-of-threads`

note that `--n` is optional.

exmaple:

`go run main.go --url=https://bitcoin.org/bitcoin.pdf --path=bitcoin-white-paper.pdf`