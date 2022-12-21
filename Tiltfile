# -*- mode: Python -*-

load('ext://cert_manager', 'deploy_cert_manager')

def build_image():
    docker_build(
      "localhost:5001/temporal-operator",
      ".",
      ignore=[
        ".git",
        ".github",
        "bundle",
        "docs",
        "examples",
        "hack",
        "out",
        "tests",
        "*.md",
        "LICENSE",
        "PROJECT",
        ]
    )

def deploy():
    k8s_yaml(
        kustomize('./config/local')
    )

build_image()
deploy_cert_manager(version="v1.10.1")
deploy()