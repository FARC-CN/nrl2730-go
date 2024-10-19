# nrl2730-go

An NRL2730 Server with Go

#### Looking for a Rust version? [Here](https://github.com/FARC-CN/nrl2730-rust) it is.

```
This program is used to forward UDP packets between clients, the first 20 bytes of the packet header are
    NRL2  4 bytes fixed     "NRL2"
    XX    2 bytes packet    length
    CPUID 7 bytes sending   device serial number
    CPUID 7 bytes receiving device serial number
```

### How to use：

Download, compile, and execute programs

```
# download
git clone https://github.com/FARC-CN/nrl2730-go

# compile
cd nrl2730-go
go build main.go

# run
./main
```
Note:

After the above program is running, it will receive and process data packets on UDP port 60050. If you need to use other ports, please specify -p [PORT], such as:

```
./main -p 6000
```

So that you know, the system's firewall must allow UDP port 60050 communication. Common system operations are as follows:

# Thanks
[BG6CQ](https://github.com/bg6cq), provided ideas for the code.

