# stemcells

Tool to fetch all four BOSH stemcells (vSphere, Openstack, vCD/vCA and AWS) of a given version (e.g., 3026) from bosh.io into the current directory.

Example:
```
$ stemcells 3026
light-bosh-stemcell-3026-aws-xen-hvm-ubuntu-trusty-go_agent.tgz (17801 bytes, 349d6a2f8cfed5380420f2f11b6ecd7a)
bosh-stemcell-3026-vsphere-esxi-ubuntu-trusty-go_agent.tgz (557272899 bytes, 9f133fc05e10236846e4732bc1257f09)
bosh-stemcell-3026-vcloud-esxi-ubuntu-trusty-go_agent.tgz (557206399 bytes, 75ffe4270032b01e4f376f9c711ad6fd)
bosh-stemcell-3026-openstack-kvm-ubuntu-trusty-go_agent-raw.tgz (530329650 bytes, 585c0bbdec3bc620fd6c17a0faccc310)
```

## How to build
Nothing more than:
```
go install github.com/mgoelzer/stemcells
```


