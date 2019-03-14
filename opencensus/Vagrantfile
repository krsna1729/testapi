# -*- mode: ruby -*-
# vi: set ft=ruby :

require 'fileutils'
require 'ipaddr'

SUPPORTED_OS = {
  "ubuntu" => {box: "generic/ubuntu1804", vm_prefix: "ubuntu", home_dir: "vagrant"},
}

# OS as the first variable so we can use it in other things.
$os             = (ENV['OS'] || "ubuntu")

# VM variables.
$num_instances  = (ENV['NODES'] || 3).to_i
$cpus           = (ENV['CPUS'] || 2).to_i
$memory         = (ENV['MEMORY'] || 4096).to_i
hosts           = {}
proxy_ip_list   = ""
$box            = SUPPORTED_OS[$os][:box]
$vm_name_prefix = SUPPORTED_OS[$os][:vm_prefix]
$home_dir       = SUPPORTED_OS[$os][:home_dir]

Vagrant.configure("2") do |config|
  config.vm.box = $box

  config.vm.synced_folder './', '/home/' + $home_dir + '/' + File.basename(Dir.getwd), type: 'rsync'

  (1..$num_instances).each do |i|
    vm_name = "%s-%02d" % [$vm_name_prefix, i]
    config.vm.define vm_name do |c|
      c.vm.hostname = vm_name
      c.vm.provider :libvirt do |lv|
        lv.cpu_mode   = "host-passthrough"
        lv.nested     = true
        lv.cpus       = $cpus
        lv.memory     = $memory
      end
      c.vm.provider :virtualbox do |vb|
        vb.cpus       = $cpus
        vb.memory     = $memory
      end
      if $box['generic/ubuntu']
        c.vm.provision "shell", privileged: true, path: "generic_ubuntu_hack.sh"
      end
      #c.vm.provision "shell", privileged: false, path: "setup_system.sh"
    end
  end
end
