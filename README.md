# jvmgo  一个用go语言写的Java虚拟机

这里是 《自己动手写Java虚拟机》原书的源代码（加上了一些我自己的注释），还有jvmgo的虚拟机程序，可以在window直接运行class文件



我使用jvmgo建了一个网站，可以在尝试运行一下Java代码，体验一下这个虚拟机的功能：http://jvm.xpqly.love/

------



jvmgo.exe的使用：

在jvmgo所在目录下，放一个Java文件，然后在windows命令行输入以下命令，编译成class文件：

(Main.java是文件名)

```
javac -encoding utf8 Main.java
```



然后输入以下命令，运行：

（"E:\soft\Java\jdk1.8.0_271\jre" 是你的Jre目录，有双引号）

（Main是 Main.java里面的类名，没有双引号）

```cmd
jvmgo -Xjre "E:\soft\Java\jdk1.8.0_271\jre" Main
```



------

如果想编译源代码为可执行文件，输入以下命令：(需要已经安装好go)

【可执行文件生成在 go的工作目录下的bin文件夹  ，如：D:\go\workspace\bin】

```
go install jvmgo/ch11
```

