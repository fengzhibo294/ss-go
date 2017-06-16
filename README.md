# ss-go

the go language version of SS (Conference's message forwarding server)<br>


## about   program structure<br>
*package config:<br>
    -define  util  variable and comm funct<br>
*package sslog:<br>   
    -potting  std  log library<br>  
*package socket: <br>
    -most import architecture code<br>
*package  db2mysql:<br>
    -potting mysql's api<br>
*package module: <br>
    -proxy forwarding module implementation<br>
*ss_server.go: <br>
    -the entrance of ss-go<br>

## build and run<br>
*go build ss_server.go   <br>
-will create a  executable file named ss_server  at this current directory of centos<br>
*go run ss_server.go<br>
-run ss-go program as foreground<br>
