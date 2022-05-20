# 概述

随着互联网上的图片，文本，视频信息日益膨胀，考虑到目前很多网民还是未成年人，因此检测其中的违规内容格外重要。本项目提供了部分违规内容的检测功能：违规图片检测（色情，恐暴）和违规文本检测（色情、辱骂、威胁等）。本项目不提供具体的数据和模型文件，由其它仓库提供模型的训练代码。

## 特点

* 深度学习模型支持docker方式部署
* Web后台和推理端之间引入消息队列和缓存技术
* 使用代理程序屏蔽推理端的动态变化
* Web后台提供面向客户端和管理员端的高可用HTTP接口
* 对于违规图片和违规文本拥有较高的分类准确率
* 支持敏感词过滤功能，包含内置的词库文件且支持用户自定义添加敏感词
* 不仅返回模型的推理结果，而且支持按照自定义的阻止策略二次处理该结果
* 系统整体进行了完备的功能测试和非功能测试

## 对比云商的内容安全服务

|               | 本系统                                             | 云商（百度云，阿里云等）           |
| ------------- | -------------------------------------------------- | ---------------------------------- |
| 违规内容类别  | 图片（色情、恐暴），文本（色情、辱骂、威胁等）     | **支持的检测任务更为丰富**         |
| 数据集/语料库 | 开源或自构建数据集，数量不多质量不佳               | **数据集/语料库理应更为丰富**      |
| 模型          | 因算力受限均选用的是经典模型                       | **模型理应更为复杂精度更高**       |
|               |                                                    |                                    |
| 推理返回结果  | **支持自定义的阻止策略对推理结果二次处理**         | 系统内置策略且不支持自定义         |
| 敏感词过滤    | **支持返回命中的敏感词和类别，且敏感词库能自定义** | 不返回命中的敏感词，词库不能自定义 |
| 超长输入文本  | **支持更长的输入文本，模型对长文本分类有优化**     | 直接截取过长的文本                 |

# 系统设计

## 整体架构设计

![image-20220520171740052](H:/0-%E6%AF%95%E4%B8%9A%E8%AE%BE%E8%AE%A1/godetector/doc/img/image-20220520171740052.png)

上图是本系统的整体架构设计图，主要分为Web后台和机器学习推理端，二者通过NSQ消息队列和Etcd数据库连接，系统用户有开发者和管理员，其通过HTTP接口与Web后台交互。Web后台利用HTTP接口层解析和分发用户请求，利用敏感词过滤模块实现对违规文本敏感词的快速过滤，利用自定义策略模块实现对推理端返回结果的自定义处理流程，利用NSQ生产者模块实现与消息队列和Etcd数据库的交互。

## 主活动图

![image-20220515120639258](H:/0-%E6%AF%95%E4%B8%9A%E8%AE%BE%E8%AE%A1/godetector/doc/img/image-20220515120639258.png)

首先，用户通过接口向Web后台发送检测请求，然后Web后台会判断待检测的内容的类型，根据是图片还是文本，该检测请求会被分发到不同的推理端，图片还需要由Web后台进行预处理。推理端在接收到调用请求后会执行模型推理的过程，主要有两条并行的违规文本检测任务，分别进行针对中文和英文的违规文本检测，两条并行的违规图片检测任务，分别进行色情图片检测和暴恐图片检测。并行执行能最大限度地减少响应时间，提高系统资源利用率。针对文本，Web后台还需要进行敏感词过滤。Web后台在等待推理结果的时候会进入同步阻塞状态，为了防止无效等待的情况，需要开启超时机制，只要达到超时时间就会返回当前结果。最后Web后台会根据管理员定义的阻止策略生成检测报告，并通过HTTP响应返回对用户请求的违规文本的相关处理建议

## 功能和接口说明

![image-20220520172045301](H:/0-%E6%AF%95%E4%B8%9A%E8%AE%BE%E8%AE%A1/godetector/doc/img/image-20220520172045301.png)

