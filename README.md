# TP-Link Tapo P110 Prometheus exporter

## Description

This is an TP-Link Tapo P110 Prometheus exporter written in Go. It collects the data from P110 devices and sends them to
Prometheus, where they can be visualized in Grafana.

This exporter handles 403 errors that are randomly are received from P110 devices.

### Dashboard

![dashboard_screen.png](img/dashboard_screen.png)

## Installation via docker compose

1. `cd docker-compose`
2. Update `TAPO_EMAIL` and `TAPO_PASSWORD` in `docker-compose.yml` or use environment variables (preferable)
3. Update `config.json` and add P110 devices with meaningful names and their IP addresses (or host:port, see below how
   to configure everything to work via Internet)
4. Run `docker compose up` or `docker compose up -d`

The exporter is running on port 8086, Prometheus on port 9090 and Grafana on port 3000.

Go to Grafana (http://localhost:3000, admin/admin), configure `Prometheus` datasource (url should
be http://prometheus:9000).
In Grafana import dashboard `21982` and selected the `Prometheus` datasource:

![import_dashboard.png](img/import_dashboard.png)

## How to use the exporter via Internet

All existing solutions require the server to be located in the same network as P110 devices, since we send requests
directly to them. There is a workaround how to run this setup on external server (some VPS for instance).

Here is an example how I did it using ASUS router.

1. Configure static IP address for P110 device (in my example it's `192.168.50.50`). You can select any local IP
   address, it doesn't matter
2. Configure port forwarding from external PORT (let's say `9876` to `192.168.50.50:80`
   ![port_forwarding.png](img/port_forwarding.png)
3. Configure DDNS in your ASUS router (I assume other routers supports similar feature) so we have a dedicated hostname
   attached to our router. Can be skipped if you have static IP address from your provider. Let's say it's `example.asuscomm.com`
4. Now in `config.json` (see step #3 from installation instruction) you can use:

```json
{
  "devices": [
    {
      "name": "Fridge",
      "ip_address": "example.asuscomm.com:9876"
    }
  ]
}
```
(pay attention I set port which we configured in step#2)

instead of 
```json
{
  "devices": [
    {
      "name": "Fridge",
      "ip_address": "192.168.50.50"
    }
  ]
}
```
5. Deploy everything on external VPS.