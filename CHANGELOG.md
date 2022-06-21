# Changelog

All notable changes to this project are documented in this file.

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
