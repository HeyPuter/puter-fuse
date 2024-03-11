
Cache for mkdir and unlink
```sh
mkdir mnt/user/three-files/another; ls mnt/user/three-files; rmdir mnt/user/three-files/another; ls mnt/user/three-files/; mkdir mnt/user/three-files/another; ls mnt/user/three-files/; sleep 5.1; echo "cache should be gone now"; ls mnt/user/three-files; rmdir mnt/user/three-files/another; ls mnt/user/three-files
```
