job "temporal" {
  datacenters = ["dc1"]
  type = "service"
  namespace = "dev"

  group "temporal" {
    count = 1

    update {
      max_parallel     = 1
      canary           = 1
      auto_revert      = true
      auto_promote     = true
      health_check     = "checks"
    }

    network {
      mode = "bridge"
      port "frontend" {
        to = 7233
        static = 8233
      }

      port "ui" {
        to = 8080
      }
    }

    service {
      name = "dev-temporal-frontend"
      port = 7233
      connect {
        sidecar_service {}
      }
      
      check {    
        type            = "grpc"
        port            = "frontend"
        interval        = "5s"
        timeout         = "2s"
        grpc_service    = "temporal.api.workflowservice.v1.WorkflowService"
        grpc_use_tls    = false
      }
    }

    service {
      name = "dev-temporalui"
      port = "ui"
      tags = [
        "traefik.enable=true",
        "traefik.http.routers.dev-temporalui.rule=Host(`temporal-dev.mgnt.fynbos.dev`)"
      ]
    }

    service {
      connect {
        sidecar_service {
          proxy {
            upstreams {
              destination_name = "dev-postgres"
              local_bind_port  = 5432
            }
          }
        }
      }
    }

    task "auto-setup" {
      driver = "docker"
      kill_timeout = "120s"

      config {
        image = "registry.gitlab.com/fynbos/fynbos/temporal-auto-setup:1.18.5"

        mounts = [{
          type     = "bind"
          source   = "secrets"
          target   = "/etc/temporal/config"
          readonly = false
        }]
      }

      template {
        data = <<EOF
        DB=postgresql
        SKIP_DB_CREATE=true
        DBNAME=temporal
        VISIBILITY_DBNAME=temporal_visibility
        SKIP_ADD_CUSTOM_SEARCH_ATTRIBUTES=true
        TEMPORAL_CLI_ADDRESS=localhost:7233
        DB_PORT=5432
        {{with secret "database-dev/static-creds/temporal"}}
        POSTGRES_USER={{.Data.username}}
        POSTGRES_PWD={{.Data.password}}
        {{end}}
        {{with secret "database-dev/static-creds/temporal_visibility"}}
        VISIBILITY_POSTGRES_USER={{.Data.username}}
        VISIBILITY_POSTGRES_PWD={{.Data.password}}
        {{end}}
        POSTGRES_SEEDS=127.0.0.1
        SQL_PLUGIN=postgres
        SERVICES=frontend,history,matching,worker
        EOF

        destination = "secrets/file.env"
        env = true
      }

      template {
        data = <<EOF
log:
  stdout: true
  level: "debug,info"

persistence:
  defaultStore: default
  visibilityStore: visibility
  numHistoryShards: 512
{{with secret "database-dev/static-creds/temporal"}}
  datastores:
    default:
      sql:
        pluginName: "postgres"
        driverName: "postgres"
        databaseName: "temporal"
        connectAddr: "localhost:5432"
        connectProtocol: "tcp"
        user: "{{.Data.username}}"
        password: "{{.Data.password}}"
        maxConnLifetime: 1h
        maxConns: 20
        secretName: ""
        tls:
          enabled: false
{{end}}
{{with secret "database-dev/static-creds/temporal_visibility"}}
    visibility:
      sql:
        pluginName: "postgres"
        driverName: "postgres"
        databaseName: "temporal_visibility"
        connectAddr: "localhost:5432"
        connectProtocol: "tcp"
        user: "{{.Data.username}}"
        password: "{{.Data.password}}"
        maxConnLifetime: 1h
        maxConns: 20
        tls:
          enabled: false
{{end}}
global:
  membership:
    name: temporal
    maxJoinDuration: 30s
    broadcastAddress: "0.0.0.0"

  pprof:
    port: 7936
    
  metrics:
    tags:
      type: frontend
    prometheus:
      timerType: histogram
      listenAddress: "0.0.0.0:9090"

services:
  frontend:
    rpc:
      grpcPort: 7233
      membershipPort: 6933
      bindOnIP: "0.0.0.0"

  history:
    rpc:
      grpcPort: 7234
      membershipPort: 6934
      bindOnIP: "0.0.0.0"

  matching:
    rpc:
      grpcPort: 7235
      membershipPort: 6935
      bindOnIP: "0.0.0.0"

  worker:
    rpc:
      grpcPort: 7239
      membershipPort: 6939
      bindOnIP: "0.0.0.0"
clusterMetadata:
  enableGlobalDomain: false
  failoverVersionIncrement: 10
  masterClusterName: "active"
  currentClusterName: "active"
  clusterInformation:
    active:
      enabled: true
      initialFailoverVersion: 1
      rpcName: "temporal-frontend"
      rpcAddress: "127.0.0.1:7933"

dcRedirectionPolicy:
  policy: "noop"
  toDC: ""
archival:
  status: "disabled"

publicClient:
  hostPort: "127.0.0.1:7233"

dynamicConfigClient:
  filepath: "/etc/temporal/config/dynamic_config.yaml"
  pollInterval: "10s"
        EOF

        destination = "secrets/config_template.yaml"
      }

      template {
        data = <<EOF
limit.maxIDLength:
  - value: 255
    constraints: {}
    EOF
        destination = "secrets/dynamic_config.yaml"
      }

      vault {}

      resources {
        cpu = 500
        memory = 2048
      }
    }

    task "ui" {
      driver = "docker"
      kill_timeout = "120s"

      config {
        image = "temporalio/ui:2.9.0"
      }

      template {
        data = <<EOH
        TEMPORAL_GRPC_ENDPOINT = "127.0.0.1:7233"
        TEMPORAL_PERMIT_WRITE_API = true
        EOH

        destination = "local/file.env"
        env         = true
      }
    }
  }
}
