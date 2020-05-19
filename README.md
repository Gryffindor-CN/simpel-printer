# 打印盒子

### 
### logs 日志采集

portWayAgent目录:  
/usr/local/print/portway_arm

启动脚本  
/usr/local/print/start.sh
```shell
cd /path
nohup ./simple-printer & 
```

开机自启动:
在/etc/rc.local中添加
```shell
/usr/local/print/start.sh
```
