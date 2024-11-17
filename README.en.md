![GitHub License](https://img.shields.io/github/license/ZiRunHua/LeapLedger)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/ZiRunHua/LeapLedger)
![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/ZiRunHua/LeapLedger/CI.yml)
![GitHub Release](https://img.shields.io/github/v/release/ZiRunHua/LeapLedger)
![Docker Pulls](https://img.shields.io/docker/pulls/xiaozirun/leap-ledger)
[![Go Report Card](https://goreportcard.com/badge/github.com/ZiRunHua/LeapLedger)](https://goreportcard.com/report/github.com/ZiRunHua/LeapLedger)
![GitHub stars](https://img.shields.io/github/stars/ZiRunHua/LeapLedger?style=social)

<h1 align="center">LeapLedger</h1>
<p align="center">
  <a href="docs/README.en.md">English</a> | <a href="README.md">简体中文</a>
</p>

`LeapLedger` is a free and open-source accounting software with a decoupled front-end and back-end. Powered by `Flutter`, it provides a smooth and seamless user experience, easily extending to iOS, Mac, and Windows in the future. The server is built using the `Gin` framework, implemented with `NATS`, `MySQL`, and `Redis`, offering fast responses and stable services. Deployment and building of client installation packages can be quickly achieved using `Docker`.

<table>
  <tr>
    <td align="center"><img src="https://github.com/user-attachments/assets/e5151e7a-6b1f-4903-b4f1-8ffdc20c1b46" alt="Description" width="150"></td>
    <td align="center"><img src="https://github.com/user-attachments/assets/03dce625-a340-4aa5-92fd-ac4e59ee18b9" alt="Description" width="150"></td>
    <td align="center"><img src="https://github.com/user-attachments/assets/fd19053c-a469-4fcd-9d1e-9371c094039c" alt="Description" width="150"></td>
    <td align="center"><img src="https://github.com/user-attachments/assets/4d605f41-18fc-41b0-bbdf-d50ae1ecc550" alt="Description" width="150"></td>
    <td align="center"><img src="https://github.com/user-attachments/assets/0579110f-66b5-4739-9cc7-bcaeef4e246f" alt="Description" width="150"></td>
  </tr>
</table>

## Client
Flutter client project transfer: [LeapLedger-App](https://github.com/ZiRunHua/LeapLedger-App)

Download the latest Android installation package: [v1.0.0](https://github.com/ZiRunHua/LeapLedger-App/releases/tag/v1.0.0). Data is periodically deleted, please deploy the server for usage.

## Table of Contents

- [What We Have](#what-we-have)
    - [Features](#features)
    - [API Services](#api-services)
- [Running/Deployment](#runningdeployment)
    - [Build Image](#build-image)
- [API Documentation](#api-documentation)
- [Protocol](#protocol)
- [Contributions](#contributions)
- [Contact Me](#contact-me)
- [Acknowledgements](#acknowledgements)

## What We Have
* :iphone: A powerful architecture based on Flutter, providing a smooth user experience, easily extendable to iOS, Mac, and Windows in the future.

* :whale: Whether it's server deployment or client packaging, everything is handled through Docker. :parasol_on_ground:

### Features
* :family_man_woman_girl_boy: Shared ledgers that sync records with your partner and family while maintaining independent bookkeeping.

* :timer_clock: Scheduled bookkeeping for fixed monthly expenses like rent and communication fees—no more manual entry hassle!

* :credit_card: Only want to handle the accounts once a month? No problem, we can import bills from Alipay and WeChat.

In addition to these, we also have:
* :books: Multi-ledger management and :earth_africa: timezone-ledgers.
* Clear views to understand recent situations.
* Easily browse records and charts. :bar_chart:

### API Services
We have a fast, responsive, and stable API service.

* :shield: Ledger authentication based on the Gin middleware and JWT identity verification ensures security.

* :zap: Asynchronous and event-driven architecture based on NATS provides faster responses.

* :mailbox: The Outbox pattern ensures data consistency while operating asynchronously. :dart:

* :floppy_disk: Dead-letter queues guarantee message retention, enhancing system stability.

* :arrows_counterclockwise: Database updates and initialization based on `Gorm` ensure seamless API service upgrades.

## Running/Deployment
Clone the project:
```bash
git clone https://github.com/ZiRunHua/LeapLedger.git
```
First, start MySQL:
```bash
docker-compose up -d leap-ledger-mysql
```
Check the MySQL logs and wait for ready for connections before executing `docker-compose up -d`:
```bash
docker-compose logs -f leap-ledger-mysql
```
```bash
docker-compose up -d
```
Access http://localhost:8080/public/health to verify the service.

If you don't want to rely on Docker, you can modify the request addresses of mysql, nats, and redis in the [./config.yaml](./config.yaml) file and run it locally

For client packaging details, visit: https://github.com/ZiRunHua/LeapLedger-App

### Build Image

The `docker/Dockerfile` can build an image with a Go compilation environment. However, since Go only needs the binary files to run, you can build a minimal image.

Minimal image:
```bash
docker build -t xiaozirun/leap-ledger:build -f docker/Dockerfile.build .
```
Remote debugging image on port 2345:
```bash
docker build -t xiaozirun/leap-ledger:debug -f docker/Dockerfile.debug .
```
Testing image:
```bash
docker build -t xiaozirun/leap-ledger:test -f docker/Dockerfile.test .
```
Please note the differences in image tags.
## API Documentation

Adopting a RESTful API design style, you can choose to view the documentation in the following formats:
* [ApiFox Documentation](https://apifox.com/apidoc/shared-df940a71-63e8-4af7-9090-1be77ba5c3df)
* [swagger.json](docs/swagger.json)
* [swagger.yaml](docs/swagger.yaml)

## Protocol
The [License](LICENSE) is [GNU Affero General Public License v3](https://www.gnu.org/licenses/agpl-3.0.html)

It can be used for learning or personal purposes, but not for commercial use.
## Contributions
The LeapLedger project is still in its early development stage, and many features and details are continuously being refined.

We welcome contributions in any form, including but not limited to:

* Code contributions: fixing bugs, developing new features, optimizing code, writing tests, etc.
* Issue feedback: submitting bug reports, suggesting improvements, etc.
If you're interested in LeapLedger, feel free to join our community and contribute to its development.

We will develop and adjust features in the `develop` branch, while the `main` branch is for releasing stable versions.

## Contact Me
Email: <a href="mailto:zeng807046079@gmail.com">zeng807046079@gmail.com</a>

## Acknowledgements
Thanks to my friend You for testing, which saved me a lot of effort. I really appreciate it!

## Stargazers over time
[![Stargazers over time](https://starchart.cc/ZiRunHua/LeapLedger.svg?variant=adaptive)](https://starchart.cc/ZiRunHua/LeapLedger)