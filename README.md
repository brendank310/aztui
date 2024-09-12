# AzTUI
[![Go](https://github.com/brendank310/aztui/actions/workflows/go.yml/badge.svg)](https://github.com/brendank310/aztui/actions/workflows/go.yml)

**aztui** is an open-source command-line tool that allows users to interact with their Azure resources using a text user interface (TUI). It provides a convenient way to manage Azure resources directly from the terminal, making it easy to monitor, manage, and perform actions on your resources without needing to use the Azure web portal.

## Features

- **Interactive TUI**: Navigate and interact with your Azure resources in a simple, user-friendly terminal interface.
- **Resource Management**: View, create, update, and delete Azure resources such as virtual machines, storage accounts, containers, and more.
- **Authentication**: Integrates with Azure's authentication mechanisms to ensure secure access to your resources.
- **Resource Overview**: Quickly get an overview of your Azure environment, including resource groups, services, and status.
- **Resource Filtering**: Easily filter and search for specific resources based on name, type, or other criteria.
- **Cross-Platform Support**: Works on Linux, macOS, and Windows.

## Instructions

Install the azcli and login using `az cli`. Once logged in you can build and run aztui with:
`make all && bin\aztui`

## Demo

![Demonstration](demo.gif)

## Plan for feature development
Mostly the feature development will be determined by how convenient the feature may be to my day to day use of Azure, though if there are features that enough people upvote as an issue I'll try to prioritize it.

## Notice

I (brendank310) do work for Microsoft on Azure, but this is a personal project done in my spare time as a tool for my convenience (I hope this is a bit more responsive than the web portal), as well as learning the tview library for other personal projects related to the [beepy device](https://beepy.sqfmi.com/).

### Install aztui

To install aztui, run the following commands:

```bash
go install github.com/yourusername/aztui@latest
```
