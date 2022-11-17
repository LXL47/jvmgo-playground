# jvmgo  一个用go语言写的Java虚拟机

这里是 《自己动手写Java虚拟机》原书的源代码（加上了一些我自己的注释），还有jvmgo的虚拟机程序，可以在window直接运行class文件

------

如果想要学习这本书，最好先学习一下jvm，推荐 《深入了解Java虚拟机》或者尚硅谷的jvm：https://www.bilibili.com/video/BV1PJ411n7xZ/


### 我使用jvmgo建了一个网站，可以在尝试运行一下Java代码，体验一下这个虚拟机的功能：http://jvm.xpqly.love:82





【推荐先把全书的代码看懂，然后再动手一步步写代码，不然会很乱，写着写着不知道书里讲什么】

【jvm/ch11里面的代码是完整的jvmgo代码，和原书一样，加上了一些我的注释，可以对照着原书看】



------

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

