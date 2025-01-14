# 学习本作理念

代码的理念就是极简！这也是本项目名字由来！

根据 奥卡姆剃刀原理，不要搞一大堆复杂机制，最简单的能实现的代码就是最好的代码。

**想要为本作贡献的同学，要学习本作的这些理念，并能够贯彻你的代码。**

**不够极简或解释不够清晰的代码我们将会进行淘汰或修正。**

# 所有issue和 PR 尽量用中文

所有issue和 PR 尽量用中文

本项目不考虑应用 i18n. 这是本项目的强制约定。

# 所发的PR是有优先级的

随着项目不断扩大，一些对不同成分的优化会有不同的优先级

重要性按如下顺序排列：

1. 代理问题，如果直接就导致无法代理，或闪退，这是重大bug，必须立即在下一个beta版本修复
2. 协议是否有bug、程序是否有安全问题、内存泄漏问题等，这个也很重要。必须在下一个Patch版本修复
3. 原作功能补充
4. 新功能添加, 如果是添加新代理协议，一般要在Minor版本中加入。
5. 代码优化、代码结构性问题。这个因为结构性问题比较复杂，需要慢慢实现，慢慢改。如果是大范围结构性的改动，在下一个Minor版本中加入。如果是完整的架构修改，在下一个Major版本中加入。
6. 编译优化，这个是低优先级，而且也是很好处理的。
7. 安装教程、一键脚本、安卓客户端等。这个有时间再说，最低优先级。安装教程我也有一份 install.md ，不定时更新。

关于版本号的定义，我们参考但不完全遵循golang的定义：
https://go.dev/doc/modules/version-numbers

我们没有0.0.0版本，初始版本就从1.x.x开始
我们beta版本提供各种bug修复，以及功能调整与新增
Patch版本一般会比上一个Patch版本相比具有新增的功能
Minor版本具有显著的功能增加
Major版本具有显著的架构调整

对于一些人的PR，我会做出一些指导，有时并给予临时性修复。我的临时修改只是一种指导性含义，作为PR的原作者，你需要自己维护自己代码的质量，要理解我的临时修改不能当作最终修改，要自己想出最完善的修改。

我们每个人都是有日常生活的，能照顾本项目已经很不错了，不要指望什么都能想到，也不要指望不犯错 ，我们要互相包容，做到自己最好，多写代码，少进行没有意义的指责。

# github action

每push一次，进行一次test

每发布一个release，进行一次 build_release

build_release 会编译适当数量的 目标平台 可执行文件。

如需要更多平台的编译文件，可以手动执行 build_release_extra, 它会要求你输入一个tag，然后它会为该tag的代码 编译出额外平台的 可执行文件。

# 文档所用语言 - Language for Documentation

这里的文档 指的是 代码注释。根据go语言的标准，代码注释会生成代码文档，所以注释就是一种文档。

中文在作为文档的情况下，是与 英文 平级的，没有定性要求 一定要用哪个语言，但是有一个基本规则，如下。

## 基本原则

Package级别的 包的一句话描述 要用英文。 其他文档的规则如下：

### 如果你是汉语母语者

遵循下面规则, 简称 “极简规则”：

如果一句注释 用中文的 utf-8 所实际占用的字节数 会比 用英文字符 所占用的 字节数 要少，则建议用中文，否则 建议用英文。

备注，一个中文字符的 utf-8表示 占 3字节。

举个例子，比如 “实现了xxx协议”， 这个“实现”，不建议用 "implements", 因为英文有10字节，而中文才6字节。同样， protocol 与 “协议” 对比来说，英文占了 8 字节，还是多。

第二个例子，"统计数据" 有 12字节，而 "statistics" 才10字节，所以要用英文表达。

第三个例子，"可返回" 有9 字节，而 "can return" 刨除空格也是9字节，但是因为英文的空格是必须有的，所以还是 中文胜出。

第四个例子, "和", 英文是 "and", 没差别，都是3字节，所以都可以；但是 "或" 英文的 "or" 更简单，所以建议用 "or". 尤其如果要用 "或者“ 的话，那字节数就更多了，更加不建议用中文。

如果字节数相差不大，那么我们还可以进一步比较 音节数量 以及 书写长度。

#### 例外

如果一个文档比较复杂，则建议用中文，毕竟主要读者是华人，自己都不好用英语写，那华人读英语就更费劲了。

如果一个文档比较简单，建议直接用英文描述。

### If Chinese is Not Your First Language

Use English.

