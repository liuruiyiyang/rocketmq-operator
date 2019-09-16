/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package constants

const BrokerClusterPrefix = "broker-cluster-"
const MasterBrokerContainerNamePrefix = "broker-master"
const ReplicaBrokerContainerNamePrefix = "broker-replica"
const AdminToolDir = "/home/rocketmq/rocketmq-4.5.0/bin/mqadmin"
const UpdateBrokerConfig  = "updateBrokerConfig"
const ParamNameServiceAddress = "namesrvAddr"
const EnvNameServiceAddress = "NAMESRV_ADDR"
const EnvReplicationMode = "REPLICATION_MODE"
const EnvBrokerId = "BROKER_ID"
const EnvBrokerClusterName = "BROKER_CLUSTER_NAME"
const EnvBrokerName = "BROKER_NAME"
const LogMountPath = "/home/rocketmq/logs"
const StoreMountPath = "/home/rocketmq/store"
const LogSubPathName = "logs"
const StoreSubPathName = "store"
const NameServiceMainContainerPortNumber  = 9876
const NameServiceMainContainerPortName = "main"
const BrokerVipContainerPort  = 10909
const BrokerVipContainerPortName = "vip"
const BrokerMainContainerPort  = 10911
const BrokerMainContainerPortName = "main"
const BrokerHighAvailabilityContainerPort  = 10912
const BrokerHighAvailabilityContainerPortName = "ha"

const StorageModeNFS  = "NFS"
const StorageModeEmptyDir  = "EmptyDir"
const StorageModeHostPath  = "HostPath"

const CheckRunningPodIntervalInSecond  = 1



