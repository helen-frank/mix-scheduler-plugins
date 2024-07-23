# mix-scheduler-plugins

## Background

In all cloud providers, like AWS, Google, and others, there are many spot instances. They are quite cheap (10% of the on-demand instances' price), but after you buy them, they could be terminated with only two minutes' notice in advance (in most scenarios, we don't set PDB, and we should perform the graceful drain).

So, I want you to design a strategy to maximize the use of spot instances without causing service interruptions, instead of relying solely on on-demand instances, to cut costs, by using distributed scheduling in a single cluster (on-demand/spot mixed or other methods for one workload). This is important because all spot instances being terminated at the same time could cause interruptions for different kinds of workloads (single replica workload, multiple replica workload).

Also, I don't want to change the scheduler already used in the K8s cluster and want to ensure the minimal components necessary in the cluster.

Notes:

> 1. On demand nodes has label: node.kubernetes.io/capacity: on-demand.
> 2. Spot node has label: node.kubernetes.io/capacity: spot.
> 3. Workloads represented as Deployments and StatefulSets.
> 4. on-demand/spot instance represented as K8s nodes in the cluster.
> 5. Only focus on scheduling control; the graceful drain after receiving the terminal notification is handled by other components.

## 设计思路

- spot实例比较适合灵活性较高或具有容错性的应用程序
- 为了保证服务的高可用性就需要在on-demand节点(非spot节点)上保持一定量应用pod, 并且在spot节点上的pod尽量分散节点部署, 避免单点spot节点下线导致的短时压力飙升, 过于加大其他pod的压力, 降低服务的可用性
- 在未特殊配置的情况下, 尽量保证应用的大部分pod会分散部署在不同的spot节点上
- 支持自定义可用性保证 mix-scheduler-plugins/availability-guarantee

### 插入点

#### score

- 当应用在集群中无被节点匹配的pod时, 应尽量调度到on-demand节点
- 存在被节点匹配的pod时,但是on-demand节点上无pod时, 应尽量调度到on-demand节点
- mix-scheduler-plugins/availability-guarantee=0 or 1, 或者未设置的情况下应保证最低可用性, spot节点均衡性调度 `getUniformlyDistributedSocre(weight, currentNodeSelctPodNum int)`, on-demand节点不调整调度(此时经过上面的选择已经可确认已至少有1个pod在on-demand节点上)
- hasOnDemandPodNum >= mix-scheduler-plugins/availability-guarantee 时, spot节点均衡性调度 `getUniformlyDistributedSocre(weight, currentNodeSelctPodNum int)`, on-demand节点不调整调度. hasOnDemandPodNum < mix-scheduler-plugins/availability-guarantee 时 on-demand节点均衡性调度`getUniformlyDistributedSocre(weight, currentNodeSelctPodNum int)`, spot节点不调整调度

#### postBind

根据pod所在调度节点, 配置pod annotations, 方便查询

## 快速开始

### 部署

```bash
kubectl apply -f examples/mix-scheduler.yaml
```

### 构建

```bash
make dockerBuild
```

### 测试

#### 初始化环境

kind.config

```yaml
# this config file contains all config fields with comments
# NOTE: this is not a particularly useful config file
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4

networking:
  # WARNING: It is _strongly_ recommended that you keep this the default
  # (127.0.0.1) for security reasons. However it is possible to change this.
  apiServerAddress: "0.0.0.0"
  # By default the API server listens on a random open port.
  # You may choose a specific port but probably don't need to in most cases.
  # Using a random port makes it easier to spin up multiple clusters.
  #apiServerPort: 6443

# patch the generated kubeadm config with some extra settings
kubeadmConfigPatches:
- |
  apiVersion: kubelet.config.k8s.io/v1beta1
  kind: KubeletConfiguration
  evictionHard:
    nodefs.available: "0%"
# patch it further using a JSON 6902 patch
kubeadmConfigPatchesJSON6902:
- group: kubeadm.k8s.io
  version: v1beta2
  kind: ClusterConfiguration
  patch: |
    - op: add
      path: /apiServer/certSANs/-
      value: my-hostname

# 1 control plane node and 11 workers
nodes:
# the control plane node config
- role: control-plane
# the three workers
- role: worker
- role: worker
- role: worker
- role: worker
- role: worker
- role: worker
- role: worker
- role: worker
- role: worker
- role: worker
- role: worker
```

```bash
kind create cluster --name k1 --config ~/config/kind/kind.config
kubectl label nodes spotnode1 node.kubernetes.io/capacity=spot
kubectl label nodes on-demandnode1 node.kubernetes.io/capacity=on-demand
```

#### 测试用例

> 部署一个单副本应用

```bash
kubectl apply -f examples/onereplicas-test-mix-scheduler.yaml
```

> 部署一个十副本应用

```bash
kubectl apply -f examples/tenreplicas-test-mix-scheduler.yaml
```

> 部署一个十副本应用, mix-scheduler-plugins/availability-guarantee = 2

```bash
kubectl apply -f examples/availability2-tenreplicas-test-mix-scheduler.yaml
```
