`nri-mta` monitors Mail Transfer Agent(MTA) send and receive response times synthetically.

### Note
In an environment where there are multiple mail servers serving a single email domain (MX record, load balancer, Active Directory + Exchange, ....) it is not possible to directly measure the send/receive performance of an individual 
mail server _unless_ SMTP and IMAP servers are *directly* addressable. It is a future enhancement to infer performance based on the send/receive path as indicated by an email's `received:` headers.

We can provide the header values as Metrics (and perhaps some day as Events or Spans), however the analysis/visualization is strictly a back-end function. Without significant back-end work these measurements are only moderately useful.

## Installation
Copy the appropriate [Release](releases/)executable to
- Linux `/var/db/newrelic-infra/custom-integrations/nri-mta`
- Windows `C:\Program Files\New Relic\newrelic-infra\newrelic-integrations\nri-mta.exe`
assuming a normal/standard Infrastructure installation.

## Configuration
### nri-mta-config.yml
Copy [`nri-mta-config.sample.yml`](nri-mta-config.sample.yml) to `nri-mta-config.yml` and place in the appropriate directory:
- Linux `/etc/newrelic-infra/integrations.d/`
- Windows `\Program Files\NewRelic\newrelic-infra\inregrations.d`

This is the file that lets Infrastructure know the integration is available. [Standard Infrastructure configuration settings apply.](https://docs.newrelic.
com/docs/infrastructure/host-integrations/infrastructure-integrations-sdk/specifications/host-integrations-standard-configuration-format/)

Additionally, these configuration settings are available under the `env` stanza in `nri-mta-config.yml`:
- `RECEIVE_PATTERNS`: a string containing the full path to the receive pattern file.
- `TIMEOUT`: an integer denoting the number of seconds to wait before a send/receive request times out. (default 30)
- `TRACE_CONFIG`:  a string containing the full path to the trace config file.

### Receive pattern file
The [regular expressions](https://github.com/google/re2/wiki/Syntax) in this file capture the header information that is used to generate Metrics. For [RFC 2821](https://www.rfc-editor.
org/rfc/rfc2821#section-4.4) compliant MTAs the single enabled expression should be sufficient:
```yaml
Received: '^Received:( from (?P<fromhost>.*?)( \((?P<fromip>.*?)\))?)?( by (?P<byhost>.*?)( \((?P<byip>(.*?))?\))?)?( via (?P<via>.*?))?( with (?P<with>.*?))?( id (?P<id>.*?))?( for (?P<for>.*?))?(; (?P<timestamp>.*))'
```
The commented patterns are meant as examples _ONLY_. Full documentation is included in the file.

*NOTE:* multiple patterns with duplicate capture groups _will_ result in duplicate metrics, be careful.

### Trace configuration file
The trace configuration file contains MTA/Client pairs, one for each send/receive test run. The MTA is the system tested, the Client is any mail system that can send and receive. For instance:
```yaml
Processors:
  - MTA:
      Kind: "IMAP"
      SMTPHost: "smtp.office365.com"
      SMTPPort: "587"
      Password: "<Email_User_Password>"
      UserName: "<office365.com_Email_User>"
      IMAPUrl: "outlook.office365.com:993"
    Client:
      Kind: "IMAP"
      SMTPHost:  "smtp.gmail.com"
      SMTPPort: "587"
      IMAPUrl: "imap.gmail.com:993"
      UserName: "<gmail.com_Email_User>"
      Password: "<Email_User_Password>"
```
Currently the only `Kind` available is `IMAP`, at some future point `MSGRAPH` ( [Microsoft's Graph API](https://learn.microsoft.com/en-us/graph/api/resources/message?view=graph-rest-1.0) ) may be supported.

## Outputs

## Troubleshooting

## Building
`go build cmd/nri-mta/nri-mta.go`

## Support
New Relic has open-sourced this project. This project is provided AS-IS WITHOUT WARRANTY OR DEDICATED SUPPORT. Issues and contributions should be reported to the project here on GitHub.

We encourage you to bring your experiences and questions to the [Explorers Hub](https://discuss.newrelic.com) where our community members collaborate on solutions and new ideas.

## Contributing

We encourage your contributions to improve nri-mta. Keep in mind when you submit your pull request, you'll need to sign the CLA via the click-through using CLA-Assistant. You only have to sign the CLA one time per project. If you have any 
questions, or to execute our corporate CLA, required if your contribution is on behalf of a company, please drop us an email at opensource@newrelic.com.


**A note about vulnerabilities**

As noted in our [security policy](../../security/policy), New Relic is committed to the privacy and security of our customers and their data. We believe that providing coordinated disclosure by security researchers and engaging with the security community are important means to achieve our security goals.

If you believe you have found a security vulnerability in this project or any of New Relic's products or websites, we welcome and greatly appreciate you reporting it to New Relic through [HackerOne](https://hackerone.com/newrelic).

If you would like to contribute to this project, review [these guidelines](./CONTRIBUTING.md).

To all contributors, we thank you!  Without your contribution, this project would not be what it is today.

## License

nri-mta is licensed under the [Apache 2.0](/LICENSE) License.