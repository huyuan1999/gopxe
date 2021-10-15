Golang 实现的 pxe 装机系统，类似于 Cobbler

## 目前实现
- TFTP
- HTTP
- 支持 CentOS7 和 RedHat7 系统安装
- 支持参数化模板文件

## 未来
- 实现 DHCP
- 实现装机状态管理
- 实现 Web 管理
- 支持更多类型的操作系统

## 项目改造计划
1. pxe server 下发（如果是第一次启动则启动bootos检查是否是物理机如果是则配置oob，bmc，raid并上报机器sn和网卡mac）
2. pxe server 接受到bootos客户端上报的信息，生成装机任务并加载default模版，根据模版生成装机菜单。一切准备完毕之后通知bootos客户端
3. bootos 客户端控制服务器重启，进入装机任务（default中必须添加inst.ks.sendmac和inst.ks.sendsn内核参数）


## 使用方式
```bash
# 安装和配置 dhcp 服务(这步后续会取消, 所有服务都将由 gopxe 内置)
$ yum -y install dhcpd
$ cat > /etc/dhcp/dhcpd.conf << EOF
option domain-name "test.com";
option domain-name-servers 223.5.5.5,223.6.6.6;

default-lease-time 600;
max-lease-time 7200;
authoritative;
ddns-update-style none;

subnet 10.0.0.0 netmask 255.0.0.0 {
        option routers                  10.0.0.1;
        option subnet-mask              255.0.0.0;
        option domain-search            "test.com";
        option domain-name-servers      223.5.5.5,223.6.6.6;
        option time-offset              -18000;     # Eastern Standard Time
        filename                        "pxelinux.0";
        range 10.1.0.1 10.1.0.10 ;  # reserved DHCPD range e.g. 10.17.224.100 10.17.224.150
}
EOF

$ systemctl enable --now dhcpd

# 启动服务, 并指定绑定的网卡
$ ./gopxe --device=eth0

# 复制 pxe 菜单文件到 tftp 目录
$ cp pxelinux/* /opt/GoPXE/tftp/

# 添加系统
$ curl -X POST -vv --form ks_param='{"rootpw": "666" }' --form version="CentOS7" --form name=CentOS7_minimal --form type=CentOS --form "iso=@/root/pxe/CentOS-7-x86_64-Minimal-2003.iso" \
http://10.1.1.1:8888/create/

$ curl -X POST -vv --form version="CentOS7" --form name=CentOS7_minimal_1 --form type=CentOS --form "iso=@/root/pxe/CentOS-7-x86_64-Minimal-2003.iso" http://10.1.1.1:8888/create/

$ curl -X POST -vv --form default=yes --form version="CentOS7" --form name=CentOS7_minimal_2 --form type=CentOS --form "iso=@/root/pxe/CentOS-7-x86_64-Minimal-2003.iso" \
http://10.1.1.1:8888/create/
```


## 参数说明
|  参数名称    | 是否必须  | 作用           | 默认值 | 可选值
|  ----       | ---- | ----               | ---- | ---- |
| type        |是    | 系统类型            | 无 |  CentOS 或者 RedHat 
| name       | 是   |  唯一标识            | 无 |  任意，但是不能重复，名称只能是字母大小写，数字以及下划线
| version   |  是   |  系统版本           | 无 |   CentOS7 或者 RedHat7
| default   |  否  |   是否为默认菜单选项  | no |   yes 或者 no
| iso       |  是  |  上传的进行文件      | 无 |   镜像文件
| ks_param  |  否 |   kickstart 参数     | 无 |   参考 ks 模板


## CentOS7 模板样例
```bash
auth --enableshadow --passalgo=sha512
text
keyboard --vckeymap=us --xlayouts='us'
lang en_US.UTF-8
reboot
zerombr
clearpart --all --initlabel
skipx

url --url={{ url|default: "http://mirror.centos.org/centos-7/7/os/x86_64/" }}

firewall {{ firewall|default: "--disabled" }}

firstboot {{ firstboot|default: "--enable" }}

{{ ignoredisk|default: "ignoredisk --only-use=sda" }}

{% if network %}
    {% for item in network %}
        {% if item.bootproto == "static" %}
            network --bootproto=static --ip={{ item.address }} --netmask={{ item.netmask }} --gateway={{ item.gateway|default: "0.0.0.0" }} --nameserver={{ item.dns|default: "223.5.5.5" }} --device={{ item.device }}
        {% else%}
            network --onboot=yes --bootproto=dhcp --device={{ item.device|default: "eth0" }}
        {% endif %}
    {% endfor %}
{% else %}
    network  --onboot=yes --bootproto=dhcp --device=eth0
{% endif %}

network  --hostname={{ hostname|default: "GoPXELinux" }}

rootpw --plaintext {{ rootpw|default: "123456" }}

selinux {{ selinux|default: "--disabled" }}

timezone {{ timezone|default:"Asia/Shanghai" }}

bootloader --append="quiet crashkernel=auto" --location=mbr --boot-drive={{ bootdrive|default: "sda" }}

{% if disk %}
    {% for item in disk %}
        part {{ item.mount }} {{ item.fstype|default: "xfs" }} --size={{ item.size }} --ondisk={{ item.ondisk|default: "/dev/sda" }} {{ item.opts }}
    {% endfor %}
{% else %}
    part /boot --fstype="xfs" --size=1024 --ondisk=/dev/sda
    part / --fstype="xfs" --grow --size=10240 --ondisk=/dev/sda
{% endif %}

# The remaining hook allows users to add their own configuration information
{{ HOOK }}

%pre
{{ pre }}
%end

%post
{{ post }}
%end

%addon com_redhat_kdump --enable --reserve-mb='auto'

%end
```
