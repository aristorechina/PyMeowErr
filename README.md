# PyMeowErr

>  本项目旨在将Windods系统下Python程序的报错内容猫娘化喵（bushi



#### 用前须知

1. 本项目目前发现了一个作者没有能力修复的bug（应该没有其他bug了，要是有我也不会修）——如果在使用input函数时传入了prompt参数，那么终端会先要求用户输入再显示prompt的内容（本来应该是先输出prompt的内容再要求用户输入的）
2. 本项目仅供娱乐，后续大概率不会进行更新（因为作者懒）
3. 本项目使用了大语言模型辅助



#### 使用方法

1. 先安装好Python，把`python.exe`所在目录添加到环境变量中
2. 进入`python.exe`所在目录，把`python.exe`的文件名改成别的（我是改成了`python3.exe`，你也可以改成别的，但是改完以后要记住，这个后面会用到）
3. 到[Releases](https://github.com/aristorechina/PyMeowErr/releases)下载我编译好的`python.exe`文件（如果不放心的话你也可以自己编译）以及配置文件`config.toml`，把这两个文件移到刚刚提到的目录中（其实上面的操作就是为了防止文件名冲突）
4. 修改配置文件`config.toml`中的`python_path = ""`，填入Python 解释器路径，默认为 `python3`（如果你第2步不是改成`python3.exe`那么你这里就填你修改后的文件名，或者指定完整的路径，如 `C:\\Python39\\python.exe`）
5. 然后就没有然后啦，后续正常使用即可喵～



#### 如何自定义自己的报错语句？

在`config.toml`中使用正则表达式进行匹配

```toml
[[replacements]]
pattern = "^ErrType: (?P<name>.+)"             # 原本的报错内容（用正则表达式匹配）
replacement = "(ErrType) 报错内容是${name}喵～" # 修改后的内容
```

`pattern`变量存储的是原本报错输出的内容，这里填入的是正则表达式（要注意转义的问题）

`replacement`变量存储的是修改后报错输出的内容

在上面的例子中，假设有一个报错原本会输出`ErrType: Meow`，经过上面的修改后输出就会变为`(ErrType) 报错内容是Meow喵～`



#### 鸣谢

- 本项目灵感来源于[杂鱼♡～杂鱼♡～程序员都是杂鱼♡～](https://www.bilibili.com/video/BV1JL6fYuEZE/)，[【中文】杂~鱼♡！人家GCC也想变得可爱嘛～](https://www.bilibili.com/video/BV1gC4y1P7t3/)
