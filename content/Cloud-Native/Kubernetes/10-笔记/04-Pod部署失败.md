---
title: 04 Pod部署失败
date: 2020-04-14T10:09:14.202627+08:00
draft: false
---

- [0.1. 错误的容器镜像/非法的仓库权限](#01-错误的容器镜像非法的仓库权限)
- [0.2. 应用启动之后又挂掉](#02-应用启动之后又挂掉)
- [0.3. 缺失 ConfigMap 或者 Secret](#03-缺失-configmap-或者-secret)
  - [0.3.1. 缺失 ConfigMap](#031-缺失-configmap)
  - [0.3.2. 缺失 Secrets](#032-缺失-secrets)
- [0.4. 活跃度/就绪状态探测失败](#04-活跃度就绪状态探测失败)
- [0.5. 超出CPU/内存的限制](#05-超出cpu内存的限制)
- [0.6. 资源配额](#06-资源配额)
- [0.7. 集群资源不足](#07-集群资源不足)
- [0.8. 持久化卷挂载失败](#08-持久化卷挂载失败)
- [0.9. 校验错误](#09-校验错误)
- [0.10. 容器镜像没有更新](#010-容器镜像没有更新)
- [0.11. 总结](#011-总结)

## 0.1. 错误的容器镜像/非法的仓库权限

其中两个最普遍的问题是：

- (a)指定了错误的容器镜像，
- (b)使用私有镜像却不提供仓库认证信息。

这在首次使用 Kubernetes 或者绑定 CI/CD 环境时尤其棘手。

让我们看个例子。

1. 首先我们创建一个名为 fail 的 deployment，它指向一个不存在的 Docker 镜像：

```bash
kubectl run fail --image=rosskukulinski/dne:v1.0.0
```

2. 然后我们查看 Pods，可以看到有一个状态为 ErrImagePull 或者 ImagePullBackOff 的 Pod：

```bash
$ kubectl get pods
NAME                    READY     STATUS             RESTARTS   AGE
fail-1036623984-hxoas   0/1       ImagePullBackOff   0          2m
```

3. 想查看更多信息，可以 describe 这个失败的 Pod：

```bash
kubectl describe pod fail-1036623984-hxoas
```

4. 查看 describe 命令的输出中
Events
这部分：

```bash
Events:
FirstSeen    LastSeen    Count   From                        SubObjectPath       Type        Reason      Message
---------    --------    -----   ----                        -------------       --------    ------      -------
5m        5m      1   {default-scheduler }                            Normal      Scheduled   Successfully assigned fail-1036623984-hxoas to gke-nrhk-1-default-pool-a101b974-wfp7
5m        2m      5   {kubelet gke-nrhk-1-default-pool-a101b974-wfp7} spec.containers{fail}   Normal      Pulling     pulling image "rosskukulinski/dne:v1.0.0"
5m        2m      5   {kubelet gke-nrhk-1-default-pool-a101b974-wfp7} spec.containers{fail}   Warning     Failed      Failed to pull image "rosskukulinski/dne:v1.0.0": Error: image rosskukulinski/dne not found
5m        2m      5   {kubelet gke-nrhk-1-default-pool-a101b974-wfp7}             Warning     FailedSync  Error syncing pod, skipping: failed to "StartContainer" for "fail" with ErrImagePull: "Error: image rosskukulinski/dne not found"

5m    11s 19  {kubelet gke-nrhk-1-default-pool-a101b974-wfp7} spec.containers{fail}   Normal  BackOff     Back-off pulling image "rosskukulinski/dne:v1.0.0"
5m    11s 19  {kubelet gke-nrhk-1-default-pool-a101b974-wfp7}             Warning FailedSync  Error syncing pod, skipping: failed to "StartContainer" for "fail" with ImagePullBackOff: "Back-off pulling image \"rosskukulinski/dne:v1.0.0\""
```

显示错误的那句话：Failed to pull image "rosskukulinski/dne:v1.0.0": Error: image rosskukulinski/dne not found 告诉我们 Kubernetes无法找到镜像 rosskukulinski/dne:v1.0.0。

**因此问题变成：为什么 Kubernetes 拉不下来镜像？**

除了
网络连接问题
外，还有三个主要元凶：

- 镜像 tag 不正确
- 镜像不存在（或者是在另一个仓库）
- Kubernetes 没有权限去拉那个镜像

>如果你没有注意到你的镜像 tag 的拼写错误，那么最好就用你本地机器测试一下。
>通常我会在本地开发机上，用 docker pull 命令，带上 完全相同的镜像 tag，来跑一下。

1. 比如上面的情况，我会运行命令 docker pull rosskukulinski/dne:v1.0.0。
如果这成功了，那么很可能 Kubernetes
没有权限
去拉取这个镜像。参考镜像拉取 Secrets 来解决这个问题。

2. 如果失败了，那么我会继续用不显式带 tag 的镜像测试 - docker pull rosskukulinski/dne - 这会尝试拉取 tag 为 latest 的镜像。如果这样成功，表明原来
指定的 tag 不存在
。这可能是人为原因，拼写错误，或者 CI/CD 的配置错误。

3. 如果 docker pull rosskukulinski/dne（不指定 tag）也失败了，那么我们碰到了一个更大的问题：我们所有的镜像仓库中都没有这个镜像。
    - 默认情况下，Kubernetes 使用 Dockerhub 镜像仓库，如果你在使用 Quay.io，AWS ECR，或者 Google Container Registry，你要在镜像地址中指定这个仓库的 URL，比如使用 Quay，镜像地址就变成 quay.io/rosskukulinski/dne:v1.0.0。

    - 如果你在使用 Dockerhub，那你应该再次确认你发布镜像到 Dockerhub 的系统，确保名字和 tag 匹配你的 deployment 正在使用的镜像。

注意：观察 Pod 状态的时候，镜像缺失和仓库权限不正确是没法区分的。其它情况下，Kubernetes 将报告一个 ErrImagePull 状态。

## 0.2. 应用启动之后又挂掉

>无论你是在 Kubernetes 上启动新应用，还是迁移应用到已存在的平台，应用在启动之后就挂掉都是一个比较常见的现象。

我们创建一个 deployment，它的应用会在1秒后挂掉：

```bash
kubectl run crasher --image=rosskukulinski/crashing-app
```

我们看一下 Pods 的状态：

```bash
$ kubectl get pods
NAME                       READY     STATUS             RESTARTS   AGE
crasher-2443551393-vuehs   0/1       CrashLoopBackOff   2          54s
```

CrashLoopBackOff 告诉我们，Kubernetes 正在尽力启动这个 Pod，但是一个或多个容器已经挂了，或者正被删除。

让我们 describe 这个 Pod 去获取更多信息：

```bash
$ kubectl describe pod crasher-2443551393-vuehs
Name:        crasher-2443551393-vuehs
Namespace:    fail
Node:        gke-nrhk-1-default-pool-a101b974-wfp7/10.142.0.2
Start Time:    Fri, 10 Feb 2017 14:20:29 -0500
Labels:        pod-template-hash=2443551393
    run=crasher
Status:        Running
IP:        10.0.0.74
Controllers:    ReplicaSet/crasher-2443551393
Containers:
crasher:
Container ID:    docker://51c940ab32016e6d6b5ed28075357661fef3282cb3569117b0f815a199d01c60
Image:        rosskukulinski/crashing-app
Image ID:        docker://sha256:cf7452191b34d7797a07403d47a1ccf5254741d4bb356577b8a5de40864653a5
Port:
State:        Terminated
  Reason:        Error
  Exit Code:    1
  Started:        Fri, 10 Feb 2017 14:22:24 -0500
  Finished:        Fri, 10 Feb 2017 14:22:26 -0500
Last State:        Terminated
  Reason:        Error
  Exit Code:    1
  Started:        Fri, 10 Feb 2017 14:21:39 -0500
  Finished:        Fri, 10 Feb 2017 14:21:40 -0500
Ready:        False
Restart Count:    4
```

好可怕，Kubernetes 告诉我们这个 Pod 正被 Terminated，因为容器里的应用挂了。我们还可以看到应用的 Exit Code 是 1。后面我们可能还会看到一个 OOMKilled 错误。

**我们的应用正在挂掉？为什么？**

首先我们查看应用日志。假定你发送应用日志到 stdout（事实上你也应该这么做），你可以使用 kubectl logs 看到应用日志:

```bash
kubectl logs crasher-2443551393-vuehs
```

不幸的是，这个 Pod 没有任何日志。这可能是因为我们正在查看一个新起的应用实例，因此我们应该查看前一个容器：

```bash
kubectl logs crasher-2443551393-vuehs --previous
```

什么！我们的应用仍然不给我们任何东西。这个时候我们应该给应用加点启动日志了，以帮助我们定位这个问题。我们也可以本地运行一下这个容器，以确定是否缺失环境变量或者挂载卷。

## 0.3. 缺失 ConfigMap 或者 Secret

Kubernetes 最佳实践建议**通过 ConfigMaps 或者 Secrets 传递应用的
运行时配置
**。这些数据可以包含数据库认证信息，API endpoints，或者其它配置信息。

一个常见的错误是，创建的 deployment 中引用的 ConfigMaps 或者 Secrets 的属性不存在，有时候甚至引用的 ConfigMaps 或者 Secrets 本身就不存在。

### 0.3.1. 缺失 ConfigMap

第一个例子，我们将尝试创建一个 Pod，它加载 ConfigMap 数据作为环境变量：

```
# configmap-pod.yaml
apiVersion: v1
kind: Pod
metadata:
name: configmap-pod
spec:
containers:
- name: test-container
  image: gcr.io/google_containers/busybox
  command: [ "/bin/sh", "-c", "env" ]
  env:
    - name: SPECIAL_LEVEL_KEY
      valueFrom:
        configMapKeyRef:
          name: special-config
          key: special.how
```

让我们创建一个 Pod：

```bash
kubectl create -f configmap-pod.yaml
```

在等待几分钟之后，我们可以查看我们的 Pod：

```bash
$ kubectl get pods
NAME            READY     STATUS              RESTARTS   AGE
configmap-pod   0/1       RunContainerError   0          3s
```

Pod 状态是 RunContainerError 。我们可以使用 kubectl describe 了解更多：

```bash
$ kubectl describe pod configmap-pod
[...]
Events:
FirstSeen    LastSeen    Count   From                        SubObjectPath           Type        Reason      Message
---------    --------    -----   ----                        -------------           --------    ------      -------
20s        20s     1   {default-scheduler }                                Normal      Scheduled   Successfully assigned configmap-pod to gke-ctm-1-sysdig2-35e99c16-tgfm
19s        2s      3   {kubelet gke-ctm-1-sysdig2-35e99c16-tgfm}   spec.containers{test-container} Normal      Pulling     pulling image "gcr.io/google_containers/busybox"
18s        2s      3   {kubelet gke-ctm-1-sysdig2-35e99c16-tgfm}   spec.containers{test-container} Normal      Pulled      Successfully pulled image "gcr.io/google_containers/busybox"
18s        2s      3   {kubelet gke-ctm-1-sysdig2-35e99c16-tgfm}                   Warning     FailedSync  Error syncing pod, skipping: failed to "StartContainer" for "test-container" with RunContainerError: "GenerateRunContainerOptions: configmaps \"special-config\" not found"
```

Events 章节的最后一条告诉我们什么地方错了。Pod 尝试访问名为 special-config 的 ConfigMap，但是在该 namespace 下找不到。一旦我们创建这个 ConfigMap，Pod 应该重启并能成功拉取运行时数据。

在 Pod 规格说明中访问 Secrets作为环境变量会产生相似的错误，就像我们在这里看到的 ConfigMap错误一样。

**通过 Volume 来访问 Secrets 或者 ConfigMap会发生什么呢？**

### 0.3.2. 缺失 Secrets

下面是一个pod规格说明，它引用了名为 myothersecret 的 Secrets，并尝试把它挂为卷：

```bash
# missing-secret.yaml
apiVersion: v1
kind: Pod
metadata:
name: secret-pod
spec:
containers:
- name: test-container
  image: gcr.io/google_containers/busybox
  command: [ "/bin/sh", "-c", "env" ]
  volumeMounts:
    - mountPath: /etc/secret/
      name: myothersecret
restartPolicy: Never
volumes:
- name: myothersecret
  secret:
    secretName: myothersecret
```

让我们用 kubectl create -f missing-secret.yaml 来创建一个 Pod。

几分钟后，我们 get Pods，可以看到 Pod 仍处于 ContainerCreating 状态：

```bash
$ kubectl get pods
NAME            READY     STATUS              RESTARTS   AGE
secret-pod   0/1       ContainerCreating   0          4h
```

这就奇怪了。我们 describe 一下，看看到底发生了什么：

```bash
$ kubectl describe pod secret-pod
Name:        secret-pod
Namespace:    fail
Node:        gke-ctm-1-sysdig2-35e99c16-tgfm/10.128.0.2
Start Time:    Sat, 11 Feb 2017 14:07:13 -0500
Labels:
Status:        Pending
IP:
Controllers:

[...]

Events:
FirstSeen    LastSeen    Count   From                        SubObjectPath   Type        Reason      Message
---------    --------    -----   ----                        -------------   --------    ------      -------
18s        18s     1   {default-scheduler }                        Normal      Scheduled   Successfully assigned secret-pod to gke-ctm-1-sysdig2-35e99c16-tgfm
18s        2s      6   {kubelet gke-ctm-1-sysdig2-35e99c16-tgfm}           Warning     FailedMount MountVolume.SetUp failed for volume "kubernetes.io/secret/337281e7-f065-11e6-bd01-42010af0012c-myothersecret" (spec.Name: "myothersecret") pod "337281e7-f065-11e6-bd01-42010af0012c" (UID: "337281e7-f065-11e6-bd01-42010af0012c") with: secrets "myothersecret" not found
```

Events 章节再次解释了问题的原因。它告诉我们 Kubelet 无法从名为 myothersecret 的 Secret 挂卷。为了解决这个问题，我们可以创建 myothersecret ，它包含必要的安全认证信息。一旦 myothersecret 创建完成，容器也将正确启动。

## 0.4. 活跃度/就绪状态探测失败

在 Kubernetes 中处理容器问题时，开发者需要学习的重要一课是，
你的容器应用是 running 状态，不代表它在工作
。

Kubernetes 提供了两个基本特性，称作**活跃度探测**和**就绪状态探测**。本质上来说，活跃度/就绪状态探测将定期地执行一个操作（例如发送一个 HTTP 请求，打开一个 tcp 连接，或者在你的容器内运行一个命令），以确认你的应用和你预想的一样在工作。

- 如果活跃度探测失败，Kubernetes 将杀掉你的容器并重新创建一个。
- 如果就绪状态探测失败，这个 Pod 将不会作为一个服务的后端 endpoint，也就是说不会流量导到这个 Pod，直到它变成 Ready。

如果你试图部署变更你的活跃度/就绪状态探测失败的应用，滚动部署将一直悬挂，因为它将等待你的所有 Pod 都变成 Ready。

**这个实际是怎样的情况？**

以下是一个 Pod 规格说明，它定义了活跃度/就绪状态探测方法，都是基于8080端口对 /healthy 路由进行健康检查：

```bash
apiVersion: v1
kind: Pod
metadata:
name: liveness-pod
spec:
containers:
- name: test-container
  image: rosskukulinski/leaking-app
  livenessProbe:
    httpGet:
      path: /healthz
      port: 8080
    initialDelaySeconds: 3
    periodSeconds: 3
  readinessProbe:
    httpGet:
      path: /healthz
      port: 8080
    initialDelaySeconds: 3
    periodSeconds: 3
```

让我们创建这个 Pod：kubectl create -f liveness.yaml，过几分钟后查看发生了什么：

```bash
$ kubectl get pods
NAME           READY     STATUS    RESTARTS   AGE
liveness-pod   0/1       Running   4          2m
```

2分钟以后，我们发现 Pod 仍然没处于 Ready 状态，并且它已被重启了4次。让我们 describe 一下查看更多信息：

```bash
$ kubectl describe pod liveness-pod
Name:        liveness-pod
Namespace:    fail
Node:        gke-ctm-1-sysdig2-35e99c16-tgfm/10.128.0.2
Start Time:    Sat, 11 Feb 2017 14:32:36 -0500
Labels:
Status:        Running
IP:        10.108.88.40
Controllers:
Containers:
test-container:
Container ID:    docker://8fa6f99e6fda6e56221683249bae322ed864d686965dc44acffda6f7cf186c7b
Image:        rosskukulinski/leaking-app
Image ID:        docker://sha256:7bba8c34dad4ea155420f856cd8de37ba9026048bd81f3a25d222fd1d53da8b7
Port:
State:        Running
  Started:        Sat, 11 Feb 2017 14:40:34 -0500
Last State:        Terminated
  Reason:        Error
  Exit Code:    137
  Started:        Sat, 11 Feb 2017 14:37:10 -0500
  Finished:        Sat, 11 Feb 2017 14:37:45 -0500
[...]
Events:
FirstSeen    LastSeen    Count   From                        SubObjectPath           Type        Reason      Message
---------    --------    -----   ----                        -------------           --------    ------      -------
8m        8m      1   {default-scheduler }                                Normal      Scheduled   Successfully assigned liveness-pod to gke-ctm-1-sysdig2-35e99c16-tgfm
8m        8m      1   {kubelet gke-ctm-1-sysdig2-35e99c16-tgfm}   spec.containers{test-container} Normal      Created     Created container with docker id 0fb5f1a56ea0; Security:[seccomp=unconfined]
8m        8m      1   {kubelet gke-ctm-1-sysdig2-35e99c16-tgfm}   spec.containers{test-container} Normal      Started     Started container with docker id 0fb5f1a56ea0
7m        7m      1   {kubelet gke-ctm-1-sysdig2-35e99c16-tgfm}   spec.containers{test-container} Normal      Created     Created container with docker id 3f2392e9ead9; Security:[seccomp=unconfined]
7m        7m      1   {kubelet gke-ctm-1-sysdig2-35e99c16-tgfm}   spec.containers{test-container} Normal      Killing     Killing container with docker id 0fb5f1a56ea0: pod "liveness-pod_fail(d75469d8-f090-11e6-bd01-42010af0012c)" container "test-container" is unhealthy, it will be killed and re-created.
8m    16s 10  {kubelet gke-ctm-1-sysdig2-35e99c16-tgfm}   spec.containers{test-container} Warning Unhealthy   Liveness probe failed: Get http://10.108.88.40:8080/healthz: dial tcp 10.108.88.40:8080: getsockopt: connection refused
8m    1s  85  {kubelet gke-ctm-1-sysdig2-35e99c16-tgfm}   spec.containers{test-container} Warning Unhealthy   Readiness probe failed: Get http://10.108.88.40:8080/healthz: dial tcp 10.108.88.40:8080: getsockopt: connection refused

```

Events 章节再次救了我们。我们可以看到活跃度探测和就绪状态探测都失败了。关键的一句话是 container "test-container" is unhealthy, it will be killed and re-created。这告诉我们 Kubernetes 正在杀这个容器，因为容器的活跃度探测失败了。

这里有三种可能性：

- 你的探测不正确，健康检查的 URL 是否改变了？
- 你的探测太敏感了， 你的应用是否要过一会才能启动或者响应？
- 你的应用永远不会对探测做出正确响应，你的数据库是否配置错了

查看 Pod 日志是一个开始调测的好地方。一旦你解决了这个问题，新的 deployment 应该就能成功了。

## 0.5. 超出CPU/内存的限制

Kubernetes 赋予集群管理员限制 Pod 和容器的 CPU 或内存数量的能力。作为应用开发者，你可能不清楚这个限制，导致 deployment 失败的时候一脸困惑。

我们试图部署一个未知 CPU/memory 请求限额的 deployment：

```yaml
# gateway.yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
name: gateway
spec:
template:
metadata:
  labels:
    app: gateway
spec:
  containers:
    - name: test-container
      image: nginx
      resources:
        requests:
          memory: 5Gi
```

你会看到我们设了 5Gi 的资源请求。让我们创建这个 deployment：kubectl create -f gateway.yaml。

现在我们可以看到我们的 Pod：

```bash
$ kubectl get pods
No resources found.
```

为啥，让我们用 describe 来观察一下我们的 deployment：

```bash
$ kubectl describe deployment/gateway
Name:            gateway
Namespace:        fail
CreationTimestamp:    Sat, 11 Feb 2017 15:03:34 -0500
Labels:            app=gateway
Selector:        app=gateway
Replicas:        0 updated | 1 total | 0 available | 1 unavailable
StrategyType:        RollingUpdate
MinReadySeconds:    0
RollingUpdateStrategy:    0 max unavailable, 1 max surge
OldReplicaSets:
NewReplicaSet:        gateway-764140025 (0/1 replicas created)
Events:
FirstSeen    LastSeen    Count   From                SubObjectPath   Type        Reason          Message
---------    --------    -----   ----                -------------   --------    ------          -------
4m        4m      1   {deployment-controller }            Normal      ScalingReplicaSet   Scaled up replica set gateway-764140025 to 1
```

基于最后一行，我们的 deployment 创建了一个 ReplicaSet（gateway-764140025） 并把它扩展到 1。这个是用来管理 Pod 生命周期的实体。我们可以 describe 这个 ReplicaSet：

```bash
$ kubectl describe rs/gateway-764140025
Name:        gateway-764140025
Namespace:    fail
Image(s):    nginx
Selector:    app=gateway,pod-template-hash=764140025
Labels:        app=gateway
    pod-template-hash=764140025
Replicas:    0 current / 1 desired
Pods Status:    0 Running / 0 Waiting / 0 Succeeded / 0 Failed
No volumes.
Events:
FirstSeen    LastSeen    Count   From                SubObjectPath   Type        Reason      Message
---------    --------    -----   ----                -------------   --------    ------      -------
6m        28s     15  {replicaset-controller }            Warning     FailedCreate    Error creating: pods "gateway-764140025-" is forbidden: [maximum memory usage per Pod is 100Mi, but request is 5368709120., maximum memory usage per Container is 100Mi, but request is 5Gi.]
```

哈知道了。集群管理员设置了每个 Pod 的最大内存使用量为 100Mi（好一个小气鬼！）。你可以运行

```bash
kubectl describe limitrange
```

来查看当前租户的限制。

你现在有3个选择：

- 要求你的集群管理员提升限额
- 减少 deployment 的请求或者限额设置
- 直接编辑限额

## 0.6. 资源配额

和资源限额类似，Kubernetes 也允许管理员给每个 namespace 设置资源配额。这些配额可以在 Pods，Deployments，PersistentVolumes，CPU，内存等资源上设置**软性**或者**硬性**限制。

让我们看看超出资源配额后会发生什么。以下是我们的 deployment 例子：

```yaml
# test-quota.yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
name: gateway-quota
spec:
template:
spec:
  containers:
    - name: test-container
      image: nginx
```

我们可用 kubectl create -f test-quota.yaml 创建，然后观察我们的 Pods：

```bash
$ kubectl get pods
NAME                            READY     STATUS    RESTARTS   AGE
gateway-quota-551394438-pix5d   1/1       Running   0          16s
```

看起来很好，现在让我们扩展到 3 个副本：

```bash
kubectl scale deploy/gateway-quota --replicas=3
```

然后再次观察 Pods：

```bash
$ kubectl get pods
NAME                            READY     STATUS    RESTARTS   AGE
gateway-quota-551394438-pix5d   1/1       Running   0          9m
```

啊，我们的pod去哪了？让我们观察一下 deployment：

```bash
$ kubectl describe deploy/gateway-quota
Name:            gateway-quota
Namespace:        fail
CreationTimestamp:    Sat, 11 Feb 2017 16:33:16 -0500
Labels:            app=gateway
Selector:        app=gateway
Replicas:        1 updated | 3 total | 1 available | 2 unavailable
StrategyType:        RollingUpdate
MinReadySeconds:    0
RollingUpdateStrategy:    1 max unavailable, 1 max surge
OldReplicaSets:
NewReplicaSet:        gateway-quota-551394438 (1/3 replicas created)
Events:
FirstSeen    LastSeen    Count   From                SubObjectPath   Type        Reason          Message
---------    --------    -----   ----                -------------   --------    ------          -------
9m        9m      1   {deployment-controller }            Normal      ScalingReplicaSet   Scaled up replica set gateway-quota-551394438 to 1
5m        5m      1   {deployment-controller }            Normal      ScalingReplicaSet   Scaled up replica set gateway-quota-551394438 to 3
```

在最后一行，我们可以看到 ReplicaSet 被告知扩展到 3 。我们用 describe 来观察一下这个 ReplicaSet 以了解更多信息：

```bash
kubectl describe replicaset gateway-quota-551394438
Name:        gateway-quota-551394438
Namespace:    fail
Image(s):    nginx
Selector:    app=gateway,pod-template-hash=551394438
Labels:        app=gateway
    pod-template-hash=551394438
Replicas:    1 current / 3 desired
Pods Status:    1 Running / 0 Waiting / 0 Succeeded / 0 Failed
No volumes.
Events:
FirstSeen    LastSeen    Count   From                SubObjectPath   Type        Reason          Message
---------    --------    -----   ----                -------------   --------    ------          -------
11m        11m     1   {replicaset-controller }            Normal      SuccessfulCreate    Created pod: gateway-quota-551394438-pix5d
11m        30s     33  {replicaset-controller }            Warning     FailedCreate        Error creating: pods "gateway-quota-551394438-" is forbidden: exceeded quota: compute-resources, requested: pods=1, used: pods=1, limited: pods=1
```

哦！我们的 ReplicaSet 无法创建更多的 pods 了，因为配额限制了：

```bash
exceeded quota: compute-resources, requested: pods=1, used: pods=1, limited: pods=1。
```

和资源限额类似，我们也有 3 个选项：

- 要求集群管理员提升该 namespace 的配额
- 删除或者收缩该 namespace 下其它的 deployment
- 直接编辑配额

## 0.7. 集群资源不足

除非你的集群开通了集群自动伸缩功能，否则总有一天你的集群中 CPU 和内存资源会耗尽。

这不是说 CPU 和内存被完全使用了,而是指它们被 Kubernetes 调度器完全使用了。如同我们在第 5 点看到的，集群管理员可以限制开发者能够申请分配给 pod 或者容器的 CPU 或者内存的数量。聪明的管理员也会设置一个默认的 CPU/内存 申请数量，在开发者未提供申请额度时使用。

如果你所有的工作都在 default 这个 namespace 下工作，你很可能有个默认值 100m 的容器 CPU申请额度，对此你甚至可能都不清楚。运行 kubectl describe ns default 检查一下是否如此。

我们假定你的 Kubernetes 集群只有一个包含 CPU 的节点。你的 Kubernetes 集群有 1000m 的可调度 CPU。

当前忽略其它的系统 pods（kubectl -n kube-system get pods），你的单节点集群能部署 10 个 pod(每个 pod 都只有一个包含 100m 的容器)。

10 Pods *(1 Container* 100m) = 1000m
 Cluster CPUs

**当你扩大到 11 个的时候，会发生什么？**

下面是一个申请 1CPU（1000m）的 deployment 例子：

```yaml
# cpu-scale.yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
name: cpu-scale
spec:
template:
metadata:
  labels:
    app: cpu-scale
spec:
  containers:
    - name: test-container
      image: nginx
      resources:
        requests:
          cpu: 1
```

我把这个应用部署到有 2 个可用 CPU 的集群。除了我的 cpu-scale 应用，Kubernetes 内部服务也在消耗 CPU 和内存。

我们可以用 kubectl create -f cpu-scale.yaml 部署这个应用，并观察 pods：

```bash
$ kubectl get pods
NAME                        READY     STATUS    RESTARTS   AGE
cpu-scale-908056305-xstti   1/1       Running   0          5m
```

第一个 pod 被调度并运行了。我们看看扩展一个会发生什么：

```bash
$ kubectl scale deploy/cpu-scale --replicas=2 deployment "cpu-scale" scaled
$ kubectl get pods
NAME                        READY     STATUS    RESTARTS   AGE
cpu-scale-908056305-phb4j   0/1       Pending   0          4m
cpu-scale-908056305-xstti   1/1       Running   0          5m
```

我们的第二个pod一直处于 Pending，被阻塞了。我们可以 describe 这第二个 pod 查看更多的信息:

```bash
$ kubectl describe pod cpu-scale-908056305-phb4j
Name:        cpu-scale-908056305-phb4j
Namespace:    fail
Node:        gke-ctm-1-sysdig2-35e99c16-qwds/10.128.0.4
Start Time:    Sun, 12 Feb 2017 08:57:51 -0500
Labels:        app=cpu-scale
    pod-template-hash=908056305
Status:        Pending
IP:
Controllers:    ReplicaSet/cpu-scale-908056305
[...]
Events:
FirstSeen    LastSeen    Count   From            SubObjectPath   Type        Reason          Message
---------    --------    -----   ----            -------------   --------    ------          -------
3m        3m      1   {default-scheduler }            Warning     FailedScheduling    pod (cpu-scale-908056305-phb4j) failed to fit in any node
fit failure on node (gke-ctm-1-sysdig2-35e99c16-wx0s): Insufficient cpu
fit failure on node (gke-ctm-1-sysdig2-35e99c16-tgfm): Insufficient cpu
fit failure on node (gke-ctm-1-sysdig2-35e99c16-qwds): Insufficient cpu
```

好吧，Events 模块告诉我们 Kubernetes 调度器（default-scheduler）无法调度这个 pod 因为它无法匹配任何节点。它甚至告诉我们每个节点哪个扩展点失败了（Insufficient cpu）。

那么我们如何解决这个问题？如果你太渴望你申请的 CPU/内存 的大小，你可以减少申请的大小并重新部署。当然，你也可以请求你的集群管理员扩展这个集群（因为很可能你不是唯一一个碰到这个问题的人）。

现在你可能会想：我们的 Kubernetes 节点是在我们的云提供商的自动伸缩群组里，为什么他们没有生效呢？

原因是，你的云提供商没有深入理解 Kubernetes 调度器是做啥的。利用 Kubernetes 的集群自动伸缩能力允许你的集群根据调度器的需求自动伸缩它自身。如果你在使用 GCE，集群伸缩能力是一个 beta 特性。

## 0.8. 持久化卷挂载失败

另一个常见错误是创建了一个引用不存在的持久化卷（PersistentVolumes）的 deployment。不论你是使用 PersistentVolumeClaims（你应该使用这个！），还是直接访问持久化磁盘，最终结果都是类似的。

下面是我们的测试 deployment，它想使用一个名为 my-data-disk 的 GCE 持久化卷：

```yaml
# volume-test.yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
name: volume-test
spec:
template:
metadata:
  labels:
    app: volume-test
spec:
  containers:
    - name: test-container
      image: nginx
      volumeMounts:
      - mountPath: /test
        name: test-volume
  volumes:
  - name: test-volume
    # This GCE PD must already exist (oops!)
    gcePersistentDisk:
      pdName: my-data-disk
      fsType: ext4
```

让我们创建这个 deployment：kubectl create -f volume-test.yaml，过几分钟后查看 pod：

```bash
kubectl get pods
NAME                           READY     STATUS              RESTARTS   AGE
volume-test-3922807804-33nux   0/1       ContainerCreating   0          3m
```

3 分钟的等待容器创建时间是很长了。让我们用 describe 来查看这个 pod，看看到底发生了什么：

```bash
$ kubectl describe pod volume-test-3922807804-33nux
Name:        volume-test-3922807804-33nux
Namespace:    fail
Node:        gke-ctm-1-sysdig2-35e99c16-qwds/10.128.0.4
Start Time:    Sun, 12 Feb 2017 09:24:50 -0500
Labels:        app=volume-test
    pod-template-hash=3922807804
Status:        Pending
IP:
Controllers:    ReplicaSet/volume-test-3922807804
[...]
Volumes:
test-volume:
Type:    GCEPersistentDisk (a Persistent Disk resource in Google Compute Engine)
PDName:    my-data-disk
FSType:    ext4
Partition:    0
ReadOnly:    false
[...]
Events:
FirstSeen    LastSeen    Count   From                        SubObjectPath   Type        Reason      Message
---------    --------    -----   ----                        -------------   --------    ------      -------
4m        4m      1   {default-scheduler }                        Normal      Scheduled   Successfully assigned volume-test-3922807804-33nux to gke-ctm-1-sysdig2-35e99c16-qwds
1m        1m      1   {kubelet gke-ctm-1-sysdig2-35e99c16-qwds}           Warning     FailedMount Unable to mount volumes for pod "volume-test-3922807804-33nux_fail(e2180d94-f12e-11e6-bd01-42010af0012c)": timeout expired waiting for volumes to attach/mount for pod "volume-test-3922807804-33nux"/"fail". list of unattached/unmounted volumes=[test-volume]
1m        1m      1   {kubelet gke-ctm-1-sysdig2-35e99c16-qwds}           Warning     FailedSync  Error syncing pod, skipping: timeout expired waiting for volumes to attach/mount for pod "volume-test-3922807804-33nux"/"fail". list of unattached/unmounted volumes=[test-volume]
3m        50s     3   {controller-manager }                       Warning     FailedMount Failed to attach volume "test-volume" on node "gke-ctm-1-sysdig2-35e99c16-qwds" with: GCE persistent disk not found: diskName="my-data-disk" zone="us-central1-a"
```

很神奇！ Events 模块留有我们一直在寻找的线索。我们的 pod 被正确调度到了一个节点（Successfully assigned volume-test-3922807804-33nux to gke-ctm-1-sysdig2-35e99c16-qwds），但是那个节点上的 kubelet 无法挂载期望的卷 test-volume。那个卷本应该在持久化磁盘被关联到这个节点的时候就被创建了，但是，正如我们看到的，controller-manager 失败了：Failed to attach volume "test-volume" on node "gke-ctm-1-sysdig2-35e99c16-qwds" with: GCE persistent disk not found: diskName="my-data-disk" zone="us-central1-a"。

最后一条信息相当清楚了：为了解决这个问题，我们需要在 GKE 的 us-central1-a 区中创建一个名为 my-data-disk 的持久化卷。一旦这个磁盘创建完成，controller-manager 将挂载这块磁盘，并启动容器创建过程。

## 0.9. 校验错误

看着整个 build-test-deploy 任务到了 deploy 步骤却失败了，原因竟是 Kubernetes 对象不合法。还有什么比这更让人沮丧的！

你可能之前也碰到过这种错误:

```bash
$ kubectl create -f test-application.deploy.yaml
error: error validating "test-application.deploy.yaml": error validating data: found invalid field resources for v1.PodSpec; if you choose to ignore these errors, turn validation off with --validate=false
```

在这个例子中，我尝试创建以下 deployment：

```yaml
# test-application.deploy.yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
name: test-app
spec:
template:
metadata:
  labels:
    app: test-app
spec:
  containers:
  - image: nginx
    name: nginx
  resources:
    limits:
      cpu: 100m
      memory: 200Mi
    requests:
      cpu: 100m
      memory: 100Mi
```

一眼望去，这个 YAML 文件是正确的，但错误消息会证明是有用的。错误说的是 found invalid field resources for v1.PodSpec，再仔细看一下 v1.PodSpec， 我们可以看到 resource 对象变成了 v1.PodSpec 的一个子对象。事实上它应该是 v1.Container 的子对象。在把 resource 对象缩进一层后，这个 deployment 对象就可以正常工作了。

除了查找缩进错误，另一个常见的错误是写错了对象名（比如 peristentVolumeClaim 写成了 persistentVolumeClaim）。这个错误曾经在我们时间很赶的时候绊住了我和另一位高级工程师。

**为了能在早期就发现这些错误，我推荐在 pre-commit 钩子或者构建的测试阶段添加一些校验步骤。**

例如，你可以用

```bash
python -c 'import yaml,sys;yaml.safe_load(sys.stdin)' < test-application.deployment.yaml
```

验证 YAML 格式

使用标识 --dry-run 来验证 Kubernetes API 对象，比如这样：

```bash
kubectl create -f test-application.deploy.yaml --dry-run --validate=true
```

重要提醒：校验 Kubernetes 对象的机制是在服务端的校验，这意味着 kubectl 必须有一个在工作的 Kubernetes 集群与之通信。不幸的是，当前 kubectl 还没有客户端的校验选项，但是已经有 issue（kubernetes/kubernetes #29410 和 kubernetes/kubernetes #11488）在跟踪这个缺失的特性了。

## 0.10. 容器镜像没有更新

我了解的在使用 Kubernetes 的大多数人都碰到过这个问题，它也确实是一个难题。

这个场景就像下面这样：

1. 使用一个镜像 tag（比如：rosskulinski/myapplication:v1） 创建一个 deployment
2. 注意到 myapplication 镜像中存在一个 bug
3. 构建了一个新的镜像，并推送到了相同的 tag（rosskukulinski/myapplication:v1）
4. 删除了所有 myapplication 的 pods，新的实例被 deployment 创建出了
5. 发现 bug 仍然存在
6. 重复 3-5 步直到你抓狂为止

这个问题关系到 Kubernetes 在启动 pod 内的容器时是如何决策是否做 docker pull 动作的。

在 v1.Container 说明中，有一个选项 ImagePullPolicy：

>Image pull policy. One of Always, Never, IfNotPresent. Defaults to Always if :latest tag is specified, or IfNotPresent otherwise.

因为我们把我们的镜像 tag 标记为 :v1，默认的镜像拉取策略是 IfNotPresent。Kubelet 在本地已经有一份 rosskukulinski/myapplication:v1 的拷贝了，因此它就不会在做 docker pull 动作了。当新的 pod 出现的时候，它仍然使用了老的有问题的镜像。

有三个方法来解决这个问题：

- 切成 :latest tag（千万不要这么做！）
- deployment 中指定 ImagePullPolicy: Always
- 使用唯一的 tag（比如基于你的代码版本控制器的 commit id）

在开发阶段或者要快速验证原型的时候，我会指定 ImagePullPolicy: Always 这样我可以使用相同的 tag 来构建和推送。

然而，在我的产品部署阶段，我使用基于 Git SHA-1 的唯一 tag。这样很容易查到产品部署的应用使用的源代码。

## 0.11. 总结

哇哦，我们有这么多地方要当心。到目前为止，你应该已经成为一个能定位，识别和修复失败的 Kubernetes 部署的专家了。

一般来说，大部分常见的部署失败都可以用下面的命令定位出来：

```bash
kubectl describe deployment/<deployname>
kubectl describe replicaset/<rsname>
kubectl get pods
kubectl describe pod/<podname>
kubectl logs <podname> --previous
```

在追求自动化，把我从繁琐的定位工作中解放出来的过程中，我写了一个 bash 脚本，它在 CI/CD 的部署过程中任何失败的时候，都可以跑。在 Jenkins/CircleCI 等的构建输出中，将显示有用的 Kubernetes 信息，帮助开发者快速找到任何明显的问题。
