# Changelog

All notable changes to this project are documented in this file.

## 0.11.1

**Release date:** 2023-02-12

Fixes:
- Fix database tls certificates mounts when database with tls enabled [#267](https://github.com/alexandrevilain/temporal-operator/pull/267)

Updates:
- Bump docker/build-push-action from 3 to 4 [#258](https://github.com/alexandrevilain/temporal-operator/pull/258)
- Bump istio.io/client-go from 1.16.1 to 1.16.2 [#259](https://github.com/alexandrevilain/temporal-operator/pull/259)
- Bump github.com/onsi/ginkgo/v2 from 2.7.0 to 2.8.0 [#260](https://github.com/alexandrevilain/temporal-operator/pull/260)
- Bump sigs.k8s.io/cluster-api from 1.3.2 to 1.3.3 [#262](https://github.com/alexandrevilain/temporal-operator/pull/262)

## 0.11.0

**Release date:** 2023-01-29

Features:
- Support cross-namespace mTLS for TemporalWorkerProcess & TemporalClusterClient [#247](https://github.com/alexandrevilain/temporal-operator/pull/247)
- Add support for namespace deletion [#251](https://github.com/alexandrevilain/temporal-operator/pull/251)
- Add support for dynamic config [#255](https://github.com/alexandrevilain/temporal-operator/pull/255)

Improvements:
- Add operator logs at e2e tests end to simplify debugging + stop kustomize throttling on github actions [#238](https://github.com/alexandrevilain/temporal-operator/pull/238)
- Updated containerImage and replaces field in CSV for 0.10.0 [#240](https://github.com/alexandrevilain/temporal-operator/pull/240)

Updates:
- Bump sigs.k8s.io/cluster-api from 1.3.1 to 1.3.2 [#242](https://github.com/alexandrevilain/temporal-operator/pull/242)
- Bump github.com/onsi/ginkgo/v2 from 2.6.1 to 2.7.0 [#243](https://github.com/alexandrevilain/temporal-operator/pull/243)
- Bump github.com/cert-manager/cert-manager from 1.10.1 to 1.11.0 [#245](https://github.com/alexandrevilain/temporal-operator/pull/245)
- Bump go.temporal.io/server from 1.19.0 to 1.19.1 [#248](https://github.com/alexandrevilain/temporal-operator/pull/248)
- Bump go.temporal.io/sdk from 1.19.0 to 1.20.0 [#249](https://github.com/alexandrevilain/temporal-operator/pull/249)
- Bump github.com/onsi/gomega from 1.24.2 to 1.25.0 [#250](https://github.com/alexandrevilain/temporal-operator/pull/250)
- Bump go.temporal.io/api from 1.14.0 to 1.15.0 [#253](https://github.com/alexandrevilain/temporal-operator/pull/253)

## 0.10.0

**Release date:** 2023-01-07

⚠️ This is a **breaking 💣** release. The release now requires cert-manager to run.

Improvements:
- Add mutating and validating admission webhooks [#229](https://github.com/alexandrevilain/temporal-operator/pull/229)
- Add support for mTLS enabled clusters in WorkerProcess [#223](https://github.com/alexandrevilain/temporal-operator/pull/223)

Updates:
- Updated ClusterServiceVersion for Operatorhub release v0.9.1 [#231](https://github.com/alexandrevilain/temporal-operator/pull/231)

## 0.9.1

**Release date:** 2022-12-24

Improvements:
- Add local development solution using tilt [#225](https://github.com/alexandrevilain/temporal-operator/pull/225)
- Cleanup APIs discovery codebase [#226](https://github.com/alexandrevilain/temporal-operator/pull/226)

Fixes:
- Add missing rule for servicemonitors in clusterrole [#221](https://github.com/alexandrevilain/temporal-operator/pull/221)
- Fix incorrect conversion between integers [#224](https://github.com/alexandrevilain/temporal-operator/pull/224)

Updates:
- Bump github.com/onsi/ginkgo/v2 from 2.6.0 to 2.6.1 [#219](https://github.com/alexandrevilain/temporal-operator/pull/219)
- Bump sigs.k8s.io/controller-runtime from 0.14.0 to 0.14.1 [#222](https://github.com/alexandrevilain/temporal-operator/pull/222)

## 0.9.0

**Release date:** 2022-12-19

Features:
- Add support for Prometheus scraping through annotations and ServiceMonitor [#201](https://github.com/alexandrevilain/temporal-operator/pull/201)
- Add flags to define namespaces used for istio and cm api checking [#198](https://github.com/alexandrevilain/temporal-operator/pull/198)

Improvements:
- Use Golang 1.19 [#215](https://github.com/alexandrevilain/temporal-operator/pull/215)

Fixes:
- Fix CRDs generation [#220](https://github.com/alexandrevilain/temporal-operator/pull/220)

Updates:
- Bump go.temporal.io/sdk from 1.17.0 to 1.18.1 [#190](https://github.com/alexandrevilain/temporal-operator/pull/190)
- Bump k8s dependencies to 0.25.4 [#194](https://github.com/alexandrevilain/temporal-operator/pull/194)
- Bump go.temporal.io/server from 1.18.4 to 1.18.5 [#195](https://github.com/alexandrevilain/temporal-operator/pull/195)
- Bump github.com/cert-manager/cert-manager from 1.10.0 to 1.10.1 [#196](https://github.com/alexandrevilain/temporal-operator/pull/196)
- Bump github.com/Masterminds/semver/v3 from 3.1.1 to 3.2.0 [#200](https://github.com/alexandrevilain/temporal-operator/pull/200)
- Bump go.temporal.io/server from 1.18.5 to 1.19.0 [#203](https://github.com/alexandrevilain/temporal-operator/pull/203)
- Bump go.temporal.io/sdk from 1.18.1 to 1.19.0 [#206](https://github.com/alexandrevilain/temporal-operator/pull/206)
- Bump helm/kind-action from 1.4.0 to 1.5.0 [#210](https://github.com/alexandrevilain/temporal-operator/pull/210)
- Bump github.com/gocql/gocql from 1.2.1 to 1.3.1 [#211](https://github.com/alexandrevilain/temporal-operator/pull/211)
- Bump istio.io/client-go from 1.16.0 to 1.16.1 [#212](https://github.com/alexandrevilain/temporal-operator/pull/212)
- Bump kubernetes dependencies to v0.26.0 [#213](https://github.com/alexandrevilain/temporal-operator/pull/213)

## 0.8.1

**Release date:** 2022-11-11

Fixes:
- temporal-sql-tool database creation flags for v1.18 have changed [#188](https://github.com/alexandrevilain/temporal-operator/pull/188)

Improvements:
- Disable fail-fast on e2e matrix [#179](https://github.com/alexandrevilain/temporal-operator/pull/179)

Updates:
- Bump go.temporal.io/server from 1.18.3 to 1.18.4 [#182](https://github.com/alexandrevilain/temporal-operator/pull/182)
- Bump github.com/onsi/gomega from 1.23.0 to 1.24.1 [#186](https://github.com/alexandrevilain/temporal-operator/pull/186)

## 0.8.0

**Release date:** 2022-10-29

Features:
- Add support for temporal v1.18.0 [#151](https://github.com/alexandrevilain/temporal-operator/pull/151)
- Add new CRD for TemporalWorkerProcess to manage application workers [#150](https://github.com/alexandrevilain/temporal-operator/pull/150)
- Add builder option to build worker process source code and then deploy [#163](https://github.com/alexandrevilain/temporal-operator/pull/163)
- Added buildAttempt to worker process [#176](https://github.com/alexandrevilain/temporal-operator/pull/176)

Fixes:
- Reduce useless objects updates by making patch updates [#160](https://github.com/alexandrevilain/temporal-operator/pull/160)

Improvements:
- Cleanup TemporalWorkerProcess api and reconciling code [#161](https://github.com/alexandrevilain/temporal-operator/pull/161)
- Cleanup reconcile codebase [#171](https://github.com/alexandrevilain/temporal-operator/pull/171)
- Add version to status so we can trigger new build if version in spec is updated [#174](https://github.com/alexandrevilain/temporal-operator/pull/174)

Updates:
- Bump github.com/gosimple/slug from 1.12.0 to 1.13.1 [#148](https://github.com/alexandrevilain/temporal-operator/pull/148)
- Bump go.temporal.io/server from 1.18.0 to 1.18.1 [#152](https://github.com/alexandrevilain/temporal-operator/pull/152)
- Bump k8s dependencies to v0.25.3 [#162](https://github.com/alexandrevilain/temporal-operator/pull/162)
- Bump github.com/cert-manager/cert-manager from 1.9.1 to 1.10.0 [#165](https://github.com/alexandrevilain/temporal-operator/pull/165)
- Bump go.temporal.io/server from 1.18.1 to 1.18.3 [#168](https://github.com/alexandrevilain/temporal-operator/pull/168)
- Bump sigs.k8s.io/e2e-framework from 0.0.7 to 0.0.8 [#169](https://github.com/alexandrevilain/temporal-operator/pull/169)
- Bump github.com/stretchr/testify from 1.8.0 to 1.8.1 [#173](https://github.com/alexandrevilain/temporal-operator/pull/173)
- Bump github.com/onsi/gomega from 1.20.1 to 1.23.0 [#175](https://github.com/alexandrevilain/temporal-operator/pull/175)

## 0.7.0

This is a **breaking 💣** release.

**Release date:** 2022-10-02

Features:
- (breaking 💣) Move GroupVersion to temporal.io and remove "Temporal" prefix in kind [#130](https://github.com/alexandrevilain/temporal-operator/pull/130)
- Add jobTtlSecondsAfterFinished param to spec to control time until jobs are deleted, changed job owner reference from controller to job making it independent. [#135](https://github.com/alexandrevilain/temporal-operator/pull/135)
- Add prometheus scraping endpoint via spec [#141](https://github.com/alexandrevilain/temporal-operator/pull/141)

Fixes:
- Revert kind names to add Temporal prefix [#140](https://github.com/alexandrevilain/temporal-operator/pull/140)
- Don't start new persistence jobs when they have been garbage collected [#142](https://github.com/alexandrevilain/temporal-operator/pull/142)

Improvements:
- Automate the operatorhub bundle creation [#131](https://github.com/alexandrevilain/temporal-operator/pull/131)
- (breaking 💣) Refactor persistence specs to avoid datastores map [#132](https://github.com/alexandrevilain/temporal-operator/pull/132)

Updates:
- Bump helm/kind-action from 1.3.0 to 1.4.0 [#127](https://github.com/alexandrevilain/temporal-operator/pull/127)
- Bump k8s dependencies to v0.25.2 and istio client-go to 1.15.1 [#143](https://github.com/alexandrevilain/temporal-operator/pull/143)

## 0.6.2

**Release date:** 2022-09-22

Fixes:
- Remove liveness probes from being configured for worker since it has no grpc endpoint [#123](https://github.com/alexandrevilain/temporal-operator/pull/123)

Improvements:
- Add manifests to build operator bundle for OLM and operatorhub [#123](https://github.com/alexandrevilain/temporal-operator/pull/123))

## 0.6.1

**Release date:** 2022-09-21

Fixes:
- Update templates and scripts to create default as well as visibility databases [#115](https://github.com/alexandrevilain/temporal-operator/pull/115)

Improvements:
- Clean e2e code base and run persistence tests in parallel [#113](https://github.com/alexandrevilain/temporal-operator/pull/113)
- Clean resources code base [#119](https://github.com/alexandrevilain/temporal-operator/pull/119)

Updates:
- Bump go.temporal.io/api from 1.11.0 to 1.12.0 [#110](https://github.com/alexandrevilain/temporal-operator/pull/110)


## 0.6.0

**Release date:** 2022-09-14

Features:
- Automatic version upgrades [#102](https://github.com/alexandrevilain/temporal-operator/pull/102)

Updates:
- Bump github.com/urfave/cli from 1.22.9 to 1.22.10 [#105](https://github.com/alexandrevilain/temporal-operator/pull/105)

## 0.5.0

**Release date:** 2022-08-29

Features:
- Add istio mTLS provider [#85](https://github.com/alexandrevilain/temporal-operator/pull/85)

Updates:
- Add support for temporal 1.17.2 and use ui v2.5.0 as default [#88](https://github.com/alexandrevilain/temporal-operator/pull/88)
- Bump go.uber.org/zap from 1.21.0 to 1.22.0 [#90](https://github.com/alexandrevilain/temporal-operator/pull/90)
- Bump go.temporal.io/sdk from 1.15.0 to 1.16.0 [#89](https://github.com/alexandrevilain/temporal-operator/pull/89)
- Bump go.uber.org/zap from 1.22.0 to 1.23.0 [#98](https://github.com/alexandrevilain/temporal-operator/pull/98)
- Bump temporal.io server to v1.17.4 [#99](https://github.com/alexandrevilain/temporal-operator/pull/99)
- Bump k8s dependencies to v0.25.0 [#100](https://github.com/alexandrevilain/temporal-operator/pull/100)

## 0.4.0

**Release date:** 2022-08-02

Features:
- Add support for linkerd as mTLS provider [#78](https://github.com/alexandrevilain/temporal-operator/pull/78)
- Add TemporalNamespace CRD to create namespaces on cluster [#81](https://github.com/alexandrevilain/temporal-operator/pull/81)

Improvements:
- Add e2e tests for kubernetes v1.24.0 [#83](https://github.com/alexandrevilain/temporal-operator/pull/83)

Updates:
- Bump github.com/cert-manager/cert-manager from 1.8.2 to 1.9.1 [#80](https://github.com/alexandrevilain/temporal-operator/pull/80)


## 0.3.1

**Release date:** 2022-07-23

Improvements:
-  Remove --disable-cert-manager flag but detect if cert-manager is available at operator setup [#76](https://github.com/alexandrevilain/temporal-operator/pull/76)

Updates:
-  Bump github.com/gocql/gocql from 1.1.0 to 1.2.0 [#62](https://github.com/alexandrevilain/temporal-operator/pull/62)

## 0.3.0

**Release date:** 2022-07-22

This release adds support for mTLS using cert-manager.

Features:
-  Add internode & frontend mTLS using cert-manager [#60](https://github.com/alexandrevilain/temporal-operator/pull/60)

Improvements:
- Add support for temporal v1.17.1 and set UI version to v2.2.1 by default [#70](https://github.com/alexandrevilain/temporal-operator/pull/70)

Updates:
- Bump github.com/stretchr/testify from 1.7.5 to 1.8.0 [#59](https://github.com/alexandrevilain/temporal-operator/pull/59)
- Bump sigs.k8s.io/controller-runtime from 0.12.2 to 0.12.3 [#61](https://github.com/alexandrevilain/temporal-operator/pull/61)
- Bump kubernetes dependencies to v0.24.3 [#73](https://github.com/alexandrevilain/temporal-operator/pull/73)

Fixes:
- Fix misspells [#71](https://github.com/alexandrevilain/temporal-operator/pull/71)
- Add missing RBAC rule for manager role to create events [#74](https://github.com/alexandrevilain/temporal-operator/pull/74)

## 0.2.0

**Release date:** 2022-06-28

This release adds better observability, support for temporal 1.17 and end2end tests.

Features:
- Add support for temporal 1.17.0 [#53](https://github.com/alexandrevilain/temporal-operator/pull/53)
- Add conditions in status reporting and add event recorder [#57](https://github.com/alexandrevilain/temporal-operator/pull/57)

Improvements:
- Bootstrap end2end test suite [#45](https://github.com/alexandrevilain/temporal-operator/pull/45)
- Add mysql persistence end2end test case [#52](https://github.com/alexandrevilain/temporal-operator/pull/52)
- Add cassandra end2end tests [#56](https://github.com/alexandrevilain/temporal-operator/pull/56)

Updates:
- Bump kubernetes dependencies to 0.24.2 [#48](https://github.com/alexandrevilain/temporal-operator/pull/48)
- Bump github.com/gocql/gocql from 1.0.0 to 1.1.0 [#51](https://github.com/alexandrevilain/temporal-operator/pull/51)
- Bump github.com/stretchr/testify from 1.7.2 to 1.7.5 [#55](https://github.com/alexandrevilain/temporal-operator/pull/55)
- Bump sigs.k8s.io/controller-runtime from 0.11.1 to 0.12.2 [#54](https://github.com/alexandrevilain/temporal-operator/pull/54)


## 0.1.1

**Release date:** 2022-06-21

This release is a regression fix release.

Fixes:
- Regression in CRD which makes cassandra required [#46](https://github.com/alexandrevilain/temporal-operator/pull/46)

## 0.1.0

**Release date:** 2022-06-21

This release adds support for Elasticsearch & Cassandra.

Features:
- Add cassandra support [#42](https://github.com/alexandrevilain/temporal-operator/pull/42)
- Add elasticsearch support [#35](https://github.com/alexandrevilain/temporal-operator/pull/35)

Improvements:
- Prune resources when they are disabled [#33](https://github.com/alexandrevilain/temporal-operator/pull/33)
- Add security context on component container and pod [#31](https://github.com/alexandrevilain/temporal-operator/pull/31)

Updates:
- Bump default version of temporalio/ui to 2.0.1 [#36](https://github.com/alexandrevilain/temporal-operator/pull/36)

Fixes:
- Fix rbac for ingresses and services [#34](https://github.com/alexandrevilain/temporal-operator/pull/34)


## 0.0.4

**Release date:** 2022-06-07

This release adds support for UI and admin tools.

Features:
- Add support for webui [#18](https://github.com/alexandrevilain/temporal-operator/pull/18)
- Add support for admin tools [#24](https://github.com/alexandrevilain/temporal-operator/pull/24)

Improvements:
- Improve API documentation [#23](https://github.com/alexandrevilain/temporal-operator/pull/23)
- Create logger adapter for persistence reconciliations [#26](https://github.com/alexandrevilain/temporal-operator/pull/26)

Fixes:
- Packages where not in public, this is now fixed. Sorry for that.

## 0.0.3

**Release date:** 2022-05-30

This release is a fix release. 
The operator was tring to to update the visibility schema with the default schema version (v1.8) which does not exist.

Fixes:
- schema init and update for visibility store [#19](https://github.com/alexandrevilain/temporal-operator/pull/19)

## 0.0.2

**Release date:** 2022-05-30

This release introduces a new way for the operator to reconcile persistence.

Improvements:
- improve persistence reconciliation by relying on the cluster status [#14](https://github.com/alexandrevilain/temporal-operator/pull/14)

## 0.0.1

**Release date:** 2022-05-26

This is the first release of the temporal operator. For now it can create a cluster using postgresSQL as default and visibility datastore.
Many improvements are needed to make it production ready. 
