package os

import (
	"errors"
	"gopxe/utils"
	"path"
	"strings"
)

const CentOS7 = `
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
`

const CentOS6 = ""
const CentOS8 = ""

type CentOS struct {
}

func (c *CentOS) Mount(source, target, fstype string) error {
	return utils.Mount(source, target, fstype)
}

func (c *CentOS) Boot(target, mountPath string) error {
	if err := utils.CopyFile(path.Join(mountPath, "isolinux/vmlinuz"), path.Join(target, "vmlinuz")); err != nil {
		return err
	}
	return utils.CopyFile(path.Join(mountPath, "isolinux/initrd.img"), path.Join(target, "initrd.img"))
}

func (c *CentOS) Default() string {
	label := `LABEL {{ name }}
	{{ IS_DEFAULT }}
    MENU LABEL {{ name }} {{ label|default: "CentOS" }}
    KERNEL {{ kernel }}
    APPEND initrd={{ initrd }} ks={{ ks }}`
	return label
}

func (c *CentOS) Template(version string) (string, error) {
	switch strings.ToUpper(version) {
	case strings.ToUpper("CentOS6"), strings.ToUpper("RedHat6"):
		return CentOS6, nil
	case strings.ToUpper("CentOS7"), strings.ToUpper("RedHat7"):
		return CentOS7, nil
	case strings.ToUpper("CentOS8"), strings.ToUpper("RedHat8"):
		return CentOS8, nil
	}
	return "", errors.New("unsupported system version")
}
