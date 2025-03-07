# coding=utf-8
# *** WARNING: this file was generated by Pulumi SDK Generator. ***
# *** Do not edit by hand unless you're certain you know what you are doing! ***

import errno
from setuptools import setup, find_packages
from setuptools.command.install import install
from subprocess import check_call


VERSION = "0.0.0"
def readme():
    try:
        with open('README.md', encoding='utf-8') as f:
            return f.read()
    except FileNotFoundError:
        return "tailscale-bastion Pulumi Package - Development Version"


setup(name='lbrlabs_pulumi_tailscalebastion',
      python_requires='>=3.9',
      version=VERSION,
      description="A Pulumi package for creating a tailscale bastion in AWS.",
      long_description=readme(),
      long_description_content_type='text/markdown',
      keywords='aws tailscale lbrlabs kind/component category/network',
      project_urls={
          'Repository': 'https://github.com/lbrlabs/pulumi-tailscale-bastion'
      },
      packages=find_packages(),
      package_data={
          'lbrlabs_pulumi_tailscalebastion': [
              'py.typed',
              'pulumi-plugin.json',
          ]
      },
      install_requires=[
          'parver>=0.2.1',
          'pulumi>=3.0.0,<4.0.0',
          'pulumi-aws>=6.0.0,<7.0.0',
          'pulumi-azure>=6.0.0,<7.0.0',
          'pulumi-kubernetes>=4.0.0,<5.0.0',
          'pulumi-tailscale>=0.0.0,<1.0.0',
          'pulumi-tls>=5.0.0,<6.0.0',
          'semver>=2.8.1',
          'typing-extensions>=4.11,<5; python_version < "3.11"'
      ],
      zip_safe=False)
