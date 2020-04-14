---
title: 13-multiple-return-values.go
date: 2019-11-25T11:15:47.530182+08:00
draft: false
hideLastModified: false
summaryImage: ""
keepImageRatio: true
tags:
- ""
- Go语言
- 样例代码
summary: 13-multiple-return-values.go
showInMenu: false

---

//  Go内置支持多个返回值，这个特性通常用在惯用的Go中，如从函数中返回结果和错误值

package main

import "fmt"

func vals() (int, int) { // 函数签名中的（int,int），表示这个函数返回两个int
	return 3, 7
}

func main() {
	a, b := vals()
	fmt.Println(a)
	fmt.Println(b)

	_, c := vals() // 如果只需要返回值中的一个子集，可以使用空白标识符_
	fmt.Println(c)
}
