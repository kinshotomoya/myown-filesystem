# myown-filesystem

local file system

## motivation

you can access to cloud file storage such as aws s3, gcp gcs, azure blob 
as if you ran linux command like `ls` `touch`

## TODO:
- [ ] sample code (access memory file directory)
- [ ] docker set up
- [ ] fix code to access localstack
- [ ] fix code to run other linux command `rm` `touch` `mv` `tree`
- [ ] fix code to access not only localstack but also other cloud storage
- [ ] fix code to umount directory when kill go process 


## how to develop in local

do umount after kill go process 
```shell
$ umount /tmp/myown-filesystem
```

if you want to confirm what filesystem are mounted, you run following command
```shell
$ mount
```