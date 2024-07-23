# Background

In all cloud providers, like AWS, Google, and others, there are many spot instances. They are quite cheap (10% of the on-demand instances' price), but after you buy them, they could be terminated with only two minutes' notice in advance (in most scenarios, we don't set PDB, and we should perform the graceful drain).

So, I want you to design a strategy to maximize the use of spot instances without causing service interruptions, instead of relying solely on on-demand instances, to cut costs, by using distributed scheduling in a single cluster (on-demand/spot mixed or other methods for one workload). This is important because all spot instances being terminated at the same time could cause interruptions for different kinds of workloads (single replica workload, multiple replica workload).

Also, I don't want to change the scheduler already used in the K8s cluster and want to ensure the minimal components necessary in the cluster.

Notes:

> 1. On demand nodes has label: node.kubernetes.io/capacity: on-demand.
> 2. Spot node has label: node.kubernetes.io/capacity: spot.
> 3. Workloads represented as Deployments and StatefulSets.
> 4. on-demand/spot instance represented as K8s nodes in the cluster.
> 5. Only focus on scheduling control; the graceful drain after receiving the terminal notification is handled by other components.


# 设计思路

- spot实例比较适合灵活性较高或具有容错性的应用程序
- 为了保证服务的高可用性就需要在on-demand节点(非spot节点)上保持一定量应用pod, 并且在spot节点上的pod尽量分散节点部署, 避免单点spot节点下线导致的短时压力飙升, 过于加大其他pod的压力, 降低服务的可用性
- 在未特殊配置的情况下, 尽量保证应用的大部分pod会分散部署在不同的spot节点上
- 支持自定义可用性保证

# 快速开始


