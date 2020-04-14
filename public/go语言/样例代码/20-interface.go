---
title: 20-interface.go
date: 2019-11-25T11:15:47.530182+08:00
draft: false
hideLastModified: false
summaryImage: ""
keepImageRatio: true
tags:
- ""
- Go语言
- 样例代码
summary: 20-interface.go
showInMenu: false

---

//  接口是方法签名的命名集合

package main

import (
	"fmt"
	"math"
)

type geometry interface { // geometry接口
	area() float64
	perim() float64
}

// rect类型和circle类型都将实现geometry接口
type rect struct {
	width, height float64
}

type circle struct {
	radius float64
}

// 在Go中要实现一个接口只需要实现接口中全部的方法即可
func (r rect) area() float64 {
	return r.width * r.height
}

func (r rect) perim() float64 {
	return 2*r.width + 2*r.height
}

func (c circle) area() float64 {
	return math.Pi * c.radius * c.radius
}

func (c circle) perim() float64 {
	return 2 * math.Pi * c.radius
}

// 如果变量具有接口类型，那么可以调用该接口的方法
func measure(g geometry) {
	fmt.Println(g)
	fmt.Println(g.area())
	fmt.Println(g.perim())
}

func main() {
	// rect和circle结构体类型都实现了geometry接口
	// 所以可以使用这些结构的实例作为参数
	r := rect{width: 3, height: 4}
	c := circle{radius: 5}
	measure(r)
	measure(c)
}
