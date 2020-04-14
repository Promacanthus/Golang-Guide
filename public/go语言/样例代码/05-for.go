---
title: 05-for.go
date: 2019-11-25T11:15:47.530182+08:00
draft: false
hideLastModified: false
summaryImage: ""
keepImageRatio: true
tags:
- ""
- Go语言
- 样例代码
summary: 05-for.go
showInMenu: false

---

// for是Golang中唯一的循环结构，for循环有三种基本形式

package main

import "fmt"

func main() {
	i := 1

	// 最基本的类型,具有单一条件
	for i <= 3 {
		fmt.Println(i)
		i = i + 1
	}

	// 经典的初始化/条件/下一步循环
	for j := 7; j < 9; j++ {
		fmt.Println(j)
	}

	// 无条件的循环将会一直执行直到break循环
	// 或者从封闭的函数中return
	for {
		fmt.Println("loop")
		break
	}

	// 使用contine继续循环的下一次迭代
	for n := 0; n <= 5; n++ {
		if n%2 == 0 {
			continue
		}
		fmt.Println(n)
	}
}
