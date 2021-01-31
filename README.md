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


## 使用方式
```bash
# 启动服务, 并指定绑定的网卡
$ ./gopxe --device=eth0

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
