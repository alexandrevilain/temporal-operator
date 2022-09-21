# Changelog

All notable changes to this project are documented in this file.

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
