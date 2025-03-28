# AzTUI

[![Go](https://github.com/brendank310/aztui/actions/workflows/go.yml/badge.svg)](https://github.com/brendank310/aztui/actions/workflows/go.yml)

**aztui** is an open-source command-line tool that allows users to interact with their Azure resources using a text user interface (TUI). It provides a convenient way to manage Azure resources directly from the terminal.

## Features

- **Interactive TUI**: Navigate and interact with your Azure resources in a simple, user-friendly terminal interface.
- **Resource Management**: View, create, update, and delete Azure resources such as virtual machines, storage accounts, containers, and more.
- **Authentication**: Integrates with Azure's authentication mechanisms to ensure secure access to your resources.
- **Resource Overview**: Quickly get an overview of your Azure environment, including resource groups, services, and status.
- **Resource Filtering**: Easily filter and search for specific resources based on name, type, or other criteria.
- **Cross-Platform Support**: Works on Linux, macOS, and Windows.

## Instructions

Install the azcli and login using `az login`. Once logged in you can build and run aztui with:
`make all && AZTUI_CONFIG_PATH=conf/default.yaml bin/aztui`

## Demo

![Demonstration](demo.gif)

## Plan for feature development

Mostly the feature development will be determined by how convenient the feature may be to my day-to-day use of Azure, though if there are features that enough people upvote as an issue, I'll try to prioritize them.

## Notice

This is an unofficial project and is not associated with Microsoft. Use it at your own risk.

### Install aztui

To install aztui, run the following commands:

```bash
go install github.com/yourusername/aztui@latest
```

To run from cloudshell, try the following:
```bash
rm -f aztui aztui.zip ./default.yaml && wget https://github.com/brendank310/aztui/releases/download/v0.0.5/aztui.zip && unzip aztui.zip && wget https://raw.githubusercontent.com/brendank310/aztui/refs[...]
```

## Contributors

- [ytimocin](https://github.com/ytimocin)
- [dominikabobik](https://github.com/dominikabobik)
- [brendank310](https://github.com/brendank310)
- [rramankutty0](https://github.com/rramankutty0)
