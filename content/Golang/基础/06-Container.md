---
title: 06-Container
date: 2019-11-25T11:15:47.522182+08:00
draft: false
---

- [0.1. heap](#01-heap)
- [0.2. list](#02-list)
  - [0.2.1. 开箱即用](#021-开箱即用)
  - [0.2.2. 延迟初始化](#022-延迟初始化)
- [0.3. ring与list的区别](#03-ring与list的区别)

container包中的容器都不是线程安全的。

## 0.1. heap

heap 是一个堆的实现。一个堆正常保证了获取/弹出最大（最小）元素的时间为o(logn)、插入元素的时间为 o(logn)。

包中有示例。

堆实现接口如下：

```golang
// src/container/heap.go
type Interface interface {
 sort.Interface
 Push(x interface{}) // add x as element Len()
 Pop() interface{} // remove and return element Len() - 1.
}
```

heap 是基于 `sort.Interface` 实现的:

```golang
// src/sort/
type Interface interface {
 Len() int
 Less(i, j int) bool
 Swap(i, j int)
}
```

因此，如果要使用官方提供的 heap，需要我们实现如下几个接口：

```go
Len() int {} // 获取元素个数
Less(i, j int) bool  {} // 比较方法
Swap(i, j int) // 元素交换方法
Push(x interface{}){} // 在末尾追加元素
Pop() interface{} // 返回末尾元素
```

然后在使用时，我们可以使用如下几种方法：

```go
// 初始化一个堆
func Init(h Interface){}
// push一个元素倒堆中
func Push(h Interface, x interface{}){}
// pop 堆顶元素
func Pop(h Interface) interface{} {}
// 删除堆中某个元素，时间复杂度 log n
func Remove(h Interface, i int) interface{} {}
// 调整i位置的元素位置（位置I的数据变更后）
func Fix(h Interface, i int){}
```

## 0.2. list

Go语言的链表实现在标准库的`container/list`代码包中。代码包中有两个公开的程序实体：

- List：双向链表
- Element：链表中元素的结构

List的方法：

1. MoveBefore & MoveAfter：把**给定元素**移动到另一个元素的前面或后面
2. MoveToFront & MoveToBack：把**给定元素**移动到链表的最前端和最后端

> 这些给定元素都是`*Element`类型，`*Element`的值就是元素的指针。

```go
func (l *List) MoveBefore(e, mark *Element)
func (l *List) MoveAfter(e, mark *Element)

func (l *List) MoveToFront(e *Element)
func (l *List) MoveToBack(e *Element)
```

在List包含的方法中，用于插入新元素的那些方法都只接受Interface{}类型的值。这些方法在内部会使用Element值，包装接收到的新元素。**这样子为了避免链表的内部关联遭到外界破坏**。

```go
func (l *List) Front() *Element     //返回链表最前端元素的指针
func (l *List) Back() *Element      //返回链表最后段元素的指针

func (l *List) InsertBefore(v interface{}, mark *Element) *Element      //在链表指定元素前插入元素并返回插入元素的指针
func (l *List) InsertAfter(v interface{}, mark *Element) *Element       //在链表指定元素后插入元素并返回插入元素的指针

func (l *List) PushFront(v interface{}) *Element        //将指定元素插入到链表头
func (l *List) PushBack(v interface{}) *Element         //将指定元素插入到链表尾
```

### 0.2.1. 开箱即用

List和Element都是结构体类型，**结构体类型的特点是它们的零值都拥有特定结构，但是没有任何定制化内容的值，值中的字段都被赋予各自类型的零值**。

> 只做声明却没有初始化的变量被赋予的缺省值就是零值，每个类型的零值都会依据该类型的特性而被设定。

`var l list.List`声明的变量l的值是一个长度为0的链表，链表持有的根元素也是一个空壳，其中包含缺省的内容。这样的链表可以开箱即用的原因在于“延迟初始化”机制。

### 0.2.2. 延迟初始化

**优点：**

把初始化操作延后，仅在实际需要的时候才进行，延迟初始化的优点在于“延后”，**它可以分散初始化操作带来的计算量和存储空间消耗**。

> 如果需要集中声明非常多的大容量切片，那么CPU和内存的使用量肯定会激增，并且只要设法让其中的切片和底层数组被回收，内存使用量才会有所下降。如果数组可以被延迟初始化，那么CPU和内存的压力就被分散到实际使用它们的时候，这些数组被实际使用的时间越分散，延迟初始化的优势就越明显。

**缺点：**

延迟初始化的缺点也在于“延后”，如果在调用链表的每个方法的时候，都需要先去判断链表是否已经被初始化，这也是计算量上的浪费，这些方法被非常频繁地调用的情况下，这种浪费的影响就开始明显，程序的性能会降低。

**解决方案：**

1. 在链表的视线中，一些方法无需对是否初始化做判断，如`Front()`和`Back()`方法，一旦发现链表的长度为0，直接返回nil
2. 在插入、删除或移动元素的方法中，只要判断传入的元素中指向所属链表的指针，是否与当前链表的指针相等就可以

> `PushFront()`、`PushBack()`、`PushBackList()`、`PushFrontList()`方法总是会先判断链表的状态，并在必要时进行延迟初始化。在向新链表中添加新元素时，肯定会调用这四个方法之一。

## 0.3. ring与list的区别

`container/ring`包中的Ring类型实现的是一个循环链表（环）。

> List在内部就是一个循环链表，它的根元素永远不会持有任何实际的元素值，而该元素的存在是为了连接这个循环链表的收尾两端。List的零值是已给只包含根元素，不包含任何实际元素值的空链表。

| 差异                 | Ring                                                       | List                                      |
| -------------------- | ---------------------------------------------------------- | ----------------------------------------- |
| 表示方式、结构复杂度 | Ring类型的数据结构仅由它自身即可代表                       | List类型需要它及Element类型联合表示       |
| 表述维度             | Ring类型的值，严格来说只代表起所属的循环链表中的一个元素   | 而List类型的值则代表一个完整的链表        |
| New函数功能          | 创建并初始化Ring，可以指定包含的元素数量，创建后长度不可变 | 创建并初始化List不能指定包含的元素数量    |
| 初始化               | var r ring.Ring声明的r是一个长度为1的循环链表              | var l list.List声明的l是一个长度为0的链表 |
| 时间复杂度           | Ring的len方法时间复杂度o(N)                                | List的len方法时间复杂度o(1)               |

List中的根元素不会持有实际元素值，计算长度时不会包含它。
