---
title: "Rabbitmq | 06 - RPC"
date: 2019-06-22T22:56:56+08:00
lastmod: 2019-06-23 22:56:56
draft: false
keywords: [mq,rabbitmq,go]
description: ""
tags: [mq,rabbitmq,golang]
categories: [tech]
author: "jayer"
---

<!-- 摘要 -->

这一节使用 RabbitMQ 构建 RPC 系统：包含一个客户端和一个可扩展的服务端，服务端是一个虚拟的 RPC Service，用来返回 Fibonacci 数

<!--more-->

有关 RPC 的说明：

尽管 RPC 在计算过程中是一个非常常见的模型，它也经常受到批评。当程序员并不注意一个函数是否是本地调用，或者是一个耗时的 RPC 时，问题就来了。像这样的迷惑就会导致产生不可预知的系统并且增加不必要的调试的复杂度。错误的使用 RPC 也将导致难以维护的意大利面式的代码产生。

> 牢记这一点，然后考虑下面的这些建议：
> 
> - 确保哪个函数明显是本地调用，哪个函数明显是远程调用
> - 系统设计要文档化，并且要使组件间的依赖清晰可见
> - 错误处理场景：当 RPC Server 宕机很长时间的时候客户端应该如何应对?

# Callback queue

一般来说，使用 RabbitMQ 处理 RPC 调用很容易：客户端发送请求消息，服务端回应响应消息。那么，为了能够收到响应消息，我们就需要在请求消息中携带一个 `回调队列的地址（callback queue address）`。我们使用默认的队列来尝试一下吧！

```go
q, err := ch.QueueDeclare(
    "",     // default queue name
    false,  // durable
    false,  // delete when usused
    true,   // exclusive
    false,  // noWait
    nil,    // arguments
)

err = ch.Publish(
    "",             // exchange
    "rpc_queue",    // routing key
    false,          // mandatory
    false,          // immediate
    amqp.Publishing {
        ContentType: "text/plain",
        CorrelationId: corrId,
        ReplyTo: q.Name,
        Body: []byte(strconv.Itoa(n)),
    }
)
```

> 消息属性说明；AMQP 0-9-1 协议预定义了 14 个消息的属性，但是大多数都不怎么使用，部分说明如下:
> 
> - `persistent`: 标记消息是否要被持久化(true | false)
> - `content_type`：消息编码类型(application/json | text/plain)
> - `reply_to`: 回调队列名
> - `correlation_id`：用来将 RPC 响应和请求关联

# Correlation Id

上面提到的为每个 RPC 请求创建一个回调队列效率并不高，但幸运的是这有一个更好的方式：为每个客户端创建一个 `回调队列(callback queue)`。

这就引出了一个新的问题，在回调队列中收到的响应并不能清晰的标明是来自那个请求的响应。这就要使用一个属性 `correlation_id`（关联id）。我们可以给每个请求都设置一个唯一的值，随后，我们在回调队列收到响应的时候就比对这个属性，基于此，我们就能将响应和请求对应起来。如果我们收到了一个未知的 `correlation_id`，说明该消息不是我们想要的，可以完全丢弃。

到这里，我们可能会有疑问：既然收到了不是来自我们请求的响应，为什么不直接报错，而是选择丢弃呢？那是因为在 Server 端存在竞争检测的可能。尽管不太可能，但 RPC server 也有可能在发送响应后就宕机了，而且在宕机前也没来得及发送 ack。如果这种情况发生了，重启 RPC 服务后会再次处理这个请求。这就是为什么在客户端上我们必须优雅的处理重复的响应，理想情况下 RPC 应该是幂等的。

# Summary

![](https://www.rabbitmq.com/img/tutorials/python-six.png)

上图中的 RPC 工作流：

- 当客户端启动的时候，会创建一个匿名独占的回调队列
- 一个 RPC 请求中，客户端发送的消息会带两个属性：`reply_to`：其值为回调队列，`correlation_id`：为每个请求设置的唯一值
- Request 将被发送到一个 RPC 队列
- RPC Server 等待来自 rpc_queue 队列的请求。当请求到达后，他就会处理其任务，并将处理结果通过 `reply_to` 字段所表示的回调队列送回给客户端。 
- 客户端等待回调队列的数据。当消息到达后，先检查 `correlatioin_id`，如果和请求中的值匹配，则将 Response 返回给应用程序。

完整 [Demo Code][#3] 

# See Also

> Thanks to the authors 🙂

* [RPC][#1]
  
[#1]:https://www.rabbitmq.com/tutorials/tutorial-six-go.html
[#3]:https://github.com/ijayer/mq-practice/tree/master/rabbitmq/golang-intro/06-rpc

# Content

- [01-hello world](../01-hello-world)
- [02-work-queues](../02-work-queues)
- [03-publish/subscribe](../03-publish-subscribe)
- [04-routing](../04-routing)
- [05-topics](../05-topics)
- [06-rpc](../06-rpc)
