# Banjo AzTUI
[![Go](https://github.com/brendank310/aztui/actions/workflows/go.yml/badge.svg)](https://github.com/brendank310/aztui/actions/workflows/go.yml)

A work in progress text user interface for controlling your Azure resources. Code is currently a mess, but the plan is clean it up and add functionality as time goes on.

# Instructions

Install the azcli and login using `az cli`. Once logged in you can build and run banjo with:
`make all && bin\banjo`

# Demo

![Demonstration](demo.gif)

# Plan for feature development
Mostly the feature development will be determined by how convenient the feature may be to my day to day use of Azure, though if there are features that enough people upvote as an issue I'll try to prioritize it.

## Todos:
* Add the ability to parse the az cli output for the VM command, to quickly expand the operations available to use on a VM, using a forked process for now (can add API based calls later).

# Notice

I (brendank310) do work for Microsoft on Azure, but this is a personal project done in my spare time as a tool for my convenience (I hope this is a bit more responsive than the web portal), as well as learning the tview library for other personal projects related to the [beepy device](https://beepy.sqfmi.com/).
