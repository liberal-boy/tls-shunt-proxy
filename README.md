# TLS 分流隧道
用于分流 TLS 流量，支持按 sni 分流，分流 http 和无特征流量。并自带静态网站服务器，适用于 vmess + TLS + Web 方案实现，并可以与 trojan 共享端口。

## 下载安装
对于 linux-amd64 可以使用脚本安装，以 root 身份执行以下命令
```shell script
bash <(curl -L -s https://raw.githubusercontent.com/liberal-boy/tls-shunt-proxy/master/dist/install.sh)
```
* 配置文件位于 `/etc/tls-shunt-proxy/config.yaml`

* 其他平台需自行编译安装

## 使用
命令行参数：
```
  -config string
        Path to config file (default "./config.yaml")
```

<details>
  <summary>点击此处展开示例配置文件</summary>
  
```yml
# listen: 监听地址
listen: 0.0.0.0:443

# vhosts: 按照按照 tls sni 扩展划分为多个虚拟 host
vhosts:

    # name 对应 tls sni 扩展的 server name
  - name: vmess.example.com

    # tlsoffloading: 解开 tls，true 为解开，解开后可以识别 http 流量，适用于 vmess over tls 和 http over tls (https) 分流等
    tlsoffloading: true

    # cert: tls 证书路径
    cert: /etc/ssl/vmess.example.com.pem

    # key: tls 私钥路径
    key: /etc/ssl/vmess.example.com.key

    # http: 识别出的 http 流量的处理方式
    http:

      # handler: fileServer 将服务一个静态网站
      handler: fileServer

      # args: 静态网站的文件路径
      args: /var/www/html

    # default: 其他流量处理方式
    default:

      # handler: proxyPass 将流量转发至另一个地址
      handler: proxyPass

      # args: 转发的目标地址
      args: 127.0.0.1:40001

      # args: 也可以使用 domain socket
      # args: unix:/path/to/ds/file

  - name: trojan.example.com

    # tlsoffloading: 解开 tls，false 为不解开，直接处理 tls 流量，适用于 trojan-gfw 等
    tlsOffloading: false

    # default: 关闭 tlsoffloading 时，目前没有识别方法，均按其他流量处理
    default:
      handler: proxyPass
      args: 127.0.0.1:8443
```
</details>