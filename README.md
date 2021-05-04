# Continuous Traceability Monitor (CTM) 

[![license](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0) [![REUSE status](https://api.reuse.software/badge/github.com/SAP/quality-continuous-traceability-monitor)](https://api.reuse.software/info/github.com/SAP/quality-continuous-traceability-monitor) [![CircleCI](https://circleci.com/gh/SAP/quality-continuous-traceability-monitor/tree/master.svg?style=svg)](https://circleci.com/gh/SAP/quality-continuous-traceability-monitor/tree/master) [![Go Report Card](https://goreportcard.com/badge/github.com/SAP/quality-continuous-traceability-monitor)](https://goreportcard.com/report/github.com/SAP/quality-continuous-traceability-monitor)

#### Why?
CTM allows you to continuously monitor the health of your product's requirements.
Test results often miss a link to the project's requirements so that it's cumbersome to understand which requirement is affected by a broken test.  
Adding CTM to e.g. your Continuous Delivery pipeline will give the complete team an always up to date view on the quality status of your project's requirements!  
Read also our blog: [Don’t monitor the health of your test cases, monitor the health of your product requirements!](https://blogs.sap.com/2019/03/07/dont-monitor-the-health-of-your-test-cases-monitor-the-health-of-your-product-requirements/) on the rational of CTM

#### How?
CTM creates a linkage between your product's requirements, its automated tests and their test results.  
The generated reports may help you e.g. provide traceability in respect to the ISO9001 requirement "Identification and traceability".  
Further more you may use CTM to store a continuous trace of your automated test results per product requirement in a so called [traceability repository](https://github.com/SAP/quality-continuous-traceability-monitor/wiki/CTM-Guidebook#9-traceability-repository). This will grant you a constant insight on your product quality from a requirements point of view. 

![CTM Motivation](https://github.com/SAP/quality-continuous-traceability-monitor/wiki/assets/images/CTM_Motivation.jpg)

## Requirements

To trace your product's requirements, they have to be maintained in either 
  * [Atlassian JIRA](https://www.atlassian.com/software/jira) 
  * [GitHub](https://github.com/)
  * [Enterprise GitHub](https://enterprise.github.com/home)
  
Your automated test results (e.g. provided by your test runner) must be available in
   * xunit XML (see [XSD Schema](http://help.catchsoftware.com/display/ET/JUnit+Format))

## Installation

#### Stable release

Please ensure you have a working [latest version of go installed](https://golang.org/doc/install). 

Get the latest version of CTM via `go get`
```
go get github.com/SAP/quality-continuous-traceability-monitor
```

## Getting Started

See our [Getting Started Guide](https://github.com/SAP/quality-continuous-traceability-monitor/wiki/Getting-Started), to learn how you can use CTM.

## Configuration

Check our [CTM Guidebook](https://github.com/SAP/quality-continuous-traceability-monitor/wiki/CTM-Guidebook) to learn about the various options you could use CTM to improve your code quality.

## How to obtain help

In case of troubles with CTM, please [file an issue](https://github.com/SAP/quality-continuous-traceability-monitor/issues) and we'll try to help you. 

## Contributing
Read and understand our [contribution guidelines](https://github.com/SAP/quality-continuous-traceability-monitor/blob/master/CONTRIBUTING.md) before opening a pull request.

## Licensing
Please see our [LICENSE](https://github.com/SAP/quality-continuous-traceability-monitor/blob/master/LICENSE) for copyright and license information. Detailed information including third-party components and their licensing/copyright information is available via the [REUSE tool](https://api.reuse.software/info/github.com/SAP/quality-continuous-traceability-monitor).
