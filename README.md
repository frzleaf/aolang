# Lancraft

Chơi LAN game qua internet

![alt text](doc/client_screen1.JPG "Title")


### I. Cài đặt

### 1. Server

Nếu bạn chưa có server thì có thể tải [file binary](https://github.com/frzleaf/lancraft/releases)
hoặc clone source code về để build như sau:

*Linux*
```
git clone https://github.com/frzleaf/lancraft
cd lancraft
export GOPATH=$GOPATH:$PWD
go build -o build/lancraft_server src/main/proxy_server.go
```

*Windows*
```
git clone https://github.com/frzleaf/lancraft
cd lancraft
set GOPATH=%GOPATH%;%cd%
go build -o build/lancraft_server.exe src/main/proxy_server.go
```
Chạy máy chủ lancraft:
```
// Tạo TCP server với địa chỉ: mylancraft.com:9999
./build/lancraft_server mylancraft.com:9999 // linux
./build/lancraft_server.exe mylancraft.com:9999 // windows
```

### 2. Client

Vào mục [release](https://github.com/frzleaf/lancraft/releases) để tìm bản hỗ trợ cho máy của bạn
hoặc build từ file ```src/main/proxy_client.go```


## II. Các game hỗ trợ
### 1. Warcraft
Các phiên bản warcraft đã thử nghiệm: 1.24e
### 2. AOE (sắp tới)
