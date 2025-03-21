# KGENT-API

A Kubernetes API client wrapper and example collection that provides simplified access to Kubernetes clusters through a REST API.

## Overview

KGENT-API is a Go-based application that wraps Kubernetes client libraries and exposes them through a RESTful API using the Gin framework. It demonstrates different approaches to interacting with Kubernetes resources and provides examples of various Kubernetes client implementations.

## Features

- RESTful API for Kubernetes resource management
- Examples of various Kubernetes client implementations:
  - ClientSet
  - DynamicClient
  - RestClient
  - DiscoveryClient
- Informer patterns and implementations
- RestMapper functionality for resource discovery
- Pod logs and events retrieval
- CORS support for web frontend integration

## Project Structure

```
├── api/                     # Main API implementation
│   ├── config/              # Kubernetes configuration setup
│   ├── controllers/         # API endpoint controllers
│   ├── services/            # Business logic services
│   └── kapi.go               # Main API entry point
├── clients/                 # Kubernetes client examples
│   ├── ClientSet/           # ClientSet example
│   ├── DynamicClient/       # DynamicClient example
│   ├── DiscoveryClient/     # DiscoveryClient example
│   └── RestClient/          # RestClient example
├── informer/                # Kubernetes informer examples
├── restmapper/              # RestMapper examples
└── go.mod                   # Go module definition
```

## Prerequisites

- Go 1.23 or later
- Access to a Kubernetes cluster
- Kubernetes configuration (kubeconfig)

## Installation

1. Clone the repository:
   ```
   git clone https://github.com/yourusername/kgent-api.git
   cd kgent-api
   ```

2. Install dependencies:
   ```
   go mod download
   ```

## Usage

### Running the API Server

```
go run api/kapi.go
```

The server will start on port 8000 by default. You can set a custom port using the `PORT` environment variable.

### API Endpoints

- **GET /health**: Health check endpoint
- **GET /api/v1/resources/:resource**: List resources of a specific type
- **DELETE /api/v1/resources/:resource**: Delete a specific resource
- **POST /api/v1/resources/:resource**: Create a new resource
- **GET /api/v1/resources/gvr**: Get GroupVersionResource information
- **GET /api/v1/pods/logs**: Get pod logs
- **GET /api/v1/pods/events**: Get pod events

### Running Client Examples

Each example in the `clients` directory can be run separately:

```
# ClientSet example
go run clients/ClientSet/clientset.go --namespace=kube-system

# DynamicClient example
go run clients/DynamicClient/dynamicclient.go --namespace=kube-system --group=apps --version=v1 --resource=deployments

# DiscoveryClient example
go run clients/DiscoveryClient/discoveryclient.go --resources

# RestClient example
go run clients/RestClient/restclient.go --namespace=kube-system
```

### Running Informer Example

```
go run informer/informer.go --type=all --namespace=default
```

### Running RestMapper Example

```
go run restmapper/restmapper.go --namespace=kube-system --resource=pods
```

## Kubernetes Client Types

The project demonstrates various client types for interacting with Kubernetes:

1. **ClientSet**: Typed client for core Kubernetes resources
2. **DynamicClient**: Generic client for any Kubernetes resource
3. **RestClient**: Low-level HTTP client for Kubernetes APIs
4. **DiscoveryClient**: Client for discovering API resources
