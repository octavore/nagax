# Naga Xtras

This is a collection of modules I have found useful in the course of building apps using [Naga](https://github.com/octavore/naga).

## Config

Config module simplifies loading of configuration options from an external JSON file, by default `config.json` but overridable using the `CONFIG_FILE` environment variable.

## Logger

Logger module provides a shared logger interface and a single point for hooking in your custom logging backend or capturing of Error log messages.

## Migrate

Migrate module handles SQL migrations, both postgres and mysql.

## Static

Static module is serves static files and supports embedding assets with [packr](https://github.com/gobuffalo/packr).