本系统的用户包括开发者和管理员，上图是本系统的用例图，仅供开发阶段调试的接口以及供管理员和开发者使用的接口可以分别查看下面两张表。

| 接口           | 类型 | 请求       | 响应         | 说明                   |
| -------------- | ---- | ---------- | ------------ | ---------------------- |
| /index/test    | GET  | -          | HTML页面     | 进入系统测试主页       |
| /image/nsfw    | POST | 图片字节流 | 检测结果JSON | 测试推理端色情图片检测 |
| /image/protest | POST | 图片字节流 | 检测结果JSON | 测试推理端暴恐图片检测 |
| /text/filter   | POST | 文本字符串 | 检测结果JSON | 测试Web后台敏感词过滤  |
| /text/cn       | POST | 文本字符串 | 检测结果JSON | 测试推理端违规中文检测 |
| /text/en       | POST | 文本字符串 | 检测结果JSON | 测试推理端违规英文检测 |

| 接口            | 类型 | 请求           | 响应         | 说明                                                       |
| --------------- | ---- | -------------- | ------------ | ---------------------------------------------------------- |
| /index/admin    | GET  | -              | HTML         | 进入管理员测试主页                                         |
| /nsq/image      | POST | 图片字节流     | 检测结果JSON | 开发者进行违规图片检测，包括色情和暴恐图片检测             |
| /nsq/text       | POST | 文本字符串     | 检测结果JSON | 开发者进行违规文本检测，包括敏感词过滤和中英文违规文本检测 |
| /admin/strategy | POST | 策略描述表达式 | -            | 管理员自定义阻止策略                                       |
| /test/health    | GET  | -              | 模型状态JSON | 管理员查看推理端模型的存活状态                             |

# 从零开始部署

* python环境准备：参照https://pytorch.org/serve/安装TorchServe，其余需要安装的包通过`pip install -r requirements.txt`安装，如果Pytorch安装的是GPU版本，可以使用下面的检测代码。如果这一步有些困难可以安装Docker即可，后面的模型我都封装到了Docker镜像中。

```python
import torch
torch.cuda.is_available()
import tensorflow as tf
# tf.test.is_gpu_available()
tf.config.list_physical_devices('GPU')
conda remove cudatoolkit --force
```

* 启动TorchServe模型的推理后端，若启动失败可以删掉logs文件夹再次尝试，另外，推理端也支持Docker方式启动，详见[README](./docker/README.md)

```shell
# 首先启动四个TorchServe模型的推理后端
torchserve --start --model-store model_store --models nsfw_model=nsfw_model.mar protest_model=protest_model.mar cn_text_model=cn_text_model.mar en_text_model=en_text_model.mar
torchserve --stop
# 也可以采用Docker方式启动四个推理端模型
docker run -itd --name torchserve_model -p9090:8080 -p8081:8081 -p8082:8082 -p7070:7070 -p7071:7071 torchserve_model:0.1
```

```shell
# 然后测试一下TorchServe是否正常运行
curl http://127.0.0.1:8080/predictions/nsfw_model -T ./test/test_img/porn.jpg
curl http://127.0.0.1:8080/predictions/protest_model -T ./test/test_img/protest_sign.jpg
curl -X POST http://127.0.0.1:8080/predictions/cn_text_model -T ./test/cn_text_test.txt
curl -X POST http://127.0.0.1:8080/predictions/en_text_model -T ./test/en_text_test.txt
```

* 通过如下命令启动NSQ消息队列和Etcd数据库

```shell
docker run -d --name Etcd-server  --publish 2379:2379  --publish 2380:2380  --env ALLOW_NONE_AUTHENTICATION=yes bitnami/etcd:latest

docker run  -itd --name nsqd -p 4150:4150 -p 4151:4151 nsqio/nsq /nsqd 
```

