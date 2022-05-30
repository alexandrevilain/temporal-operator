# Changelog

All notable changes to this project are documented in this file.

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
