---
title: 31-range-over-channels.go
date: 2020-01-10T20:03:35.697951+08:00
draft: false
hideLastModified: false
summaryImage: ""
keepImageRatio: true
tags:
- ""
- Go语言
- 样例代码
summary: 31-range-over-channels.go
showInMenu: false

---

//  使用for和range遍历基础数据结构,使用同样的句法遍历通道中的值

package main

import "fmt"

func main() {
	queue := make(chan string, 2) // 在queue通道中遍历这两这个值
	queue <- "one"
	queue <- "two"
	close(queue)

	//range遍历从queue通道中接收到的每一个元素
	for elem := range queue { // 因为在上面的代码中将通道关闭了，遍历将在获取到2个元素之后终止
		fmt.Println(elem)
	}
}

//  非空通道可以被关闭，被关闭后仍然可以接收剩余的值