* 通过`go mod tidy && go build`构建WEB后台可执行文件，配置conf.yaml文件，即可启动后台WEB服务。
* 通过`cd server && go build`构建推理端代理的可执行文件，配置conf.yaml文件，即可启动推理端代理程序。
* 访问 http://127.0.0.1:8000/index/test 内部调试，访问 http://127.0.0.1:8000/index/admin 管理员测试主页。

# 模型&算法实现

这一节介绍系统构建或使用的四个深度学习模型，以及敏感词过滤算法，数学表达式执行引擎的实现

## 违规图片检测：

### 色情图片检测：

* 实现思路

```
- 首先获取部分色情图片数据集
- 然后将OPEN_NSFW模型的Caffe文件转换为PyTorch支持的格式
- 接着使用PyTorch复现ResNet50网络
- 最后加载转换后的模型权重并进行微调训练。
```

* 分类标签：

```
- porn：pornography images
- hentai：hentai images and pornographic drawings
- sexy：sexually explicit images, but not pornography
- neutral：safe for work neutral images
- drawings：safe for work drawings and anime
```

* 参考资料：

\- 数据来源：https://github.com/alex000kim/nsfw_data_scraper

\- Tensorflow2模型：https://github.com/GantMan/nsfw_model 

\- Pytorch模型：https://github.com/yangbisheng2009/nsfw-resnet.git

\- Yahoo开源模型：https://github.com/yahoo/open_nsfw

* 开发体会：

```
- 其他来源的数据集质量没有nsfw_data_scraper高，主要体现在sexy和pron的图片分不清楚。
- 对于正常标签的图片可以用ImageNet的数据集扩充。
- 追求使用EfficientV2这种新的模型或者数据增强的方法会浪费很多时间，效果也没多大提升。
- 保底的策略是把Yahoo的开源caffe模型转换成Pytorch模型，免去训练的耗时。
```

* TorchServe运行方式：

```shell
torchserve --start --model-store model_store --models nsfw_model=nsfw_model.mar
# 关闭方式
# torchserve --stop 
```

* Curl测试接口：

test/test_img 文件夹下面包含提供的测试图片，文件名为真实标签

```
curl http://127.0.0.1:8080/predictions/nsfw_model -T ./test/test_img/porn.jpg
```

运行结果：

```
{
  "porn": 0.6816046833992004,
  "sexy": 0.2803221046924591,
  "neutral": 0.03365671634674072,
  "hentai": 0.0038238749839365482,
  "drawings": 0.0005926391459070146
}
```

测试结论：

```
- 色情图片能够准确识别出来
- 类似比基尼的性感图片不会误判成色情图片
- 类似大卫雕像的图片不会误判成色情图片
```

### 抗议/暴力图片检测

* 实现思路

```
2017年UCLA的Won等人收集和整理了一份针对抗议和暴力集会图片的分类数据集，这份数据集共有40764张图片，其中有11659张抗议和暴力集会图片。论文中使用该数据集基于ResNet50神经网络训练了分类模型，并且取得了不错的效果。虽然论文没有提供训练好的模型文件，但是提供了其所用数据集的下载地址，本文初步浏览发现数据集质量很高，所以该数据集将会直接用于暴恐图片检测模型的训练。
```

* 分类标签

```
- protest：contains a binary value: 1 (protest) or 0 (not protest)
- violence：contains a real number in range between 0 (least violent) and 1 (most violent) 
- Other visual attributes are all binary: 1 (with the attribute), 0 (without the attribute)。Specifically, this includes: "sign", "photo", "fire", "police", "children", "group_20", "group_100", "flag", "night", "shouting".
- Please note that only protest images (images for which <protest> = 1) have violence and visual attribute annotations.
```

- 参考资料：

