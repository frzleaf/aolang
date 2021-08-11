# AO LÀNG

Chơi LAN game qua internet

![alt text](doc/client_screen1.JPG "Title")


### I. Cài đặt

### 1. Server

#### Tự động tải file & chạy ở port: 9999

```sh -c "$(wget -O- https://raw.githubusercontent.com/frzleaf/aolang/master/server.sh)"```

_Lưu ý: cách này chỉ áp dụng cho server linux_


#### Tự build

```
git clone https://github.com/frzleaf/aolang
cd aolang
export GOPATH=$GOPATH:$PWD
# set GOPATH=%GOPATH%;%cd%         # for Windows environment
go build -o build/aolang_server src/main/proxy_server.go
```


### 2. Client

Vào mục [release](https://github.com/frzleaf/aolang/releases) để tìm bản hỗ trợ cho máy của bạn
hoặc build từ file ```src/main/proxy_client.go```


## II. Các game hỗ trợ

### 1. Warcraft
Version đã thử nghiệm: 1.24e

### 2. AOE (sắp tới)
