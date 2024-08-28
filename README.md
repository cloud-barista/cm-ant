[![Release Version](https://img.shields.io/github/v/release/cloud-barista/cm-ant?color=blue)](https://github.com/cloud-barista/cm-ant/releases/latest)
[![Swagger API Doc](https://img.shields.io/badge/API%20Doc-Swagger-brightgreen)](https://cloud-barista.github.io/api/?url=https://raw.githubusercontent.com/cloud-barista/cm-ant/main/api/swagger.yaml)
[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/cloud-barista/cm-ant?label=go.mod)](https://github.com/cloud-barista/cm-ant/blob/main/go.mod)
[![License](https://img.shields.io/github/license/cloud-barista/cm-ant?color=blue)](https://github.com/cloud-barista/cm-ant/blob/main/LICENSE)

# CM-ANT Cloud Migration Validation Framework

```text
🧨 [WARNING]
🧨 CM-ANT is currently under development.
🧨 So, we do not recommend using the current release in production.
🧨 Please note that the functionalities of CM-ANT are not stable and secure yet.
🧨 If you have any difficulties in using CM-ANT, please let us know.
🧨 (Open an issue or Join the Cloud-Migrator Slack)
```

---

# Overview
The Cloud Migration Validation Framework is designed to validate the performance, pricing, and cost-effectiveness before and after the cloud migration process (hereafter referred to as migration).

It provides two main categories of functionality:

- Predication of cloud infrastructure transtion pricing and validation of resource usage cost.
- On-demand performance evaluation and validation of the target cloud infrastructure.

### Predication of cloud infrastructure transtion pricing and validation of resource usage cost
This feature provides:

- Price information for the recommended or targeted infrastructure specifications before the migration begins.
- Operational cost information for specific CSPs (Cloud Service Providers).
- Predicted cost information.

### On-demand performance evaluation and validation of the target cloud infrastructure
This feature provides:

- Performance evaluation of applications operating on the migrated infrastructure.
- Performance validation information based on the evaluation results.

These functionalities are integrated with other subsystems, namely `CB-Tumblebug` and `CB-Spider`, to function properly. Therefore, for CM-ANT to operate correctly, the related subsystems must be running on the same environment.


---

# Index 📖

1. [Prerequisites 📝](#prerequisites-)
2. [How to Run 🚀](#how-to-run-)
3. [Usage Configuration ⚙️](#usage-configuration-)
4. [How to Use 🔍](#how-to-use-)

---

## Prerequisites 📝

### Envionment
- OS: Ubuntu 22.04
- Language: Go 1.23.0
- Container: Docker 25.0.0

### Subsystem Dependency
- CB-Spider : v0.9.0 <- **cost explorer anycall handler not yet implemented version.**
- CB-Tumblebug : v0.9.7

---

## How to Run 🚀

### 1) Download CM-ANT 🐜
Clone the CM-ANT Repository from github.
```bash
git clone https://github.com/cloud-barista/cm-ant.git
```

### 2) Start related Subsystem

```bash
cd cm-ant
docker compose up -d

⠧ Network cm-ant_cm-ant-net         Created        31.7s 
⠧ Network cm-ant_cb-tumblebug-net   Created        31.7s 
⠦ Network cm-ant_cb-spider-net      Created        31.6s 
✔ Container cm-ant-ant-postgres-1   Healthy        31.1s 
✔ Container cb-tumblebug-etcd       Started        1.2s 
✔ Container cb-spider               Started        1.3s 
✔ Container cb-tumblebug            Started        2.0s 
✔ Container cm-ant                  Started        31.4s 
```

---

##  Usage Configuration ⚙️
Using CM-ANT independently comes with some limitations.  \
To fully utilize all the features offered by CM-ANT, you need to use functionalities provided by various subsystems. \
This means that there is a dependency on other subsystems, and proper user configuration is required to correctly use the features provided by these subsystems.

### User credential registration  ⭐⭐
In CM-ANT, it is necessary to register user credentials for each CSP. Registered user's CSP credentials are used for tasks such as provisioning virtual machines in a remote environment during performance evaluations, or for retrieving price or cost information from CSP.

Among the subsystems used by CM-ANT, CB-TUMBLEBUG provides a user-friendly process for registering and storing multi-cloud information. It is recommended to register user credentials using the credential registration method provided by CB-TUMBLEBUG.


Follow the guide for initializing CB-Tumblebug to configure multi-cloud information.

> 👉 [Initialize CB-Tumblebug to configure Multi-Cloud info](https://github.com/cloud-barista/cb-tumblebug?tab=readme-ov-file#3-initialize-cb-tumblebug-to-configure-multi-cloud-info)

### Pre-Configuration for Performance Evaluation ⭐⭐
To correctly use the performance evaluation features provided by CM-ANT, the following steps are required:

- Register appropriate permissions for VM provisioning with the registered credentials. [TBD]

### Pre-Configuration for Price and Cost Features ⭐⭐
To correctly use the  price and cost features provided by CM-ANT, the following steps are required:

- Enable AWS Cost Explorer and set up daily granularity resource-level data.
- Register appropriate permissions for price and cost retrieval with the registered credentials. [TBD]



#### Enable AWS Cost Explorer
1. Open the [Cost Explorer page](https://console.aws.amazon.com/cost-management/home) in the AWS Management Console.
2. If Cost Explorer is already enabled, you can view the cost information used.
3. If Cost Explorer is not enabled, select "Launch Cost Explorer" on the Cost Explorer start page.



#### Enable Cost Explorer Resource-level data at daily granularity
1) Navigate to the enabled [Cost Explorer page](https://console.aws.amazon.com/cost-management/home).
2) In the left navigation pane, go to the Preferences & Settings > Cost Management Preferences tab.
3) On the central screen, go to the tab labeled Cost Explorer.
4) Check the box for Detailed Data > Daily granularity resource-level data.
5) Select the services for resource-level identification provided by CM-ANT.
    - Cost Explorer, EC2-Others, EC2-Instance, VPC, Tax

---

## How to Use 🔍
#### 👉 [CM-ANT Swagger API Doc](https://cloud-barista.github.io/api/?url=https://raw.githubusercontent.com/cloud-barista/cm-ant/main/api/swagger.yaml)
[TBD]