\- 论文：[Protest Activity Detection and Perceived Violence Estimation from Social Media Images](https://arxiv.org/abs/1709.06204)

\- 代码：https://github.com/wondonghyeon/protest-detection-violence-estimation.git

\- 数据：通过邮件的方式请求论文作者提供

- 开发体会：

```
- 调研阶段没有找到相关的公开数据集，于是这个任务直接用的Donghyeon Won等人的成果，万分感谢。
- 类似上面这种基于图片分类的恐暴检测方案粒度有点大，部分场景下检测效果不佳，有精力尝试基于目标检测技术实现
```

* TorchServe运行方式：

```shell
torchserve --start --model-store model_store --models protest_model=protest_model.mar
# 关闭方式
# torchserve --stop 
```

* Curl测试接口：

```shell
curl http://127.0.0.1:8080/predictions/protest_model -T ./test/test_img/protest_sign.jpg
```

运行结果：

```
{
  "sign": 0.17170631885528564,
  "protest": 0.17033971846103668,
  "children": 0.07608138769865036,
  "violence": 0.07127571851015091,
  "group_20": 0.06621399521827698
}
```

测试结论：

```
- protest_sign.jpg中的标语抗议被准确识别了出来
- protest_fire.png中的纵火抗议被准确识别了出来
```

## 违规文本检测

### 敏感词过滤

* 目标概述：

```
- 敏感词过滤指的是检测文本中存在的敏感词，比如色情类，谩骂类，政治类的敏感词，本节将会返回查询文本命中的敏感词列表以及对应的类别
- 该部分对于个人开发者而言最重要的是敏感词库的构建，我将会按照类别搜集整理网上的各种敏感词库，最终版在godetector\filter\sensitive-dicts目录中
- 多模式匹配的模型可以选择Trie树，AC自动机，双数组Trie树等，本文选择的是AC自动机，并用双数组Trie对它进行了优化，即在双数组Trie上加了FailLink
```

* 敏感词库构建：

```
- 尽力搜集网上的敏感词库并整理，最终的敏感词库单独分类：广告（120条），政治（1037条），暴恐（614条），民生（515条），网址（14595条），色情（560条），其他(14331条)。
- 敏感词库构建时采取的处理手段包括：不同来源的词库合并去重，敏感词繁简体转换，敏感词拆字（如‘奶’拆成‘女’和‘乃’），拼音替代（如‘操’转成‘wo cao’)。代码见sensitive-filter文件夹。
- 特别注意的是，本词库搜集了部分违规网站的网址和部分恶意的DNS，详细来源自项目https://github.com/stamparm/maltrail/，该项目在初次启动的时候会从各大安全网站下载目前最新的恶意网址和DNS信息，我们只需要在这里加上自己的导出逻辑即可。
```

* 文本匹配算法：


```
- 对于普通匹配算法，如果遍历查找时间复杂度是O(n^2)，用二分查找法时间复杂度是O(logn)，如果用TreeMap去匹配，时间复杂度是O(logn)，这里的n指的是词典的大小，如果用HashMap的话，时间复杂度是O(1)，但是空间复杂度又上去了，所以，想要找到一种速度又快，同时内存又省的数据结构，来完成这个匹配操作。
- 问题定义：给一个很长很长的母串 长度为n，然后给m个小的模式串，求这m个模式串里边有多少个是母串的字串。如果用KMP让每一模式串与母串进行匹配，这样时间复杂度为O((n + len(m))*m)，其实对这m个模式串建立一个DFA的话还可以更快。DFA的特征：有一个有限状态集合和一些从一个状态通向另一个状态的边，每条边上标记有一个符号，其中一个状态是初态，某些状态是终态。但不同于不确定的有限自动机，DFA中不会有从同一状态出发的两条边标志有相同的符号。
- 我们通过构造这m个模式串的前缀树，其实就是构造了一种DFA，如果要求一个母串包含哪些模式串，以用母串作为DFA的输入，在DFA上行走，走到终止节点，就意味着匹配了相应的模式串。其实，如果我们基于Trie树构造Trie图的话又能进一步提高速度。
- AC自动机是Trie图的一种构造方式，它是解决字符串多模匹配问题的利器，借鉴KMP中避免母串在匹配过程种指针回溯的方法的思想（通过next数组避免指针做不必要的前移），在Trie图中定义了前缀指针的概念，从根节点沿边到节点p我们可以得到一个字符串S，节点p的前缀指针定义为：指向树中出现过的S的最长的后缀。
- 构造前缀指针的步骤为：根据深度一一求出每一个节点的前缀指针。对于当前节点，设他的父节点与他的边上的字符为Ch，如果他的父节点的前缀指针所指向的节点的儿子中，有通过Ch字符指向的儿子，那么当前节点的前缀指针指向该儿子节点，否则通过当前节点的父节点的前缀指针所指向点的前缀指针，继续向上查找，直到到达根节点为止。再考虑一下它的时间复杂度，设m个模式串的总长度为LEN，所以算法总的时间复杂度为O(LEN + n)
- 预处理阶段需要应对常见的反过滤手段，比如无关字符干扰，拼音替代等，由于之前构建敏感词库的过程中已经做了部分数据增强，所以此时输入字符串检测时只需要采取去除无关字符/停用词的方法
```

### 中文违规文本检测

关于为什么要将中文和英文的违规文本检测分开实现，主要是因为二者的分词，预训练模型，数据集等方面确实存在很大的差异，实在没有办法用一个模型实现，遂分开

* 参考资料：

\- 数据集来源：翻译自Kaggle提供的类似的英文数据集（[Toxic Comment Classification Challenge | Kaggle](https://www.kaggle.com/c/jigsaw-toxic-comment-classification-challenge)），使用各大云商的翻译API接口翻译了1.5万条左右（主要是因为没钱）

\- 代码参考：预训练语言模型最经典的库https://github.com/huggingface/transformers.git，主要代码修改自Kaggle的有害评论分类任务https://github.com/unitaryai/detoxify.git

* TorchServe运行方式：

```shell
torchserve --start --model-store model_store --models cn_text_model=cn_text_model.mar
# 关闭方式
# torchserve --stop 
```

* Curl测试接口：

./test/cn_text_test.txt包含提供的测试文本，含有违规文本

```
curl -X POST http://127.0.0.1:8080/predictions/cn_text_model -T ./test/cn_text_test.txt
```

运行结果：

```
0.9769131
```

该值对应的是违规文本的概率，一般而言大于0.7即可认为是有害文本

* Bug解决记录：

该开始的时候我非常困惑curl发送到后台的文本都是乱码，最终找到问题原因：不是因为curl请求时没有将数据正确编码，只是因为python的logging框架控制台打印只支持ASCII编码，定位这个问题确实花费了我不少精力，解决方案就是添加编码格式为utf-8的FileHandler，查看log文件发现后台正常接受了中文文本。

```python
# 解决了logging打印中文乱码的问题
import logging
root_logger= logging.getLogger()
root_logger.setLevel(logging.DEBUG) # or whatever
handler = logging.FileHandler('d:/test.log', 'w', 'utf-8') # or whatever
formatter = logging.Formatter('%(name)s %(message)s') # or whatever
handler.setFormatter(formatter) # Pass handler as a parameter, not assign
root_logger.addHandler(handler)
root_logger.info(sentences)
```

### 英文违规文本检测

* 分类标签：

```
- toxicity: 有害的
- severe_toxicity: 极度有害的
- obscene：淫秽的
- threat：威胁语气
- insult：侮辱语气
- identity_attack：身份歧视（性别，种族等）
```

* 参考资料：

\- 数据集来源：Kaggle的有害评论分类任务提供的英文数据集（[Toxic Comment Classification Challenge | Kaggle](https://www.kaggle.com/c/jigsaw-toxic-comment-classification-challenge)、[Jigsaw Unintended Bias in Toxicity Classification](https://www.kaggle.com/c/jigsaw-unintended-bias-in-toxicity-classification)、[Jigsaw Multilingual Toxic Comment Classification](https://www.kaggle.com/c/jigsaw-multilingual-toxic-comment-classification)）

\- 代码参考：预训练语言模型最经典的库https://github.com/huggingface/transformers.git，主要代码修改自Kaggle的有害评论分类任务https://github.com/unitaryai/detoxify.git

* TorchServe运行方式：

第一次启动时会联网下载一些 huggingface transformers 的配置文件，请确保网络畅通

```shell
torchserve --start --model-store model_store --models en_text_model=en_text_model.mar
# 关闭方式
# torchserve --stop 
```

* Curl测试接口：

./test/en_text_test.txt包含提供的测试文本，含有违规文本，其真实标签是toxicity和obscene

```
curl -X POST http://127.0.0.1:8080/predictions/en_text_model -T ./test/en_text_test.txt
```

运行结果：

```
{'toxicity': 0.990424, 'severe_toxicity': 0.07567663, 'obscene': 0.9390912, 'threat': 0.0045800083, 'insult': 0.84097433, 'identity_attack': 0.00858633}
```

该值对应的是违规文本所属每一类的概率，虽然误分类了insult标签，但是无伤大雅，可见模型效果很好。

## 阻止策略设计

* 针对图片检测（阈值0.12可调）

```
- 在色情图片检测中概率最大的标签是porn或者hentai时，判定为色情违规图片
- 在抗议/暴力图片检测中，由于此任务的标签有12个之多，所以每一个标签的softmax概率值不会很大，此处如果protest标签的概率值超过0.12且有另一个标签的概率值超过0.12则可以判定抗议/暴力图片违规
- 上面两个子任务任意有一个出现违规则最终给出阻止当前内容的建议
```

* 针对文本过滤（权重weight可调）（参考一下TF-IDF你还要考虑文本长短）

```
- 在敏感词匹配中，有广告，政治，暴恐，民生，网址，色情，其他共7类敏感词，考虑到敏感词库的质量原因，不便于出现敏感词就阻止，设第i类敏感词库中匹配到的敏感词数量为num(i)，第i类的重要性记作weight(i)，i属于0-6，此处给出加权的敏感词数量S的计算公式，如果该值大于等于2则判断为违规文本
```

$S=\sum_{i=0}^6weight(i)*num(i)$ 其中 $\sum_{i=0}^6weight(i)=7$

* 针对文本检测（阈值0.7可调）

```
- 如果用户请求文本中含有中文则进行中文违规文本检测，返回的值大于0.7则有相当的把握认为是有害文本
- 如果用户请求文本中含有英文则进行英文违规文本检测，返回结果中toxicity、severe_toxicity、obscene、threat、insult、identity_attack这6个标签任意有一个概率值大于0.7则有相当的把握认为是有害文本
- 如果敏感词匹配任务或者NLP违规文本检测任务判断输入文本违规，那么最终给出阻止当前内容的建议
```

# 系统测试

在推理端依次部署1组、2组和3组违规文本检测模型记为Group1、Group2和Group3，一个违规文本检测模型组包括一个中文违规文本检测模型和一个英文违规文本检测模型。然后分别测量每个组的基本性能指标。测量性能指标的具体方法是，使用本机向Web后台发起100个连接，并在15秒内最大限度地发起违规文本检测的POST请求，请求体中包含了不同的违规文本。下表是3组测试的基本情况的对比，其中总请求数指的是在15秒内本机能最大限度发起的HTTP请求数量，平均延迟指的是所有请求延迟的平均值，成功率指的是HTTP请求成功的概率。

|        | 总请求数 | 吞吐量 | 平均延迟 | 成功率 |
| ------ | -------- | ------ | -------- | ------ |
| Group1 | 1314次   | 86QPS  | 343ms    | 100%   |
| Group2 | 1548次   | 92QPS  | 311ms    | 100%   |
| Group3 | 1763次   | 116QPS | 255ms    | 100%   |
